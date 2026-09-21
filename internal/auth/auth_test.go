package auth_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/nguyensongtai/lingora-api/internal/auth"
)

const testSecret = "secret-du-dai-cho-hs256-toi-thieu-32-byte"

func sign(t *testing.T, method jwt.SigningMethod, key any, claims jwt.Claims) string {
	t.Helper()

	token, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func adminClaims(expiry time.Time) jwt.MapClaims {
	return jwt.MapClaims{
		"sub":  "user-1",
		"role": auth.RoleAdmin,
		"exp":  expiry.Unix(),
	}
}

// protected trả về handler đã bọc middleware, kèm cờ cho biết next có chạy không.
func protected(t *testing.T) (http.Handler, *bool) {
	t.Helper()

	verifier, err := auth.NewVerifier(testSecret)
	if err != nil {
		t.Fatalf("NewVerifier() returned error: %v", err)
	}

	reached := false
	handler := verifier.RequireRole(auth.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		claims, ok := auth.ClaimsFrom(r.Context())
		if !ok {
			t.Error("claims missing from request context")
		}
		if claims.Subject != "user-1" {
			t.Errorf("claims.Subject = %q, want user-1", claims.Subject)
		}
		w.WriteHeader(http.StatusOK)
	}))
	return handler, &reached
}

func call(handler http.Handler, authorization string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/courses", nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func TestNewVerifierRejectsShortSecret(t *testing.T) {
	t.Parallel()

	if _, err := auth.NewVerifier("qua-ngan"); !errors.Is(err, auth.ErrWeakSecret) {
		t.Fatalf("error = %v, want it to match ErrWeakSecret", err)
	}
}

func TestRequireRoleAcceptsValidAdminToken(t *testing.T) {
	t.Parallel()

	handler, reached := protected(t)
	token := sign(t, jwt.SigningMethodHS256, []byte(testSecret), adminClaims(time.Now().Add(time.Hour)))

	if recorder := call(handler, "Bearer "+token); recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", recorder.Code, recorder.Body.String())
	}
	if !*reached {
		t.Error("next handler was not reached")
	}
}

func TestRequireRoleRejectsBadTokens(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		authorization func(t *testing.T) string
		wantStatus    int
	}{
		"không có header": {
			authorization: func(*testing.T) string { return "" },
			wantStatus:    http.StatusUnauthorized,
		},
		"sai scheme": {
			authorization: func(t *testing.T) string {
				return "Basic " + sign(t, jwt.SigningMethodHS256, []byte(testSecret), adminClaims(time.Now().Add(time.Hour)))
			},
			wantStatus: http.StatusUnauthorized,
		},
		"chữ ký sai khoá": {
			authorization: func(t *testing.T) string {
				return "Bearer " + sign(t, jwt.SigningMethodHS256, []byte("khoa-khac-nhung-van-du-dai-32-byte!!"), adminClaims(time.Now().Add(time.Hour)))
			},
			wantStatus: http.StatusUnauthorized,
		},
		"token hết hạn": {
			authorization: func(t *testing.T) string {
				return "Bearer " + sign(t, jwt.SigningMethodHS256, []byte(testSecret), adminClaims(time.Now().Add(-time.Minute)))
			},
			wantStatus: http.StatusUnauthorized,
		},
		"thiếu exp": {
			authorization: func(t *testing.T) string {
				return "Bearer " + sign(t, jwt.SigningMethodHS256, []byte(testSecret), jwt.MapClaims{"sub": "user-1", "role": auth.RoleAdmin})
			},
			wantStatus: http.StatusUnauthorized,
		},
		"alg none": {
			authorization: func(t *testing.T) string {
				return "Bearer " + sign(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, adminClaims(time.Now().Add(time.Hour)))
			},
			wantStatus: http.StatusUnauthorized,
		},
		"role không đủ quyền": {
			authorization: func(t *testing.T) string {
				return "Bearer " + sign(t, jwt.SigningMethodHS256, []byte(testSecret), jwt.MapClaims{
					"sub": "user-1", "role": "student", "exp": time.Now().Add(time.Hour).Unix(),
				})
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			handler, reached := protected(t)

			recorder := call(handler, tc.authorization(t))
			if recorder.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", recorder.Code, tc.wantStatus, recorder.Body.String())
			}
			if *reached {
				t.Error("next handler ran although the request should have been rejected")
			}
			if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q, want the shared JSON error body", contentType)
			}
		})
	}
}
