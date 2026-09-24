package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nguyensongtai/lingora-api/internal/testdb"
	"github.com/nguyensongtai/lingora-api/internal/user"
)

func newRepo(t *testing.T) (*user.Repo, pgx.Tx) {
	t.Helper()

	tx := testdb.New(t)
	return user.NewRepo(tx), tx
}

func attempt(t *testing.T, tx pgx.Tx, fn func(*user.Repo) error) error {
	t.Helper()

	return testdb.Attempt(t, tx, func(nested pgx.Tx) error {
		return fn(user.NewRepo(nested))
	})
}

func hash(value string) *string { return &value }

func TestRepoEmailIsUniqueIgnoringCase(t *testing.T) {
	t.Parallel()

	repo, tx := newRepo(t)
	ctx := context.Background()

	if _, err := repo.Create(ctx, user.CreateParams{
		Email: "An@Lingora.vn", PasswordHash: hash("$2a$12$x"), DisplayName: "An", Role: user.RoleStudent,
	}); err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	// Unique index đặt trên lower(email): viết hoa khác đi cũng là cùng một người.
	err := attempt(t, tx, func(r *user.Repo) error {
		_, createErr := r.Create(ctx, user.CreateParams{
			Email: "an@lingora.VN", PasswordHash: hash("$2a$12$y"), DisplayName: "Ai đó", Role: user.RoleStudent,
		})
		return createErr
	})
	if !errors.Is(err, user.ErrEmailTaken) {
		t.Fatalf("error = %v, want ErrEmailTaken", err)
	}

	// Và tìm cũng không phân biệt hoa thường.
	found, err := repo.GetByEmail(ctx, "AN@LINGORA.VN")
	if err != nil {
		t.Fatalf("GetByEmail() returned error: %v", err)
	}
	if found.DisplayName != "An" {
		t.Errorf("DisplayName = %q, want %q", found.DisplayName, "An")
	}
}

func TestRepoLinkGoogleOnlyOnce(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)
	ctx := context.Background()

	account, err := repo.Create(ctx, user.CreateParams{
		Email: "gan-google@lingora.vn", PasswordHash: hash("$2a$12$x"), DisplayName: "An", Role: user.RoleStudent,
	})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	linked, err := repo.LinkGoogle(ctx, account.ID, "google-sub-test-1")
	if err != nil {
		t.Fatalf("LinkGoogle() returned error: %v", err)
	}
	if linked.GoogleSub == nil || *linked.GoogleSub != "google-sub-test-1" {
		t.Fatalf("GoogleSub = %v, want google-sub-test-1", linked.GoogleSub)
	}

	// Câu UPDATE có WHERE google_sub IS NULL, nên gắn lần hai không ghi đè.
	if _, err := repo.LinkGoogle(ctx, account.ID, "google-sub-test-2"); !errors.Is(err, user.ErrGoogleAlreadyLinked) {
		t.Errorf("gắn lần hai: error = %v, want ErrGoogleAlreadyLinked", err)
	}

	if _, err := repo.GetByGoogleSub(ctx, "google-sub-test-1"); err != nil {
		t.Errorf("GetByGoogleSub() returned error: %v", err)
	}
	if _, err := repo.GetByGoogleSub(ctx, "google-sub-khong-co"); !errors.Is(err, user.ErrNotFound) {
		t.Errorf("sub không tồn tại: error = %v, want ErrNotFound", err)
	}
}

func TestRepoGoogleSubIsUniqueAcrossUsers(t *testing.T) {
	t.Parallel()

	repo, tx := newRepo(t)
	ctx := context.Background()

	first, err := repo.Create(ctx, user.CreateParams{
		Email: "google-mot@lingora.vn", DisplayName: "Một", Role: user.RoleStudent,
		GoogleSub: hash("google-sub-dung-chung"),
	})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	if first.PasswordHash != nil {
		t.Error("tài khoản tạo từ Google lại có mật khẩu")
	}

	// Một tài khoản Google không được gắn vào hai người.
	err = attempt(t, tx, func(r *user.Repo) error {
		_, createErr := r.Create(ctx, user.CreateParams{
			Email: "google-hai@lingora.vn", DisplayName: "Hai", Role: user.RoleStudent,
			GoogleSub: hash("google-sub-dung-chung"),
		})
		return createErr
	})
	if !errors.Is(err, user.ErrGoogleAccountTaken) {
		t.Errorf("error = %v, want ErrGoogleAccountTaken", err)
	}
}

func TestRepoRejectsAnAccountWithNoCredential(t *testing.T) {
	t.Parallel()

	repo, tx := newRepo(t)
	ctx := context.Background()

	// CHECK users_has_credential: không mật khẩu, không Google thì không có
	// cách nào đăng nhập, nên hàng đó không được phép tồn tại.
	err := attempt(t, tx, func(r *user.Repo) error {
		_, createErr := r.Create(ctx, user.CreateParams{
			Email: "khong-co-gi@lingora.vn", DisplayName: "X", Role: user.RoleStudent,
		})
		return createErr
	})
	if err == nil {
		t.Fatal("Create() thành công, want database từ chối")
	}
	_ = repo
}

func TestRepoUpdateProfileChangesOnlyWhatIsGiven(t *testing.T) {
	t.Parallel()

	repo, tx := newRepo(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, user.CreateParams{
		Email: "ho-so@lingora.vn", PasswordHash: hash("$2a$12$x"), DisplayName: "Cũ", Role: user.RoleStudent,
	})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	if created.DailyGoalXP != 50 {
		t.Errorf("DailyGoalXP mặc định = %d, want 50", created.DailyGoalXP)
	}

	goal := 100
	updated, err := repo.UpdateProfile(ctx, created.ID, user.ProfileParams{DailyGoalXP: &goal})
	if err != nil {
		t.Fatalf("UpdateProfile() returned error: %v", err)
	}
	if updated.DailyGoalXP != 100 || updated.DisplayName != "Cũ" {
		t.Errorf("sau khi sửa = %+v, want mục tiêu 100 và tên giữ nguyên", updated)
	}

	// CHECK trong database là lưới cuối nếu service để lọt một mức lạ.
	odd := 35
	err = attempt(t, tx, func(r *user.Repo) error {
		_, updateErr := r.UpdateProfile(ctx, created.ID, user.ProfileParams{DailyGoalXP: &odd})
		return updateErr
	})
	if err == nil {
		t.Error("UpdateProfile(35 XP) returned nil error, want vi phạm users_daily_goal_xp_allowed")
	}
}

func TestRepoUpdatePasswordSkipsGoogleOnlyAccounts(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)
	ctx := context.Background()

	withPassword, err := repo.Create(ctx, user.CreateParams{
		Email: "co-mat-khau@lingora.vn", PasswordHash: hash("$2a$12$cu"), DisplayName: "A", Role: user.RoleStudent,
	})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	googleOnly, err := repo.Create(ctx, user.CreateParams{
		Email: "chi-google@lingora.vn", GoogleSub: hash("sub-123"), DisplayName: "B", Role: user.RoleStudent,
	})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	if err := repo.UpdatePassword(ctx, withPassword.ID, "$2a$12$moi"); err != nil {
		t.Fatalf("UpdatePassword() returned error: %v", err)
	}
	reread, err := repo.GetByID(ctx, withPassword.ID)
	if err != nil || reread.PasswordHash == nil || *reread.PasswordHash != "$2a$12$moi" {
		t.Errorf("hash sau khi đổi = %v (err %v), want $2a$12$moi", reread.PasswordHash, err)
	}

	if err := repo.UpdatePassword(ctx, googleOnly.ID, "$2a$12$moi"); !errors.Is(err, user.ErrNotFound) {
		t.Errorf("UpdatePassword(chỉ Google) error = %v, want ErrNotFound", err)
	}
}
