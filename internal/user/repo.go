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
		GoogleSub:    params.GoogleSub,
	})
	if err != nil {
		if isUniqueViolation(err, "users_email_key") {
			return User{}, fmt.Errorf("create user %q: %w", params.Email, ErrEmailTaken)
		}
		if isUniqueViolation(err, "users_google_sub_key") {
			return User{}, fmt.Errorf("create user %q: %w", params.Email, ErrGoogleAccountTaken)
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return toUser(row), nil
}

// GetByGoogleSub tìm người dùng theo id Google đã gắn.
func (r *Repo) GetByGoogleSub(ctx context.Context, googleSub string) (User, error) {
	row, err := r.q.GetUserByGoogleSub(ctx, &googleSub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("get user by google sub: %w", ErrNotFound)
		}
		return User{}, fmt.Errorf("get user by google sub: %w", err)
	}
	return toUser(row), nil
}

// LinkGoogle gắn tài khoản Google vào một người dùng đã có. Không có hàng nào
// đổi nghĩa là người đó đã gắn Google khác rồi.
func (r *Repo) LinkGoogle(ctx context.Context, id, googleSub string) (User, error) {
	userID, err := postgres.ParseUUID(id)
	if err != nil {
		return User{}, fmt.Errorf("%w: %q", ErrInvalidID, id)
	}

	row, err := r.q.LinkGoogleSub(ctx, db.LinkGoogleSubParams{ID: userID, GoogleSub: &googleSub})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("link google to %s: %w", id, ErrGoogleAlreadyLinked)
		}
		if isUniqueViolation(err, "users_google_sub_key") {
			return User{}, fmt.Errorf("link google to %s: %w", id, ErrGoogleAccountTaken)
		}
		return User{}, fmt.Errorf("link google to %s: %w", id, err)
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

// UpdateProfile đổi tên hiển thị và/hoặc mục tiêu XP.
func (r *Repo) UpdateProfile(ctx context.Context, id string, params ProfileParams) (User, error) {
	userID, err := postgres.ParseUUID(id)
	if err != nil {
		return User{}, fmt.Errorf("%w: %q", ErrInvalidID, id)
	}

	var goal *int32
	if params.DailyGoalXP != nil {
		value := int32(*params.DailyGoalXP)
		goal = &value
	}
	row, err := r.q.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		ID:          userID,
		DisplayName: params.DisplayName,
		DailyGoalXp: goal,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("update profile %s: %w", id, ErrNotFound)
		}
		return User{}, fmt.Errorf("update profile %s: %w", id, err)
	}
	return toUser(row), nil
}

// UpdatePassword thay hash mật khẩu. Tài khoản không có mật khẩu (chỉ Google)
// không có hàng nào khớp, và trả ErrNotFound.
func (r *Repo) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	userID, err := postgres.ParseUUID(id)
	if err != nil {
		return fmt.Errorf("%w: %q", ErrInvalidID, id)
	}

	affected, err := r.q.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: &passwordHash,
	})
	if err != nil {
		return fmt.Errorf("update password %s: %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("update password %s: %w", id, ErrNotFound)
	}
	return nil
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
		GoogleSub:    row.GoogleSub,
		DailyGoalXP:  int(row.DailyGoalXp),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
