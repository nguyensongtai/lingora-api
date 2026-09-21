package user

import "errors"

var (
	// ErrNotFound: người dùng không tồn tại hoặc đã bị xoá mềm.
	ErrNotFound = errors.New("user: not found")

	// ErrEmailTaken: email đã thuộc về một tài khoản chưa xoá khác.
	ErrEmailTaken = errors.New("user: email already taken")

	// ErrInvalidID: id truyền vào không phải UUID hợp lệ.
	ErrInvalidID = errors.New("user: invalid id")
)
