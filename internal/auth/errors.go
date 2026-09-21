package auth

import "errors"

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
)
