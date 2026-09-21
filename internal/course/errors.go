package course

import "errors"

var (
	// ErrNotFound: khoá học không tồn tại hoặc đã bị xoá mềm.
	ErrNotFound = errors.New("course: not found")

	// ErrSlugTaken: slug đã được một khoá học chưa xoá khác sử dụng.
	ErrSlugTaken = errors.New("course: slug already taken")

	// ErrInvalidID: id truyền vào không phải UUID hợp lệ.
	ErrInvalidID = errors.New("course: invalid id")
)
