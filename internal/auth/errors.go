package auth

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var (
	// ErrWeakSecret: secret quá ngắn để ký HS256 an toàn.
	ErrWeakSecret = errors.New("auth: jwt secret must be at least 32 bytes")

	// ErrInvalidCredentials: sai email hoặc sai mật khẩu. Cố ý không phân biệt
	// hai trường hợp để không lộ email nào đang tồn tại.
	ErrInvalidCredentials = errors.New("auth: invalid credentials")

	// ErrSessionNotFound: refresh token không tồn tại, đã hết hạn hoặc đã thu hồi.
	ErrSessionNotFound = errors.New("auth: session not found")

	// ErrInvalidUserID: id người dùng không phải UUID hợp lệ.
	ErrInvalidUserID = errors.New("auth: invalid user id")

	// ErrGoogleNotConfigured: chưa đặt GOOGLE_CLIENT_ID/SECRET nên tính năng tắt.
	ErrGoogleNotConfigured = errors.New("auth: google sign-in is not configured")

	// ErrGoogleExchangeFailed: Google từ chối code, hoặc token trả về không hợp lệ.
	ErrGoogleExchangeFailed = errors.New("auth: google exchange failed")

	// ErrGoogleEmailUnverified: Google chưa xác minh email của tài khoản đó.
	ErrGoogleEmailUnverified = errors.New("auth: google email not verified")

	// ErrRateLimited: thử quá nhiều lần trong một khoảng thời gian ngắn. Dùng
	// errors.As với *RateLimitedError để lấy thời gian phải chờ.
	ErrRateLimited = errors.New("auth: too many attempts")

	// ErrValidation: dữ liệu đăng ký không hợp lệ. Dùng errors.As với
	// *ValidationError để lấy chi tiết từng field.
	ErrValidation = errors.New("auth: validation failed")
)

// ValidationError gom lỗi theo từng field để handler map thẳng sang details
// của body lỗi {code, message, details}.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	names := make([]string, 0, len(e.Fields))
	for name := range e.Fields {
		names = append(names, name)
	}
	sort.Strings(names)

	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, fmt.Sprintf("%s: %s", name, e.Fields[name]))
	}
	return "auth: validation failed (" + strings.Join(parts, "; ") + ")"
}

// Unwrap cho phép errors.Is(err, ErrValidation) nhận diện mọi lỗi validate.
func (e *ValidationError) Unwrap() error { return ErrValidation }

type validationBuilder struct {
	fields map[string]string
}

func (b *validationBuilder) add(field, message string) {
	if b.fields == nil {
		b.fields = make(map[string]string)
	}
	if _, exists := b.fields[field]; !exists {
		b.fields[field] = message
	}
}

// err trả về nil khi không có lỗi nào được ghi nhận.
func (b *validationBuilder) err() error {
	if len(b.fields) == 0 {
		return nil
	}
	return &ValidationError{Fields: b.fields}
}

// RateLimitedError mang theo thời gian còn phải chờ để handler đặt Retry-After.
type RateLimitedError struct {
	RetryAfter time.Duration
}

func (e *RateLimitedError) Error() string {
	return "auth: too many attempts, retry after " + e.RetryAfter.String()
}

func (e *RateLimitedError) Unwrap() error { return ErrRateLimited }
