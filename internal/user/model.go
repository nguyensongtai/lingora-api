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
// không bao giờ được đưa ra ngoài tầng HTTP; nil nghĩa là tài khoản chỉ đăng
// nhập bằng Google.
type User struct {
	ID           string
	Email        string
	PasswordHash *string
	DisplayName  string
	Role         Role
	GoogleSub    *string
	// DailyGoalXP là mục tiêu XP mỗi ngày, một trong DailyGoals.
	DailyGoalXP int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DailyGoals là ba mức mục tiêu XP mỗi ngày được chọn — khớp CHECK
// users_daily_goal_xp_allowed.
var DailyGoals = [...]int{20, 50, 100}

// ValidDailyGoal cho biết một mức có nằm trong DailyGoals hay không.
func ValidDailyGoal(xp int) bool {
	for _, goal := range DailyGoals {
		if goal == xp {
			return true
		}
	}
	return false
}

// ProfileParams là partial update hồ sơ; nil nghĩa là giữ nguyên. Giá trị đã
// được validate ở tầng trên.
type ProfileParams struct {
	DisplayName *string
	DailyGoalXP *int
}

// HasPassword cho biết tài khoản có đăng nhập bằng mật khẩu được hay không.
func (u User) HasPassword() bool {
	return u.PasswordHash != nil
}

// CreateParams là dữ liệu đã hợp lệ để tạo người dùng; mật khẩu phải được băm
// từ trước, repo không tự băm. Đúng một trong hai — PasswordHash hoặc
// GoogleSub — phải khác nil, và database cũng kiểm lại điều đó.
type CreateParams struct {
	Email        string
	PasswordHash *string
	DisplayName  string
	Role         Role
	GoogleSub    *string
}
