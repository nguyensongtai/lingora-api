package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nguyensongtai/lingora-api/internal/db"
	"github.com/nguyensongtai/lingora-api/internal/platform/postgres"
)

const uniqueViolation = "23505"

// Repo đọc ghi người dùng trên PostgreSQL.
type Repo struct {
	q *db.Queries
}

// NewRepo nhận DBTX nên dùng được cả với pool lẫn transaction.
func NewRepo(dbtx db.DBTX) *Repo {
	return &Repo{q: db.New(dbtx)}
}

// Create thêm một người dùng mới.
func (r *Repo) Create(ctx context.Context, params CreateParams) (User, error) {
	row, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Email:        params.Email,
		PasswordHash: params.PasswordHash,
		DisplayName:  params.DisplayName,
		Role:         db.UserRole(params.Role),
	})
	if err != nil {
		if isUniqueViolation(err, "users_email_key") {
			return User{}, fmt.Errorf("create user %q: %w", params.Email, ErrEmailTaken)
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return toUser(row), nil
}

// GetByEmail tìm người dùng theo email, không phân biệt hoa thường.
func (r *Repo) GetByEmail(ctx context.Context, email string) (User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("get user by email: %w", ErrNotFound)
		}
		return User{}, fmt.Errorf("get user by email: %w", err)
	}
	return toUser(row), nil
}

// GetByID tìm người dùng theo id.
func (r *Repo) GetByID(ctx context.Context, id string) (User, error) {
	userID, err := postgres.ParseUUID(id)
	if err != nil {
		return User{}, fmt.Errorf("%w: %q", ErrInvalidID, id)
	}

	row, err := r.q.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("get user %s: %w", id, ErrNotFound)
		}
		return User{}, fmt.Errorf("get user %s: %w", id, err)
	}
	return toUser(row), nil
}

func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == uniqueViolation && pgErr.ConstraintName == constraint
}

func toUser(row db.User) User {
	return User{
		ID:           postgres.UUIDString(row.ID),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		DisplayName:  row.DisplayName,
		Role:         Role(row.Role),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
