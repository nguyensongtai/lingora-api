package practice

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound: từ không tồn tại, hoặc người học chưa mở khoá nó. Hai chuyện
	// đó cố ý không phân biệt: nói "có từ này nhưng bạn chưa được học" là tiết
	// lộ nội dung bài chưa học.
	ErrNotFound = errors.New("practice: question not found for this user")

	// ErrInvalidKind: dạng câu hỏi không nằm trong enum.
	ErrInvalidKind = errors.New("practice: unknown question kind")

	// ErrInvalidID: id truyền vào không phải UUID hợp lệ.
	ErrInvalidID = errors.New("practice: invalid id")

	// ErrInvalidResult: kết quả gửi lên không khớp với một lượt luyện của bài.
	// Dùng errors.As với *ResultError để lấy chi tiết theo field.
	ErrInvalidResult = errors.New("practice: invalid result")
)

// ResultError gom lỗi theo field để handler trả thẳng ra details.
type ResultError struct {
	Fields map[string]string
}

func (e *ResultError) Error() string {
	return fmt.Sprintf("%s: %v", ErrInvalidResult, e.Fields)
}

// Unwrap cho phép errors.Is(err, ErrInvalidResult).
func (e *ResultError) Unwrap() error { return ErrInvalidResult }
