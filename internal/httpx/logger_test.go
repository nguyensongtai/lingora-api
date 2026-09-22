package httpx_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/nguyensongtai/lingora-api/internal/httpx"
)

// capture chạy một request qua RequestLogger và trả về các dòng log JSON.
//
// Nó thay slog.Default() nên những test dùng nó KHÔNG được t.Parallel(): hai
// test song song sẽ cùng tráo một biến toàn cục, và test này từng đỏ đúng vì
// lý do đó — log của nó chảy vào buffer của test kia.
func capture(t *testing.T, handler http.Handler, method, target string) []map[string]any {
	t.Helper()

	var buffer bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))
	t.Cleanup(func() { slog.SetDefault(previous) })

	// RequestID để kiểm middleware có gắn được id vào dòng log hay không.
	wrapped := middleware.RequestID(httpx.RequestLogger(handler))
	wrapped.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, target, nil))

	var lines []map[string]any
	for _, raw := range bytes.Split(bytes.TrimSpace(buffer.Bytes()), []byte("\n")) {
		if len(raw) == 0 {
			continue
		}
		var line map[string]any
		if err := json.Unmarshal(raw, &line); err != nil {
			t.Fatalf("dòng log không phải JSON: %v (%s)", err, raw)
		}
		lines = append(lines, line)
	}
	return lines
}

func respondWith(status int, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
}

func TestRequestLoggerRecordsTheRequest(t *testing.T) {
	lines := capture(t, respondWith(http.StatusOK, "xin chao"), http.MethodGet, "/v1/courses")

	if len(lines) != 1 {
		t.Fatalf("số dòng log = %d, want 1", len(lines))
	}
	line := lines[0]

	for field, want := range map[string]any{
		"method": "GET",
		"path":   "/v1/courses",
		"status": float64(http.StatusOK),
		"bytes":  float64(len("xin chao")),
		"level":  "INFO",
	} {
		if line[field] != want {
			t.Errorf("%s = %v, want %v", field, line[field], want)
		}
	}
	if line["request_id"] == "" || line["request_id"] == nil {
		t.Error("thiếu request_id: không đối chiếu được với log của handler")
	}
	if _, ok := line["took_ms"]; !ok {
		t.Error("thiếu took_ms")
	}
}

// Handler không gọi WriteHeader thì net/http vẫn trả 200; log phải nói đúng thế
// chứ không phải 0.
func TestRequestLoggerReportsTheImplicitOK(t *testing.T) {
	silent := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

	lines := capture(t, silent, http.MethodGet, "/v1/courses")

	if lines[0]["status"] != float64(http.StatusOK) {
		t.Errorf("status = %v, want 200", lines[0]["status"])
	}
}

func TestRequestLoggerPicksLevelFromStatus(t *testing.T) {
	cases := map[int]string{
		http.StatusOK:                  "INFO",
		http.StatusNoContent:           "INFO",
		http.StatusBadRequest:          "WARN",
		http.StatusUnauthorized:        "WARN",
		http.StatusTooManyRequests:     "WARN",
		http.StatusInternalServerError: "ERROR",
		http.StatusServiceUnavailable:  "ERROR",
	}

	for status, want := range cases {
		t.Run(http.StatusText(status), func(t *testing.T) {
			lines := capture(t, respondWith(status, ""), http.MethodGet, "/v1/courses")

			if lines[0]["level"] != want {
				t.Errorf("level = %v, want %v", lines[0]["level"], want)
			}
		})
	}
}

// Probe bị gọi vài giây một lần; để chúng ở mức Info thì log thật bị nhấn chìm.
func TestRequestLoggerSkipsProbes(t *testing.T) {
	for _, path := range []string{"/healthz", "/readyz"} {
		t.Run(path, func(t *testing.T) {
			lines := capture(t, respondWith(http.StatusOK, ""), http.MethodGet, path)

			if len(lines) != 0 {
				t.Errorf("số dòng log = %d, want 0 cho %s", len(lines), path)
			}
		})
	}
}

// Query string do client đặt và log thường được chuyển đi nơi khác lưu.
func TestRequestLoggerOmitsTheQueryString(t *testing.T) {
	lines := capture(t, respondWith(http.StatusOK, ""), http.MethodGet,
		"/v1/courses?cursor=bi-mat&status=published")

	if lines[0]["path"] != "/v1/courses" {
		t.Errorf("path = %v, want /v1/courses không kèm query", lines[0]["path"])
	}
	for field, value := range lines[0] {
		if text, ok := value.(string); ok && bytes.Contains([]byte(text), []byte("bi-mat")) {
			t.Errorf("query string lọt vào field %s: %v", field, value)
		}
	}
}
