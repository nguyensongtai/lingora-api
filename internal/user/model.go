// Package user quản lý danh tính người dùng. Việc đăng nhập và phát hành token
// thuộc package auth.
package user

import "time"

// Role quyết định người dùng được làm gì.
type Role string

const (
	RoleAdmin   Role = "admin"
	RoleStudent Role = "student"
)

// Valid cho biết giá trị có nằm trong enum user_role của database hay không.
func (r Role) Valid() bool {
	switch r {
	case RoleAdmin, RoleStudent:
		return true
	default:
		return false
	}
}

// User là model nghiệp vụ. PasswordHash chỉ dùng trong quá trình xác thực và
// không bao giờ được đưa ra ngoài tầng HTTP.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	Role         Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CreateParams là dữ liệu đã hợp lệ để tạo người dùng; mật khẩu phải được băm
// từ trước, repo không tự băm.
type CreateParams struct {
	Email        string
	PasswordHash string
	DisplayName  string
	Role         Role
}
