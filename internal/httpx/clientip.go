package httpx

import (
	"crypto/subtle"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// Header BFF dùng để nói với API ai thật sự đang gọi. Đăng nhập, đăng ký và
// đăng nhập Google đi qua Route Handler của Next, nên nếu không có hai header
// này thì API chỉ thấy IP của máy chủ web — và mọi người dùng chung một hạn
// mức đăng nhập theo IP của cả site.
const (
	BFFSecretHeader   = "X-Lingora-Bff"
	BFFClientIPHeader = "X-Lingora-Client-Ip"
)

// ClientIPConfig nói request được tin nguồn IP nào.
type ClientIPConfig struct {
	// ProxyHeader là header mà proxy phía trước ĐẢM BẢO tự đặt, ví dụ
	// "X-Forwarded-For" (lấy giá trị đầu tiên) hay "X-Real-IP". Rỗng nghĩa là
	// không có proxy đáng tin: dùng đúng địa chỉ kết nối tới.
	ProxyHeader string
	// BFFSecret bật header X-Lingora-Client-Ip. Rỗng thì header đó bị bỏ qua.
	BFFSecret string
}

// ClientIP đặt r.RemoteAddr thành IP của người dùng thật, theo đúng những
// nguồn cấu hình cho phép — để hạn mức và log phía sau chỉ đọc RemoteAddr.
//
// Thay cho middleware.RealIP của chi: nó tin True-Client-IP, X-Real-IP và
// X-Forwarded-For từ BẤT KỲ ai, nên khi API ra internet thì kẻ dò mật khẩu chỉ
// cần đổi header mỗi lần là thoát hạn mức theo IP.
func ClientIP(cfg ClientIPConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ip, ok := resolveClientIP(r, cfg); ok {
				r.RemoteAddr = net.JoinHostPort(ip.String(), "0")
			}
			// Không để header của BFF đi tiếp: không ai ở phía sau được đọc lại
			// secret, và một giá trị không được tin thì không nên nằm đó.
			r.Header.Del(BFFSecretHeader)
			r.Header.Del(BFFClientIPHeader)
			next.ServeHTTP(w, r)
		})
	}
}

func resolveClientIP(r *http.Request, cfg ClientIPConfig) (netip.Addr, bool) {
	if cfg.BFFSecret != "" {
		secret := r.Header.Get(BFFSecretHeader)
		if secret != "" && subtle.ConstantTimeCompare([]byte(secret), []byte(cfg.BFFSecret)) == 1 {
			if ip, err := netip.ParseAddr(strings.TrimSpace(r.Header.Get(BFFClientIPHeader))); err == nil {
				return ip, true
			}
		}
	}

	if cfg.ProxyHeader != "" {
		raw := r.Header.Get(cfg.ProxyHeader)
		// X-Forwarded-For là một chuỗi; proxy đáng tin đặt IP người dùng ở đầu.
		first, _, _ := strings.Cut(raw, ",")
		if ip, err := netip.ParseAddr(strings.TrimSpace(first)); err == nil {
			return ip, true
		}
	}
	return netip.Addr{}, false
}
