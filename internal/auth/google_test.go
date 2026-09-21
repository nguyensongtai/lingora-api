package auth_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/nguyensongtai/lingora-api/internal/auth"
	"github.com/nguyensongtai/lingora-api/internal/user"
)

const testClientID = "client-id-cua-lingora.apps.googleusercontent.com"

// idToken dựng một id_token giả. Chỉ phần payload có ý nghĩa: Exchange cố ý
// không kiểm chữ ký, vì token đến thẳng từ endpoint của Google qua TLS.
func idToken(t *testing.T, claims map[string]any) string {
	t.Helper()

	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	encode := base64.RawURLEncoding.EncodeToString
	return encode([]byte(`{"alg":"RS256"}`)) + "." + encode(payload) + ".chu-ky-gia"
}

// googleStub đóng vai endpoint token của Google.
func googleStub(t *testing.T, claims map[string]any) (*auth.GoogleClient, *url.Values) {
	t.Helper()

	var received url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		received = r.PostForm

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"id_token": idToken(t, claims)})
	}))
	t.Cleanup(server.Close)

	client := auth.NewGoogleClient(auth.GoogleConfig{
		ClientID:     testClientID,
		ClientSecret: "client-secret",
		TokenURL:     server.URL,
	})
	return client, &received
}

func googleClaims(overrides map[string]any) map[string]any {
	claims := map[string]any{
		"iss":            "https://accounts.google.com",
		"aud":            testClientID,
		"sub":            "google-sub-123",
		"email":          "an@gmail.com",
		"email_verified": true,
		"name":           "Nguyễn An",
	}
	for key, value := range overrides {
		claims[key] = value
	}
	return claims
}

func TestExchangeSendsTheAuthorizationCode(t *testing.T) {
	t.Parallel()

	client, received := googleStub(t, googleClaims(nil))

	identity, err := client.Exchange(context.Background(), "ma-code", "https://lingora.vn/callback")
	if err != nil {
		t.Fatalf("Exchange() returned error: %v", err)
	}

	if got := received.Get("code"); got != "ma-code" {
		t.Errorf("code = %q, want %q", got, "ma-code")
	}
	if got := received.Get("grant_type"); got != "authorization_code" {
		t.Errorf("grant_type = %q, want authorization_code", got)
	}
	if got := received.Get("redirect_uri"); got != "https://lingora.vn/callback" {
		t.Errorf("redirect_uri = %q, want nó được chuyển tiếp nguyên vẹn", got)
	}
	if identity.Subject != "google-sub-123" || identity.Email != "an@gmail.com" {
		t.Errorf("identity = %+v, want sub và email từ claims", identity)
	}
}

func TestExchangeRejectsTokensForAnotherProject(t *testing.T) {
	t.Parallel()

	for name, claims := range map[string]map[string]any{
		"aud của dự án khác":    googleClaims(map[string]any{"aud": "client-id-cua-nguoi-khac"}),
		"iss không phải Google": googleClaims(map[string]any{"iss": "https://evil.example"}),
		"thiếu sub":             googleClaims(map[string]any{"sub": ""}),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, _ := googleStub(t, claims)

			_, err := client.Exchange(context.Background(), "ma-code", "https://lingora.vn/callback")
			if !errors.Is(err, auth.ErrGoogleExchangeFailed) {
				t.Fatalf("Exchange() error = %v, want ErrGoogleExchangeFailed", err)
			}
		})
	}
}

/* ---------- ba nhánh của SignInWithGoogle ---------- */

type stubExchanger struct {
	identity auth.GoogleIdentity
	err      error
}

func (s stubExchanger) Exchange(context.Context, string, string) (auth.GoogleIdentity, error) {
	return s.identity, s.err
}

func googleService(t *testing.T, users *fakeUsers, exchanger auth.Config) *auth.Service {
	t.Helper()

	exchanger.Secret = servicePassword + "-secret-du-dai-cho-hs256"
	exchanger.AccessTTL = 900_000_000_000
	exchanger.RefreshTTL = 900_000_000_000
	service, err := auth.NewService(users, newFakeSessions(t), exchanger)
	if err != nil {
		t.Fatalf("NewService() returned error: %v", err)
	}
	return service
}

// notFound là nhánh "chưa có tài khoản nào" của cả hai lần tra cứu.
func notFound(context.Context, string) (user.User, error) {
	return user.User{}, user.ErrNotFound
}

func verifiedIdentity() auth.GoogleIdentity {
	return auth.GoogleIdentity{
		Subject:       "google-sub-123",
		Email:         "an@gmail.com",
		EmailVerified: true,
		Name:          "Nguyễn An",
	}
}

func TestSignInWithGoogleUsesTheLinkedAccount(t *testing.T) {
	t.Parallel()

	linked := user.User{ID: "user-1", Email: "an@gmail.com", DisplayName: "An", Role: user.RoleStudent, GoogleSub: ptr("google-sub-123")}
	users := &fakeUsers{
		t:        t,
		byGoogle: func(context.Context, string) (user.User, error) { return linked, nil },
	}

	pair, err := googleService(t, users, auth.Config{Google: stubExchanger{identity: verifiedIdentity()}}).
		SignInWithGoogle(context.Background(), "ma-code", "https://lingora.vn/callback")
	if err != nil {
		t.Fatalf("SignInWithGoogle() returned error: %v", err)
	}

	if pair.User.ID != "user-1" {
		t.Errorf("user = %q, want tài khoản đã gắn", pair.User.ID)
	}
}

func TestSignInWithGoogleLinksAnExistingEmail(t *testing.T) {
	t.Parallel()

	existing := user.User{ID: "user-2", Email: "an@gmail.com", DisplayName: "An", Role: user.RoleStudent, PasswordHash: ptr("hash")}
	var linkedTo, linkedSub string
	users := &fakeUsers{
		t:       t,
		byEmail: func(context.Context, string) (user.User, error) { return existing, nil },
		linkGoogle: func(_ context.Context, id, sub string) (user.User, error) {
			linkedTo, linkedSub = id, sub
			existing.GoogleSub = &sub
			return existing, nil
		},
	}

	if _, err := googleService(t, users, auth.Config{Google: stubExchanger{identity: verifiedIdentity()}}).
		SignInWithGoogle(context.Background(), "ma-code", "https://lingora.vn/callback"); err != nil {
		t.Fatalf("SignInWithGoogle() returned error: %v", err)
	}

	if linkedTo != "user-2" || linkedSub != "google-sub-123" {
		t.Errorf("LinkGoogle(%q, %q), want (user-2, google-sub-123)", linkedTo, linkedSub)
	}
}

func TestSignInWithGoogleCreatesAPasswordlessAccount(t *testing.T) {
	t.Parallel()

	var created user.CreateParams
	users := &fakeUsers{
		t:       t,
		byEmail: notFound,
		create: func(_ context.Context, params user.CreateParams) (user.User, error) {
			created = params
			return user.User{ID: "user-3", Email: params.Email, Role: params.Role}, nil
		},
	}

	if _, err := googleService(t, users, auth.Config{Google: stubExchanger{identity: verifiedIdentity()}}).
		SignInWithGoogle(context.Background(), "ma-code", "https://lingora.vn/callback"); err != nil {
		t.Fatalf("SignInWithGoogle() returned error: %v", err)
	}

	if created.PasswordHash != nil {
		t.Error("tài khoản tạo từ Google không được có mật khẩu")
	}
	if created.GoogleSub == nil || *created.GoogleSub != "google-sub-123" {
		t.Errorf("GoogleSub = %v, want google-sub-123", created.GoogleSub)
	}
	if created.Role != user.RoleStudent {
		t.Errorf("role = %q, want student", created.Role)
	}
	if created.DisplayName != "Nguyễn An" {
		t.Errorf("display_name = %q, want tên Google trả về", created.DisplayName)
	}
}

func TestSignInWithGoogleRefusesUnverifiedEmail(t *testing.T) {
	t.Parallel()

	// Nếu không chặn, bất kỳ ai tạo được tài khoản Google mang email của người
	// khác sẽ chiếm được tài khoản Lingora của họ ở nhánh gắn theo email.
	identity := verifiedIdentity()
	identity.EmailVerified = false

	_, err := googleService(t, &fakeUsers{t: t}, auth.Config{Google: stubExchanger{identity: identity}}).
		SignInWithGoogle(context.Background(), "ma-code", "https://lingora.vn/callback")

	if !errors.Is(err, auth.ErrGoogleEmailUnverified) {
		t.Fatalf("error = %v, want ErrGoogleEmailUnverified", err)
	}
}

func TestSignInWithGoogleIsOffWhenUnconfigured(t *testing.T) {
	t.Parallel()

	service := googleService(t, &fakeUsers{t: t}, auth.Config{})

	if service.GoogleEnabled() {
		t.Error("GoogleEnabled() = true khi chưa cấu hình")
	}

	_, err := service.SignInWithGoogle(context.Background(), "ma-code", "https://lingora.vn/callback")
	if !errors.Is(err, auth.ErrGoogleNotConfigured) {
		t.Fatalf("error = %v, want ErrGoogleNotConfigured", err)
	}
}

func TestSignInWithGoogleFallsBackToTheEmailLocalPart(t *testing.T) {
	t.Parallel()

	identity := verifiedIdentity()
	identity.Name = "   "

	var created user.CreateParams
	users := &fakeUsers{
		t:       t,
		byEmail: notFound,
		create: func(_ context.Context, params user.CreateParams) (user.User, error) {
			created = params
			return user.User{ID: "user-4", Email: params.Email}, nil
		},
	}

	if _, err := googleService(t, users, auth.Config{Google: stubExchanger{identity: identity}}).
		SignInWithGoogle(context.Background(), "ma-code", "https://lingora.vn/callback"); err != nil {
		t.Fatalf("SignInWithGoogle() returned error: %v", err)
	}

	// Giao diện luôn cần một cái tên để hiển thị, không được để rỗng.
	if created.DisplayName != "an" {
		t.Errorf("display_name = %q, want %q", created.DisplayName, "an")
	}
}
