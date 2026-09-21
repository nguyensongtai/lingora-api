package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nguyensongtai/lingora-api/internal/db"
	"github.com/nguyensongtai/lingora-api/internal/platform/postgres"
)

// Session là một refresh token còn hiệu lực trong database.
type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// SessionRepo lưu refresh token dưới dạng băm.
type SessionRepo struct {
	q *db.Queries
}

// NewSessionRepo nhận DBTX nên dùng được cả với pool lẫn transaction.
func NewSessionRepo(dbtx db.DBTX) *SessionRepo {
	return &SessionRepo{q: db.New(dbtx)}
}

// Create lưu một phiên mới; tokenHash là SHA-256 của token gốc.
func (r *SessionRepo) Create(ctx context.Context, userID string, tokenHash []byte, expiresAt time.Time) (Session, error) {
	id, err := postgres.ParseUUID(userID)
	if err != nil {
		return Session{}, fmt.Errorf("%w: %q", ErrInvalidUserID, userID)
	}

	row, err := r.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    id,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return Session{}, fmt.Errorf("create refresh token: %w", err)
	}
	return toSession(row), nil
}

// GetActive trả về phiên còn hiệu lực; token hết hạn hoặc đã thu hồi là
// ErrSessionNotFound.
func (r *SessionRepo) GetActive(ctx context.Context, tokenHash []byte) (Session, error) {
	row, err := r.q.GetActiveRefreshToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, fmt.Errorf("get refresh token: %w", ErrSessionNotFound)
		}
		return Session{}, fmt.Errorf("get refresh token: %w", err)
	}
	return toSession(row), nil
}

// Revoke thu hồi đúng một phiên. Thu hồi phiên không tồn tại không phải lỗi:
// đăng xuất hai lần vẫn là đăng xuất.
func (r *SessionRepo) Revoke(ctx context.Context, tokenHash []byte) error {
	if _, err := r.q.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

// RevokeAllForUser thu hồi mọi phiên của một người dùng.
func (r *SessionRepo) RevokeAllForUser(ctx context.Context, userID string) error {
	id, err := postgres.ParseUUID(userID)
	if err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidUserID, userID)
	}

	if _, err := r.q.RevokeAllUserRefreshTokens(ctx, id); err != nil {
		return fmt.Errorf("revoke all refresh tokens: %w", err)
	}
	return nil
}

// DeleteExpired dọn token đã hết hạn hoặc đã thu hồi, trả về số hàng xoá được.
func (r *SessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	affected, err := r.q.DeleteExpiredRefreshTokens(ctx)
	if err != nil {
		return 0, fmt.Errorf("delete expired refresh tokens: %w", err)
	}
	return affected, nil
}

func toSession(row db.RefreshToken) Session {
	return Session{
		ID:        postgres.UUIDString(row.ID),
		UserID:    postgres.UUIDString(row.UserID),
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	}
}
