package user

import "errors"

var (
	// ErrNotFound: người dùng không tồn tại hoặc đã bị xoá mềm.
	ErrNotFound = errors.New("user: not found")

	// ErrEmailTaken: email đã thuộc về một tài khoản chưa xoá khác.
	ErrEmailTaken = errors.New("user: email already taken")

	// ErrInvalidID: id truyền vào không phải UUID hợp lệ.
	ErrInvalidID = errors.New("user: invalid id")

	// ErrGoogleAccountTaken: tài khoản Google này đã gắn với người dùng khác.
	ErrGoogleAccountTaken = errors.New("user: google account already linked to another user")

	// ErrGoogleAlreadyLinked: người dùng này đã gắn một tài khoản Google khác.
	ErrGoogleAlreadyLinked = errors.New("user: user already linked to another google account")
)
