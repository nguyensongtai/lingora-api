package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"

	"github.com/nguyensongtai/lingora-api/internal/ratelimit"
	"github.com/nguyensongtai/lingora-api/internal/user"
)

// LoginEmailRule giới hạn số lần thử trên MỘT tài khoản, bất kể đến từ đâu.
// Hạn mức theo IP nằm ở middleware; hai lớp chặn hai kiểu tấn công khác nhau:
// một kẻ dò nhiều mật khẩu của một người, và một kẻ dò nhiều người từ một máy.
var LoginEmailRule = ratelimit.Rule{
	Name:   "login-email",
	Limit:  5,
	Window: 15 * time.Minute,
}

// refreshTokenBytes là độ dài entropy của refresh token trước khi mã hoá base64.
const refreshTokenBytes = 32

// users là những gì Service cần ở kho danh tính.
type users interface {
	Create(ctx context.Context, params user.CreateParams) (user.User, error)
	GetByEmail(ctx context.Context, email string) (user.User, error)
	GetByID(ctx context.Context, id string) (user.User, error)
	GetByGoogleSub(ctx context.Context, googleSub string) (user.User, error)
	LinkGoogle(ctx context.Context, id, googleSub string) (user.User, error)
}

// googleExchanger đổi authorization code lấy danh tính. Khai báo ở phía
// consumer nên GoogleClient không phải biết đến nó; nil nghĩa là tính năng tắt.
type googleExchanger interface {
	Exchange(ctx context.Context, code, redirectURI string) (GoogleIdentity, error)
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
	google     googleExchanger
	limiter    ratelimit.Limiter
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
	// Google để nil thì đăng nhập Google tắt, và endpoint trả 503.
	Google googleExchanger
	// Limiter để nil thì dùng bộ đếm trong bộ nhớ. Không có đường nào tắt hẳn
	// hạn mức đăng nhập: tắt im lặng là cách hỏng tệ nhất của một lớp bảo vệ.
	Limiter ratelimit.Limiter
}

// NewService từ chối secret yếu ngay lúc boot.
func NewService(userRepo users, sessionRepo sessions, cfg Config) (*Service, error) {
	if len(cfg.Secret) < minSecretLen {
		return nil, fmt.Errorf("%w (got %d)", ErrWeakSecret, len(cfg.Secret))
	}
	limiter := cfg.Limiter
	if limiter == nil {
		limiter = ratelimit.NewMemory()
	}

	return &Service{
		users:      userRepo,
		sessions:   sessionRepo,
		google:     cfg.Google,
		limiter:    limiter,
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		now:        time.Now,
	}, nil
}

// maxDisplayNameLen khớp với chỗ hiển thị hẹp nhất trong giao diện.
const maxDisplayNameLen = 100

// emailPattern chỉ chặn những chuỗi rõ ràng không phải email. Xác minh thật sự
// là việc của email xác thực, không phải của regex.
var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// Register tạo tài khoản học viên rồi đăng nhập luôn — người mới đăng ký không
// có lý do gì phải nhập lại mật khẩu vừa đặt.
//
// Vai trò luôn là student: không có đường nào tự nhận quyền admin từ bên ngoài.
func (s *Service) Register(ctx context.Context, email, password, displayName string) (TokenPair, error) {
	email = strings.TrimSpace(email)
	displayName = strings.TrimSpace(displayName)

	var v validationBuilder
	if !emailPattern.MatchString(email) {
		v.add("email", "chưa đúng định dạng email")
	}
	if len([]rune(password)) < MinPasswordLen {
		v.add("password", fmt.Sprintf("cần ít nhất %d ký tự", MinPasswordLen))
	}
	if displayName == "" {
		v.add("display_name", "không được để trống")
	} else if utf8.RuneCountInString(displayName) > maxDisplayNameLen {
		v.add("display_name", fmt.Sprintf("tối đa %d ký tự", maxDisplayNameLen))
	}
	if err := v.err(); err != nil {
		return TokenPair{}, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return TokenPair{}, fmt.Errorf("register: %w", err)
	}

	// Email trùng được phát hiện bằng unique index chứ không phải một câu SELECT
	// đứng trước: hai lần đăng ký cùng lúc thì đúng một lần thắng.
	created, err := s.users.Create(ctx, user.CreateParams{
		Email:        email,
		PasswordHash: &hash,
		DisplayName:  displayName,
		Role:         user.RoleStudent,
	})
	if err != nil {
		return TokenPair{}, fmt.Errorf("register: %w", err)
	}

	return s.issue(ctx, created)
}

// Login đổi email và mật khẩu lấy một cặp token.
func (s *Service) Login(ctx context.Context, email, password string) (TokenPair, error) {
	// Khoá theo email đã chuẩn hoá để đổi hoa thường không lách được hạn mức.
	limitKey := LoginEmailRule.Key(strings.ToLower(strings.TrimSpace(email)))
	decision, err := s.limiter.Allow(ctx, limitKey, LoginEmailRule.Limit, LoginEmailRule.Window)
	if err != nil {
		return TokenPair{}, fmt.Errorf("login: %w", err)
	}
	if !decision.Allowed {
		return TokenPair{}, &RateLimitedError{RetryAfter: decision.RetryAfter}
	}

	found, err := s.users.GetByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			// Vẫn chạy bcrypt để thời gian phản hồi không tố cáo email có thật.
			burnPasswordTime(password)
			return TokenPair{}, ErrInvalidCredentials
		}
		return TokenPair{}, fmt.Errorf("login: %w", err)
	}

	// Tài khoản chỉ có Google thì không có mật khẩu để so. Vẫn chạy bcrypt giả
	// rồi trả đúng lỗi như sai mật khẩu: câu trả lời không được tiết lộ tài
	// khoản đó đăng nhập bằng cách nào.
	if !found.HasPassword() {
		burnPasswordTime(password)
		return TokenPair{}, ErrInvalidCredentials
	}

	if err := VerifyPassword(*found.PasswordHash, password); err != nil {
		return TokenPair{}, err
	}

	// Vào được thì xoá bộ đếm, nếu không một người gõ sai vài lần rồi gõ đúng
	// vẫn bị khoá ở lần đăng nhập sau.
	if err := s.limiter.Reset(ctx, limitKey); err != nil {
		slog.WarnContext(ctx, "reset login rate limit", slog.Any("error", err))
	}

	return s.issue(ctx, found)
}

// GoogleEnabled cho biết endpoint đăng nhập Google có dùng được hay không.
func (s *Service) GoogleEnabled() bool { return s.google != nil }

// SignInWithGoogle đổi authorization code lấy phiên đăng nhập.
//
// Ba nhánh, theo thứ tự: đã gắn Google thì vào thẳng; có tài khoản cùng email
// thì gắn thêm Google vào đó; chưa có gì thì tạo tài khoản mới không mật khẩu.
//
// Nhánh giữa chỉ chạy khi Google đã xác minh email — nếu không, bất kỳ ai tạo
// được một tài khoản Google mang email của người khác sẽ chiếm được tài khoản
// Lingora của họ.
func (s *Service) SignInWithGoogle(ctx context.Context, code, redirectURI string) (TokenPair, error) {
	if s.google == nil {
		return TokenPair{}, ErrGoogleNotConfigured
	}

	var v validationBuilder
	if strings.TrimSpace(code) == "" {
		v.add("code", "không được rỗng")
	}
	if strings.TrimSpace(redirectURI) == "" {
		v.add("redirect_uri", "không được rỗng")
	}
	if err := v.err(); err != nil {
		return TokenPair{}, err
	}

	identity, err := s.google.Exchange(ctx, code, redirectURI)
	if err != nil {
		return TokenPair{}, fmt.Errorf("google sign-in: %w", err)
	}
	if !identity.EmailVerified {
		return TokenPair{}, ErrGoogleEmailUnverified
	}

	linked, err := s.users.GetByGoogleSub(ctx, identity.Subject)
	switch {
	case err == nil:
		return s.issue(ctx, linked)
	case !errors.Is(err, user.ErrNotFound):
		return TokenPair{}, fmt.Errorf("google sign-in: %w", err)
	}

	existing, err := s.users.GetByEmail(ctx, identity.Email)
	switch {
	case err == nil:
		attached, linkErr := s.users.LinkGoogle(ctx, existing.ID, identity.Subject)
		if linkErr != nil {
			return TokenPair{}, fmt.Errorf("google sign-in: %w", linkErr)
		}
		return s.issue(ctx, attached)
	case !errors.Is(err, user.ErrNotFound):
		return TokenPair{}, fmt.Errorf("google sign-in: %w", err)
	}

	created, err := s.users.Create(ctx, user.CreateParams{
		Email:       identity.Email,
		DisplayName: displayNameFor(identity),
		Role:        user.RoleStudent,
		GoogleSub:   &identity.Subject,
	})
	if err != nil {
		return TokenPair{}, fmt.Errorf("google sign-in: %w", err)
	}
	return s.issue(ctx, created)
}

// displayNameFor lấy tên Google trả về; tài khoản không đặt tên thì lùi về
// phần trước @ của email, vì giao diện luôn cần một cái tên để hiển thị.
func displayNameFor(identity GoogleIdentity) string {
	if name := strings.TrimSpace(identity.Name); name != "" {
		return name
	}
	local, _, _ := strings.Cut(identity.Email, "@")
	return local
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

	account.PasswordHash = nil
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
