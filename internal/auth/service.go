package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/nguyensongtai/lingora-api/internal/user"
)

// refreshTokenBytes là độ dài entropy của refresh token trước khi mã hoá base64.
const refreshTokenBytes = 32

// users là những gì Service cần ở kho danh tính.
type users interface {
	GetByEmail(ctx context.Context, email string) (user.User, error)
	GetByID(ctx context.Context, id string) (user.User, error)
}

// sessions là những gì Service cần ở kho phiên đăng nhập.
type sessions interface {
	Create(ctx context.Context, userID string, tokenHash []byte, expiresAt time.Time) (Session, error)
	GetActive(ctx context.Context, tokenHash []byte) (Session, error)
	Revoke(ctx context.Context, tokenHash []byte) error
}

// TokenPair là kết quả của một lần đăng nhập hoặc làm mới phiên.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	// ExpiresIn là số giây còn hiệu lực của access token.
	ExpiresIn int64
	User      user.User
}

// Service xử lý đăng nhập, làm mới và đăng xuất.
type Service struct {
	users      users
	sessions   sessions
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration

	// now tách ra để test kiểm soát được thời gian.
	now func() time.Time
}

// Config gom tham số vòng đời token.
type Config struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// NewService từ chối secret yếu ngay lúc boot.
func NewService(userRepo users, sessionRepo sessions, cfg Config) (*Service, error) {
	if len(cfg.Secret) < minSecretLen {
		return nil, fmt.Errorf("%w (got %d)", ErrWeakSecret, len(cfg.Secret))
	}
	return &Service{
		users:      userRepo,
		sessions:   sessionRepo,
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		now:        time.Now,
	}, nil
}

// Login đổi email và mật khẩu lấy một cặp token.
func (s *Service) Login(ctx context.Context, email, password string) (TokenPair, error) {
	found, err := s.users.GetByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			// Vẫn chạy bcrypt để thời gian phản hồi không tố cáo email có thật.
			burnPasswordTime(password)
			return TokenPair{}, ErrInvalidCredentials
		}
		return TokenPair{}, fmt.Errorf("login: %w", err)
	}

	if err := VerifyPassword(found.PasswordHash, password); err != nil {
		return TokenPair{}, err
	}

	return s.issue(ctx, found)
}

// Refresh xoay vòng phiên: token cũ bị thu hồi ngay khi token mới được cấp, nên
// một refresh token chỉ dùng được đúng một lần.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	hash := hashToken(refreshToken)

	session, err := s.sessions.GetActive(ctx, hash)
	if err != nil {
		return TokenPair{}, fmt.Errorf("refresh: %w", err)
	}

	found, err := s.users.GetByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			// Tài khoản đã bị xoá: thu hồi phiên rồi coi như phiên không tồn tại.
			_ = s.sessions.Revoke(ctx, hash)
			return TokenPair{}, ErrSessionNotFound
		}
		return TokenPair{}, fmt.Errorf("refresh: %w", err)
	}

	if err := s.sessions.Revoke(ctx, hash); err != nil {
		return TokenPair{}, fmt.Errorf("refresh: %w", err)
	}

	return s.issue(ctx, found)
}

// Logout thu hồi một phiên. Token không còn hiệu lực cũng coi là thành công.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if err := s.sessions.Revoke(ctx, hashToken(refreshToken)); err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	return nil
}

// CurrentUser đọc người dùng của access token đang dùng.
func (s *Service) CurrentUser(ctx context.Context, userID string) (user.User, error) {
	found, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return user.User{}, fmt.Errorf("current user: %w", err)
	}
	return found, nil
}

func (s *Service) issue(ctx context.Context, account user.User) (TokenPair, error) {
	issuedAt := s.now()

	accessToken, err := s.signAccessToken(account, issuedAt)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, err := newRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}

	expiresAt := issuedAt.Add(s.refreshTTL)
	if _, err := s.sessions.Create(ctx, account.ID, hashToken(refreshToken), expiresAt); err != nil {
		return TokenPair{}, fmt.Errorf("issue tokens: %w", err)
	}

	account.PasswordHash = ""
	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
		User:         account,
	}, nil
}

func (s *Service) signAccessToken(account user.User, issuedAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims{
		Role: string(account.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   account.ID,
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(s.accessTTL)),
		},
	})

	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// newRefreshToken sinh token mờ; nó không mang thông tin nên chỉ database mới
// nói được nó thuộc về ai.
func newRefreshToken() (string, error) {
	raw := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
