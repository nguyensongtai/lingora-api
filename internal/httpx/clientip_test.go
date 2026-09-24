package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nguyensongtai/lingora-api/internal/httpx"
)

const bffSecret = "secret-cua-bff-du-dai-toi-thieu-32-byte"

// seen chạy middleware rồi trả về RemoteAddr và header mà handler phía sau thấy.
func seen(t *testing.T, cfg httpx.ClientIPConfig, headers map[string]string) (string, http.Header) {
	t.Helper()

	var addr string
	var header http.Header
	handler := httpx.ClientIP(cfg)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		addr, header = r.RemoteAddr, r.Header.Clone()
	}))

	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.RemoteAddr = "10.0.0.5:51234"
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	handler.ServeHTTP(httptest.NewRecorder(), req)
	return addr, header
}

func TestClientIPIgnoresForwardingHeadersByDefault(t *testing.T) {
	t.Parallel()

	// Không cấu hình proxy nào: mọi header tự khai đều bị bỏ qua. Đây chính là
	// lỗ của middleware.RealIP — đổi header là đổi được IP.
	addr, _ := seen(t, httpx.ClientIPConfig{}, map[string]string{
		"X-Forwarded-For": "203.0.113.9",
		"X-Real-IP":       "203.0.113.9",
		"True-Client-IP":  "203.0.113.9",
	})
	if addr != "10.0.0.5:51234" {
		t.Errorf("RemoteAddr = %q, want giữ nguyên địa chỉ kết nối", addr)
	}
}

func TestClientIPTakesTheFirstEntryOfTheTrustedHeader(t *testing.T) {
	t.Parallel()

	addr, _ := seen(t, httpx.ClientIPConfig{ProxyHeader: "X-Forwarded-For"}, map[string]string{
		"X-Forwarded-For": "203.0.113.9, 10.1.2.3",
	})
	if addr != "203.0.113.9:0" {
		t.Errorf("RemoteAddr = %q, want 203.0.113.9:0", addr)
	}

	// Header rác thì giữ địa chỉ kết nối, không đoán.
	addr, _ = seen(t, httpx.ClientIPConfig{ProxyHeader: "X-Forwarded-For"}, map[string]string{
		"X-Forwarded-For": "không-phải-ip",
	})
	if addr != "10.0.0.5:51234" {
		t.Errorf("header rác: RemoteAddr = %q, want giữ nguyên", addr)
	}
}

func TestClientIPTrustsTheBFFOnlyWithTheSecret(t *testing.T) {
	t.Parallel()

	cfg := httpx.ClientIPConfig{BFFSecret: bffSecret}

	addr, header := seen(t, cfg, map[string]string{
		httpx.BFFSecretHeader:   bffSecret,
		httpx.BFFClientIPHeader: "198.51.100.7",
	})
	if addr != "198.51.100.7:0" {
		t.Errorf("secret đúng: RemoteAddr = %q, want 198.51.100.7:0", addr)
	}
	if header.Get(httpx.BFFSecretHeader) != "" || header.Get(httpx.BFFClientIPHeader) != "" {
		t.Error("header của BFF lọt xuống handler phía sau")
	}

	addr, _ = seen(t, cfg, map[string]string{
		httpx.BFFSecretHeader:   "doan-bua",
		httpx.BFFClientIPHeader: "198.51.100.7",
	})
	if addr != "10.0.0.5:51234" {
		t.Errorf("secret sai: RemoteAddr = %q, want giữ nguyên", addr)
	}
}

// Không cấu hình secret thì header của BFF vô dụng, kể cả khi gửi chuỗi rỗng.
func TestClientIPWithoutASecretIgnoresTheBFFHeaders(t *testing.T) {
	t.Parallel()

	addr, _ := seen(t, httpx.ClientIPConfig{}, map[string]string{
		httpx.BFFSecretHeader:   "",
		httpx.BFFClientIPHeader: "198.51.100.7",
	})
	if addr != "10.0.0.5:51234" {
		t.Errorf("RemoteAddr = %q, want giữ nguyên", addr)
	}
}

// BFF có quyền ưu tiên hơn header của proxy: với request đi qua BFF, proxy chỉ
// thấy máy chủ web.
func TestClientIPPrefersTheBFFOverTheProxyHeader(t *testing.T) {
	t.Parallel()

	addr, _ := seen(t, httpx.ClientIPConfig{ProxyHeader: "X-Forwarded-For", BFFSecret: bffSecret}, map[string]string{
		"X-Forwarded-For":       "10.9.9.9",
		httpx.BFFSecretHeader:   bffSecret,
		httpx.BFFClientIPHeader: "198.51.100.7",
	})
	if addr != "198.51.100.7:0" {
		t.Errorf("RemoteAddr = %q, want IP người dùng do BFF gửi", addr)
	}
}
