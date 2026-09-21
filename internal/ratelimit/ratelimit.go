// Package ratelimit đếm số lần thử trong một cửa sổ thời gian, dùng để chặn dò
// mật khẩu.
package ratelimit

import (
	"context"
	"time"
)

// Decision là kết quả một lần hỏi. RetryAfter chỉ có ý nghĩa khi bị từ chối.
type Decision struct {
	Allowed    bool
	Remaining  int
	RetryAfter time.Duration
}

// Limiter đếm lượt theo khoá. Khai báo ở đây chứ không ở phía consumer vì cả
// middleware lẫn tầng nghiệp vụ đều dùng chung, và bản cài đặt sẽ đổi khi có
// một kho dùng chung giữa nhiều instance.
type Limiter interface {
	// Allow ghi nhận một lượt và cho biết có được đi tiếp hay không.
	Allow(ctx context.Context, key string, limit int, window time.Duration) (Decision, error)
	// Reset xoá bộ đếm của một khoá, dùng khi đăng nhập thành công.
	Reset(ctx context.Context, key string) error
}

// Rule là một hạn mức đã đặt tên, để chỗ gọi không phải nhớ hai con số.
type Rule struct {
	Name   string
	Limit  int
	Window time.Duration
}

// Key ghép tên hạn mức với định danh, tránh hai hạn mức khác nhau đụng khoá.
func (r Rule) Key(identifier string) string {
	return r.Name + ":" + identifier
}
