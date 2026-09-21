package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nguyensongtai/lingora-api/internal/httpx"
)

const allowedOrigin = "http://localhost:3000"

func corsHandler(t *testing.T) (http.Handler, *bool) {
	t.Helper()

	reached := false
	handler := httpx.CORS([]string{allowedOrigin, "https://lingora.vn"})(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			reached = true
			w.WriteHeader(http.StatusOK)
		}),
	)
	return handler, &reached
}

func request(handler http.Handler, method, origin, requestMethod string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/v1/courses", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if requestMethod != "" {
		req.Header.Set("Access-Control-Request-Method", requestMethod)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	t.Parallel()

	handler, reached := corsHandler(t)

	recorder := request(handler, http.MethodGet, allowedOrigin, "")
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != allowedOrigin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, allowedOrigin)
	}
	if !*reached {
		t.Error("request did not reach the next handler")
	}
}

func TestCORSDoesNotEchoUnknownOrigin(t *testing.T) {
	t.Parallel()

	handler, reached := corsHandler(t)

	recorder := request(handler, http.MethodGet, "https://ke-tan-cong.example", "")
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want it unset for an unknown origin", got)
	}
	// Request thường vẫn đi tiếp: chặn là việc của trình duyệt, không phải của server.
	if !*reached {
		t.Error("a same-origin style request was blocked, want it to pass through")
	}
}

func TestCORSAlwaysVariesOnOrigin(t *testing.T) {
	t.Parallel()

	handler, _ := corsHandler(t)

	for _, origin := range []string{allowedOrigin, "https://ke-tan-cong.example", ""} {
		recorder := request(handler, http.MethodGet, origin, "")
		if vary := recorder.Header().Values("Vary"); !containsValue(vary, "Origin") {
			t.Errorf("origin %q: Vary = %v, want it to include Origin", origin, vary)
		}
	}
}

func TestCORSAnswersPreflightFromAllowedOrigin(t *testing.T) {
	t.Parallel()

	handler, reached := corsHandler(t)

	recorder := request(handler, http.MethodOptions, allowedOrigin, http.MethodPatch)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", recorder.Code)
	}
	if *reached {
		t.Error("preflight reached the next handler, want it answered by the middleware")
	}

	methods := recorder.Header().Get("Access-Control-Allow-Methods")
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete} {
		if !strings.Contains(methods, method) {
			t.Errorf("Access-Control-Allow-Methods = %q, want it to include %s", methods, method)
		}
	}
	if headers := recorder.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(headers, "Authorization") {
		t.Errorf("Access-Control-Allow-Headers = %q, want it to include Authorization", headers)
	}
	if recorder.Header().Get("Access-Control-Max-Age") == "" {
		t.Error("Access-Control-Max-Age is unset, want the preflight to be cacheable")
	}
}

func TestCORSRejectsPreflightFromUnknownOrigin(t *testing.T) {
	t.Parallel()

	handler, reached := corsHandler(t)

	recorder := request(handler, http.MethodOptions, "https://ke-tan-cong.example", http.MethodPost)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", recorder.Code)
	}
	if *reached {
		t.Error("rejected preflight still reached the next handler")
	}
}

func containsValue(values []string, want string) bool {
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			if strings.TrimSpace(part) == want {
				return true
			}
		}
	}
	return false
}
