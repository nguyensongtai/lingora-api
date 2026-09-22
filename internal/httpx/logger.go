package httpx

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestLogger ghi một dòng cho mỗi request đã hoàn tất.
//
// chi có sẵn middleware.Logger nhưng nó in text ra stdout, trong khi phần còn
// lại của API ghi JSON qua slog — hai định dạng lẫn nhau trong cùng một luồng
// log thì công cụ thu thập nào cũng chỉ đọc được một nửa.
//
// Mức log theo status để cái đáng chú ý nổi lên: 5xx là Error, 4xx là Warn
// (sai phía client, vẫn đáng theo dõi vì một tràng 401 là dấu hiệu dò mật
// khẩu), còn lại là Info.
//
// Cố ý KHÔNG ghi query string: nó chứa tham số do client đặt, và log thường
// được chuyển sang nơi khác lưu. Path đủ để biết chuyện gì xảy ra.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Endpoint kiểm tra sức khoẻ bị gọi vài giây một lần; để mức Info thì
		// chúng nhấn chìm mọi thứ khác.
		if isProbe(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		recorder := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		started := time.Now()

		next.ServeHTTP(recorder, r)

		status := recorder.Status()
		// Handler không gọi WriteHeader thì net/http vẫn trả 200.
		if status == 0 {
			status = http.StatusOK
		}

		slog.Log(r.Context(), levelFor(status), "http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", status),
			slog.Int("bytes", recorder.BytesWritten()),
			// Mili giây dạng số thực: slog.Duration ghi ra nano giây, đúng cho
			// máy nhưng ai đọc log cũng phải tự đếm chữ số.
			slog.Float64("took_ms", float64(time.Since(started).Microseconds())/1000),
			slog.String("ip", clientIP(r)),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
	})
}

func isProbe(path string) bool {
	return path == "/healthz" || path == "/readyz"
}

func levelFor(status int) slog.Level {
	switch {
	case status >= http.StatusInternalServerError:
		return slog.LevelError
	case status >= http.StatusBadRequest:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}
