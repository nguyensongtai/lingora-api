package practice

import "errors"

var (
	// ErrNotFound: từ không tồn tại, hoặc người học chưa mở khoá nó. Hai chuyện
	// đó cố ý không phân biệt: nói "có từ này nhưng bạn chưa được học" là tiết
	// lộ nội dung bài chưa học.
	ErrNotFound = errors.New("practice: question not found for this user")

	// ErrInvalidKind: dạng câu hỏi không nằm trong enum.
	ErrInvalidKind = errors.New("practice: unknown question kind")
)
