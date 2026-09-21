package httpx

import (
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/nguyensongtai/lingora-api/internal/ratelimit"
)

// RateLimitByIP chặn theo địa chỉ IP gọi tới. Dùng cho những route mà một lần
// gọi tốn kém hoặc đáng nghi khi lặp nhiều: đăng nhập, đăng ký.
//
// Đây là lớp chặn thô. Hạn mức riêng cho từng tài khoản nằm ở tầng nghiệp vụ,
// nơi biết được email người dùng gõ.
func RateLimitByIP(limiter ratelimit.Limiter, rule ratelimit.Rule) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			decision, err := limiter.Allow(r.Context(), rule.Key(clientIP(r)), rule.Limit, rule.Window)
			if err != nil {
				// Limiter hỏng thì cho đi tiếp: chặn hết người dùng vì một kho
				// đếm trục trặc còn tệ hơn tạm thời mất lớp chặn này.
				slog.ErrorContext(r.Context(), "rate limiter failed, allowing request", slog.Any("error", err))
				next.ServeHTTP(w, r)
				return
			}

			if !decision.Allowed {
				WriteRateLimited(w, decision.RetryAfter)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// WriteRateLimited trả 429 kèm Retry-After để client biết chờ bao lâu.
func WriteRateLimited(w http.ResponseWriter, retryAfter time.Duration) {
	seconds := int(retryAfter.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(seconds))

	Error(w, http.StatusTooManyRequests, CodeRateLimited,
		"Bạn đã thử quá nhiều lần, vui lòng chờ một lát rồi thử lại.", nil)
}

// clientIP lấy phần host của RemoteAddr. Middleware RealIP đã chuẩn hoá giá trị
// này từ header của proxy trước khi request tới đây.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
