package auth_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nguyensongtai/lingora-api/internal/auth"
	"github.com/nguyensongtai/lingora-api/internal/user"
)

func fieldsOf(t *testing.T, err error) map[string]string {
	t.Helper()

	var validationErr *auth.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %v, want *ValidationError", err)
	}
	return validationErr.Fields
}

/* ---------- hồ sơ ---------- */

func TestUpdateProfileTrimsAndPassesOnlyWhatChanged(t *testing.T) {
	t.Parallel()

	var got user.ProfileParams
	users := &fakeUsers{t: t, profile: func(_ context.Context, _ string, params user.ProfileParams) (user.User, error) {
		got = params
		account := adminAccount(t)
		account.DisplayName = *params.DisplayName
		return account, nil
	}}

	name := "  Lan Anh  "
	updated, err := newService(t, users, newFakeSessions(t)).UpdateProfile(context.Background(), "u-1", &name, nil)
	if err != nil {
		t.Fatalf("UpdateProfile() returned error: %v", err)
	}
	if got.DisplayName == nil || *got.DisplayName != "Lan Anh" || got.DailyGoalXP != nil {
		t.Errorf("gửi xuống kho %+v, want tên đã cắt khoảng trắng và mục tiêu giữ nguyên", got)
	}
	if updated.PasswordHash != nil {
		t.Error("hash mật khẩu lọt ra khỏi service")
	}
	// Xoá hash không được xoá luôn sự thật là tài khoản có mật khẩu.
	if !updated.HasPassword() {
		t.Error("HasPassword() = false với một tài khoản có mật khẩu")
	}
}

func TestUpdateProfileRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	blank, long, goal := "   ", strings.Repeat("a", 101), 30
	cases := map[string]struct {
		name  *string
		goal  *int
		field string
	}{
		"không gửi gì":       {nil, nil, "body"},
		"tên rỗng":           {&blank, nil, "display_name"},
		"tên quá dài":        {&long, nil, "display_name"},
		"mục tiêu ngoài mức": {nil, &goal, "daily_goal_xp"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// profile nil: dữ liệu sai thì không được chạm tới kho.
			_, err := newService(t, &fakeUsers{t: t}, newFakeSessions(t)).
				UpdateProfile(context.Background(), "u-1", tc.name, tc.goal)
			if _, ok := fieldsOf(t, err)[tc.field]; !ok {
				t.Errorf("Fields = %v, want có key %s", fieldsOf(t, err), tc.field)
			}
		})
	}
}

/* ---------- đổi mật khẩu ---------- */

const newPassword = "mat-khau-moi-du-dai"

func accountUsers(t *testing.T, account user.User) *fakeUsers {
	return &fakeUsers{t: t, byID: func(context.Context, string) (user.User, error) { return account, nil }}
}

func TestChangePasswordSignsOutEverySessionAndIssuesANewOne(t *testing.T) {
	t.Parallel()

	account := adminAccount(t)
	users := accountUsers(t, account)
	sessions := newFakeSessions(t)

	pair, err := newService(t, users, sessions).
		ChangePassword(context.Background(), account.ID, servicePassword, newPassword)
	if err != nil {
		t.Fatalf("ChangePassword() returned error: %v", err)
	}

	if err := auth.VerifyPassword(users.passwordSet, newPassword); err != nil {
		t.Errorf("hash đã lưu không khớp mật khẩu mới: %v", err)
	}
	if len(sessions.revokedAll) != 1 || sessions.revokedAll[0] != account.ID {
		t.Errorf("revokedAll = %v, want thu hồi mọi phiên của %s", sessions.revokedAll, account.ID)
	}
	// Thu hồi trước, cấp sau: phiên mới phải là phiên duy nhất còn sống.
	if pair.RefreshToken == "" || len(sessions.active) != 1 {
		t.Errorf("phiên còn sống = %d, want đúng một phiên mới", len(sessions.active))
	}
	if pair.User.PasswordHash != nil || !pair.User.HasPassword() {
		t.Errorf("User trả kèm token: hash = %v, HasPassword = %v — want không hash, vẫn có mật khẩu",
			pair.User.PasswordHash, pair.User.HasPassword())
	}
}

func TestChangePasswordRejectsAWrongCurrentPasswordWithoutTouchingAnything(t *testing.T) {
	t.Parallel()

	account := adminAccount(t)
	users := accountUsers(t, account)
	sessions := newFakeSessions(t)

	_, err := newService(t, users, sessions).
		ChangePassword(context.Background(), account.ID, "khong-phai-mat-khau", newPassword)

	// 400 theo field, không phải ErrInvalidCredentials (401): phiên vẫn hợp lệ.
	if _, ok := fieldsOf(t, err)["current_password"]; !ok {
		t.Errorf("Fields = %v, want current_password", fieldsOf(t, err))
	}
	if users.passwordSet != "" || len(sessions.revokedAll) != 0 {
		t.Error("đã đổi mật khẩu hoặc thu hồi phiên dù mật khẩu hiện tại sai")
	}
}

func TestChangePasswordValidatesTheNewPassword(t *testing.T) {
	t.Parallel()

	for name, next := range map[string]string{
		"quá ngắn":          "ngan",
		"trùng mật khẩu cũ": servicePassword,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			account := adminAccount(t)
			_, err := newService(t, accountUsers(t, account), newFakeSessions(t)).
				ChangePassword(context.Background(), account.ID, servicePassword, next)
			if _, ok := fieldsOf(t, err)["new_password"]; !ok {
				t.Errorf("Fields = %v, want new_password", fieldsOf(t, err))
			}
		})
	}
}

func TestChangePasswordRefusesAGoogleOnlyAccount(t *testing.T) {
	t.Parallel()

	account := adminAccount(t)
	account.PasswordHash = nil

	_, err := newService(t, accountUsers(t, account), newFakeSessions(t)).
		ChangePassword(context.Background(), account.ID, "bat-ky-gi-cung-duoc", newPassword)
	if !errors.Is(err, auth.ErrNoPassword) {
		t.Fatalf("error = %v, want ErrNoPassword", err)
	}
}

// Ô "mật khẩu hiện tại" cũng là một cửa dò mật khẩu.
func TestChangePasswordIsRateLimited(t *testing.T) {
	t.Parallel()

	account := adminAccount(t)
	service := newService(t, accountUsers(t, account), newFakeSessions(t))

	for range auth.PasswordChangeRule.Limit {
		_, _ = service.ChangePassword(context.Background(), account.ID, "sai-mat-khau-hien-tai", newPassword)
	}
	_, err := service.ChangePassword(context.Background(), account.ID, servicePassword, newPassword)

	var limited *auth.RateLimitedError
	if !errors.As(err, &limited) {
		t.Fatalf("error = %v, want *RateLimitedError sau %d lần sai", err, auth.PasswordChangeRule.Limit)
	}
}
