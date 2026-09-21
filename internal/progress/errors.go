package progress

import "errors"

var (
	// ErrInvalidID: id truyền vào không phải UUID hợp lệ.
	ErrInvalidID = errors.New("progress: invalid id")

	// ErrLessonNotFound: bài học không tồn tại, đã xoá mềm, hoặc thuộc khoá đã xoá.
	ErrLessonNotFound = errors.New("progress: lesson not found")
)
