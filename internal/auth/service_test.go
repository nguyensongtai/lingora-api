package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nguyensongtai/lingora-api/internal/auth"
	"github.com/nguyensongtai/lingora-api/internal/user"
)

type fakeUsers struct {
	t *testing.T

	create     func(context.Context, user.CreateParams) (user.User, error)
	byEmail    func(context.Context, string) (user.User, error)
	byID       func(context.Context, string) (user.User, error)
	byGoogle   func(context.Context, string) (user.User, error)
	linkGoogle func(context.Context, string, string) (user.User, error)
	profile    func(context.Context, string, user.ProfileParams) (user.User, error)
	// passwordSet ghi lại hash mới nhất được lưu.
	passwordSet string
}

func (f *fakeUsers) UpdateProfile(ctx context.Context, id string, params user.ProfileParams) (user.User, error) {
	if f.profile == nil {
		f.t.Fatal("UpdateProfile called unexpectedly")
	}
	return f.profile(ctx, id, params)
}

func (f *fakeUsers) UpdatePassword(_ context.Context, _ string, passwordHash string) error {
	f.passwordSet = passwordHash
	return nil
}

func (f *fakeUsers) GetByGoogleSub(ctx context.Context, googleSub string) (user.User, error) {
	if f.byGoogle == nil {
		return user.User{}, user.ErrNotFound
	}
	return f.byGoogle(ctx, googleSub)
}

func (f *fakeUsers) LinkGoogle(ctx context.Context, id, googleSub string) (user.User, error) {
	if f.linkGoogle == nil {
		f.t.Fatal("LinkGoogle called unexpectedly")
	}
	return f.linkGoogle(ctx, id, googleSub)
}

func (f *fakeUsers) Create(ctx context.Context, params user.CreateParams) (user.User, error) {
	if f.create == nil {
		f.t.Fatal("Create called unexpectedly")
	}
	return f.create(ctx, params)
}

func (f *fakeUsers) GetByEmail(ctx context.Context, email string) (user.User, error) {
	if f.byEmail == nil {
		f.t.Fatal("GetByEmail called unexpectedly")
	}
	return f.byEmail(ctx, email)
}

func (f *fakeUsers) GetByID(ctx context.Context, id string) (user.User, error) {
	if f.byID == nil {
		f.t.Fatal("GetByID called unexpectedly")
	}
	return f.byID(ctx, id)
}

type fakeSessions struct {
	t *testing.T

	created    []string
	revoked    [][]byte
	revokedAll []string
	active     map[string]auth.Session
}

func (f *fakeSessions) RevokeAllForUser(_ context.Context, userID string) error {
	f.revokedAll = append(f.revokedAll, userID)
	for hash, session := range f.active {
		if session.UserID == userID {
			delete(f.active, hash)
		}
	}
	return nil
}

func newFakeSessions(t *testing.T) *fakeSessions {
	return &fakeSessions{t: t, active: map[string]auth.Session{}}
}

func (f *fakeSessions) Create(_ context.Context, userID string, tokenHash []byte, expiresAt time.Time) (auth.Session, error) {
	session := auth.Session{ID: "session-1", UserID: userID, ExpiresAt: expiresAt}
	f.created = append(f.created, userID)
	f.active[string(tokenHash)] = session
	return session, nil
}

func (f *fakeSessions) GetActive(_ context.Context, tokenHash []byte) (auth.Session, error) {
	session, ok := f.active[string(tokenHash)]
	if !ok {
		return auth.Session{}, auth.ErrSessionNotFound
	}
	return session, nil
}

func (f *fakeSessions) Revoke(_ context.Context, tokenHash []byte) error {
	f.revoked = append(f.revoked, tokenHash)
	delete(f.active, string(tokenHash))
	return nil
}

const servicePassword = "mat-khau-du-dai-12"

func newService(t *testing.T, users *fakeUsers, sessions *fakeSessions) *auth.Service {
	t.Helper()

	service, err := auth.NewService(users, sessions, auth.Config{
		Secret:     testSecret,
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 30 * 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewService() returned error: %v", err)
	}
	return service
}

func adminAccount(t *testing.T) user.User {
	t.Helper()

	hash, err := auth.HashPassword(servicePassword)
	if err != nil {
		t.Fatalf("HashPassword() returned error: %v", err)
	}
	return user.User{
		ID:           "018f3a9c-7b2e-7c31-9a55-0f1d2e3a4b5c",
		Email:        "admin@lingora.vn",
		PasswordHash: &hash,
		DisplayName:  "Quản trị",
		Role:         user.RoleAdmin,
	}
}

func TestNewServiceRejectsWeakSecret(t *testing.T) {
	t.Parallel()

	_, err := auth.NewService(&fakeUsers{t: t}, newFakeSessions(t), auth.Config{Secret: "qua-ngan"})
	if !errors.Is(err, auth.ErrWeakSecret) {
		t.Fatalf("error = %v, want it to match ErrWeakSecret", err)
	}
}

func TestLoginIssuesUsableTokenPair(t *testing.T) {
	t.Parallel()

	account := adminAccount(t)
	sessions := newFakeSessions(t)
	service := newService(t, &fakeUsers{t: t, byEmail: func(context.Context, string) (user.User, error) {
		return account, nil
	}}, sessions)

	pair, err := service.Login(context.Background(), "  admin@lingora.vn  ", servicePassword)
	if err != nil {
		t.Fatalf("Login() returned error: %v", err)
	}

	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("Login() returned an empty token")
	}
	if pair.ExpiresIn != int64((15 * time.Minute).Seconds()) {
		t.Errorf("ExpiresIn = %d, want 900", pair.ExpiresIn)
	}
	if pair.User.PasswordHash != nil {
		t.Error("TokenPair carries the password hash, want it stripped")
	}
	if len(sessions.created) != 1 || sessions.created[0] != account.ID {
		t.Errorf("sessions created = %v, want one for %s", sessions.created, account.ID)
	}

	// Access token phải qua được chính middleware của hệ thống.
	verifier, err := auth.NewVerifier(testSecret)
	if err != nil {
		t.Fatalf("NewVerifier() returned error: %v", err)
	}
	reached := false
	handler := verifier.RequireRole(auth.RoleAdmin)(okHandler(&reached))
	if recorder := call(handler, "Bearer "+pair.AccessToken); recorder.Code != 200 {
		t.Fatalf("access token rejected by the middleware: %d %s", recorder.Code, recorder.Body.String())
	}
	if !reached {
		t.Error("middleware did not pass the request through")
	}
}

func TestLoginRejectsBadCredentials(t *testing.T) {
	t.Parallel()

	account := adminAccount(t)

	tests := map[string]struct {
		users    *fakeUsers
		password string
	}{
		"email không tồn tại": {
			users:    &fakeUsers{byEmail: func(context.Context, string) (user.User, error) { return user.User{}, user.ErrNotFound }},
			password: servicePassword,
		},
		"sai mật khẩu": {
			users:    &fakeUsers{byEmail: func(context.Context, string) (user.User, error) { return account, nil }},
			password: "mat-khau-sai-hoan-toan",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.users.t = t
			sessions := newFakeSessions(t)

			_, err := newService(t, tc.users, sessions).Login(context.Background(), account.Email, tc.password)
			// Cùng một lỗi cho cả hai nhánh nên client không dò được email nào có thật.
			if !errors.Is(err, auth.ErrInvalidCredentials) {
				t.Fatalf("error = %v, want it to match ErrInvalidCredentials", err)
			}
			if len(sessions.created) != 0 {
				t.Error("a session was created for a failed login")
			}
		})
	}
}

func TestRefreshRotatesTheToken(t *testing.T) {
	t.Parallel()

	account := adminAccount(t)
	sessions := newFakeSessions(t)
	service := newService(t, &fakeUsers{
		t:       t,
		byEmail: func(context.Context, string) (user.User, error) { return account, nil },
		byID:    func(context.Context, string) (user.User, error) { return account, nil },
	}, sessions)

	first, err := service.Login(context.Background(), account.Email, servicePassword)
	if err != nil {
		t.Fatalf("Login() returned error: %v", err)
	}

	second, err := service.Refresh(context.Background(), first.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() returned error: %v", err)
	}
	if second.RefreshToken == first.RefreshToken {
		t.Error("Refresh() reused the same refresh token, want a rotated one")
	}

	// Token cũ phải chết ngay: dùng lại là dấu hiệu token bị đánh cắp.
	if _, err := service.Refresh(context.Background(), first.RefreshToken); !errors.Is(err, auth.ErrSessionNotFound) {
		t.Fatalf("reusing the old refresh token gave %v, want ErrSessionNotFound", err)
	}
}

func TestRefreshRejectsUnknownToken(t *testing.T) {
	t.Parallel()

	service := newService(t, &fakeUsers{t: t}, newFakeSessions(t))

	if _, err := service.Refresh(context.Background(), "khong-phai-token-that"); !errors.Is(err, auth.ErrSessionNotFound) {
		t.Fatalf("error = %v, want it to match ErrSessionNotFound", err)
	}
}

func TestRefreshRevokesSessionOfDeletedUser(t *testing.T) {
	t.Parallel()

	account := adminAccount(t)
	sessions := newFakeSessions(t)
	service := newService(t, &fakeUsers{
		t:       t,
		byEmail: func(context.Context, string) (user.User, error) { return account, nil },
		byID:    func(context.Context, string) (user.User, error) { return user.User{}, user.ErrNotFound },
	}, sessions)

	pair, err := service.Login(context.Background(), account.Email, servicePassword)
	if err != nil {
		t.Fatalf("Login() returned error: %v", err)
	}

	if _, err := service.Refresh(context.Background(), pair.RefreshToken); !errors.Is(err, auth.ErrSessionNotFound) {
		t.Fatalf("error = %v, want it to match ErrSessionNotFound", err)
	}
	if len(sessions.revoked) == 0 {
		t.Error("session of a deleted user was left active")
	}
}

func TestLogoutRevokesTheSession(t *testing.T) {
	t.Parallel()

	account := adminAccount(t)
	sessions := newFakeSessions(t)
	service := newService(t, &fakeUsers{
		t:       t,
		byEmail: func(context.Context, string) (user.User, error) { return account, nil },
		byID:    func(context.Context, string) (user.User, error) { return account, nil },
	}, sessions)

	pair, err := service.Login(context.Background(), account.Email, servicePassword)
	if err != nil {
		t.Fatalf("Login() returned error: %v", err)
	}

	if err := service.Logout(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("Logout() returned error: %v", err)
	}
	if _, err := service.Refresh(context.Background(), pair.RefreshToken); !errors.Is(err, auth.ErrSessionNotFound) {
		t.Fatalf("refresh after logout gave %v, want ErrSessionNotFound", err)
	}

	// Đăng xuất hai lần vẫn là đăng xuất.
	if err := service.Logout(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("second Logout() returned error: %v", err)
	}
}

func TestPasswordHashIsSaltedPerCall(t *testing.T) {
	t.Parallel()

	first, err := auth.HashPassword(servicePassword)
	if err != nil {
		t.Fatalf("HashPassword() returned error: %v", err)
	}
	second, err := auth.HashPassword(servicePassword)
	if err != nil {
		t.Fatalf("HashPassword() returned error: %v", err)
	}

	if first == second {
		t.Error("cùng một mật khẩu cho ra hai hash giống nhau, tức là thiếu salt")
	}
	if err := auth.VerifyPassword(first, servicePassword); err != nil {
		t.Errorf("VerifyPassword() on a fresh hash returned %v", err)
	}
	if err := auth.VerifyPassword(first, "mat-khau-khac-han"); !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("VerifyPassword() with a wrong password gave %v, want ErrInvalidCredentials", err)
	}
}

func TestRegisterCreatesAStudentAndLogsThemIn(t *testing.T) {
	t.Parallel()

	var got user.CreateParams
	users := &fakeUsers{
		t: t,
		create: func(_ context.Context, params user.CreateParams) (user.User, error) {
			got = params
			return user.User{ID: "user-1", Email: params.Email, DisplayName: params.DisplayName, Role: params.Role}, nil
		},
	}

	service := newService(t, users, newFakeSessions(t))
	pair, err := service.Register(context.Background(), "  An@Example.com  ", "mat-khau-du-dai-2026", "  Nguyễn An  ")
	if err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}

	// Vai trò do server quyết định, không nhận từ body — không có đường nào tự
	// nhận quyền admin khi đăng ký.
	if got.Role != user.RoleStudent {
		t.Errorf("role = %q, want %q", got.Role, user.RoleStudent)
	}
	if got.Email != "An@Example.com" {
		t.Errorf("email = %q, want nó được trim", got.Email)
	}
	if got.DisplayName != "Nguyễn An" {
		t.Errorf("display_name = %q, want nó được trim", got.DisplayName)
	}
	if got.PasswordHash == nil || *got.PasswordHash == "mat-khau-du-dai-2026" {
		t.Error("password_hash phải là một hash bcrypt, không phải mật khẩu thô")
	}
	if got.GoogleSub != nil {
		t.Error("đăng ký bằng email không được gắn kèm tài khoản Google")
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Error("Register() phải trả luôn cặp token, không bắt đăng nhập lại")
	}
}

func TestRegisterRejectsBadInput(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		email, password, displayName string
		field                        string
	}{
		"email sai định dạng": {"khong-phai-email", "mat-khau-du-dai-2026", "An", "email"},
		"mật khẩu quá ngắn":   {"an@example.com", "ngan", "An", "password"},
		"tên rỗng":            {"an@example.com", "mat-khau-du-dai-2026", "   ", "display_name"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// create không được gọi: sai dữ liệu thì không chạm tới database.
			service := newService(t, &fakeUsers{t: t}, newFakeSessions(t))

			_, err := service.Register(context.Background(), tc.email, tc.password, tc.displayName)

			var validationErr *auth.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("Register() error = %v, want *ValidationError", err)
			}
			if _, ok := validationErr.Fields[tc.field]; !ok {
				t.Errorf("Fields = %v, want có key %q", validationErr.Fields, tc.field)
			}
		})
	}
}

func TestRegisterSurfacesDuplicateEmail(t *testing.T) {
	t.Parallel()

	users := &fakeUsers{
		t: t,
		create: func(context.Context, user.CreateParams) (user.User, error) {
			return user.User{}, user.ErrEmailTaken
		},
	}

	_, err := newService(t, users, newFakeSessions(t)).
		Register(context.Background(), "an@example.com", "mat-khau-du-dai-2026", "An")

	if !errors.Is(err, user.ErrEmailTaken) {
		t.Fatalf("Register() error = %v, want ErrEmailTaken", err)
	}
}

func TestLoginRejectsAnAccountWithoutAPassword(t *testing.T) {
	t.Parallel()

	// Tài khoản chỉ đăng nhập bằng Google: không có hash nào để so.
	googleOnly := user.User{
		ID:          "018f3a9c-7b2e-7c31-9a55-0f1d2e3a4b60",
		Email:       "an@example.com",
		DisplayName: "An",
		Role:        user.RoleStudent,
		GoogleSub:   ptr("google-123"),
	}
	users := &fakeUsers{
		t:       t,
		byEmail: func(context.Context, string) (user.User, error) { return googleOnly, nil },
	}

	_, err := newService(t, users, newFakeSessions(t)).
		Login(context.Background(), "an@example.com", "mat-khau-nao-do-2026")

	// Lỗi phải giống hệt trường hợp sai mật khẩu: câu trả lời không được tiết
	// lộ tài khoản đó đăng nhập bằng cách nào.
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func ptr[T any](value T) *T { return &value }

func TestLoginRateLimitsPerEmail(t *testing.T) {
	t.Parallel()

	account := adminAccount(t)
	users := &fakeUsers{
		t:       t,
		byEmail: func(context.Context, string) (user.User, error) { return account, nil },
	}
	service := newService(t, users, newFakeSessions(t))

	for attempt := 1; attempt <= auth.LoginEmailRule.Limit; attempt++ {
		_, err := service.Login(context.Background(), account.Email, "sai-mat-khau")
		if !errors.Is(err, auth.ErrInvalidCredentials) {
			t.Fatalf("lần %d: error = %v, want ErrInvalidCredentials", attempt, err)
		}
	}

	_, err := service.Login(context.Background(), account.Email, "sai-mat-khau")

	var limited *auth.RateLimitedError
	if !errors.As(err, &limited) {
		t.Fatalf("error = %v, want *RateLimitedError", err)
	}
	if limited.RetryAfter <= 0 {
		t.Errorf("RetryAfter = %v, want lớn hơn 0", limited.RetryAfter)
	}

	// Đổi hoa thường không lách được hạn mức: khoá tính trên email đã chuẩn hoá.
	if _, err := service.Login(context.Background(), "ADMIN@LINGORA.VN", "sai-mat-khau"); !errors.Is(err, auth.ErrRateLimited) {
		t.Errorf("email viết hoa: error = %v, want vẫn bị chặn", err)
	}
}

func TestLoginResetsTheLimitOnSuccess(t *testing.T) {
	t.Parallel()

	account := adminAccount(t)
	users := &fakeUsers{
		t:       t,
		byEmail: func(context.Context, string) (user.User, error) { return account, nil },
	}
	service := newService(t, users, newFakeSessions(t))

	// Gõ sai vài lần rồi gõ đúng: bộ đếm phải về 0, nếu không một người hay
	// quên mật khẩu sẽ tự khoá mình ở lần đăng nhập sau.
	for range auth.LoginEmailRule.Limit - 1 {
		if _, err := service.Login(context.Background(), account.Email, "sai-mat-khau"); err == nil {
			t.Fatal("Login() với mật khẩu sai lại thành công")
		}
	}
	if _, err := service.Login(context.Background(), account.Email, servicePassword); err != nil {
		t.Fatalf("Login() đúng mật khẩu returned error: %v", err)
	}

	for attempt := range auth.LoginEmailRule.Limit {
		_, err := service.Login(context.Background(), account.Email, "sai-mat-khau")
		if errors.Is(err, auth.ErrRateLimited) {
			t.Fatalf("bị chặn lại ở lần %d sau khi đăng nhập thành công", attempt+1)
		}
	}
}
