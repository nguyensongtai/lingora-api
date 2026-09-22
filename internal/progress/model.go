// Package progress ghi nhận việc người học đánh dấu hoàn thành từng bài, và
// suy ra XP, chuỗi ngày học từ chính những mốc hoàn thành đó.
package progress

import "time"

// CourseProgress là tỉ lệ hoàn thành của một khoá. LessonCount là mẫu số nên
// vẫn có mặt cả khi khoá chưa có bài nào — lúc đó bằng 0.
type CourseProgress struct {
	CourseID       string
	LessonCount    int64
	CompletedCount int64
}

// DayActivity là một ngày trong tuần gần nhất, tính theo múi giờ Việt Nam.
type DayActivity struct {
	Day       time.Time
	Completed int64
}

// Snapshot là toàn bộ tiến độ của một người học, đủ để dựng mọi màn hiện có
// mà không phải gọi thêm lần nào.
type Snapshot struct {
	Courses            []CourseProgress
	CompletedLessonIDs []string
	// LatestCourseID là khoá được đụng tới gần đây nhất; nil khi chưa học gì.
	LatestCourseID *string

	// TodayXP và GoalXP tính bằng XP; StreakDays là số ngày học liên tiếp.
	TodayXP     int64
	GoalXP      int64
	XPPerLesson int64
	StreakDays  int64
	// Week là bảy ngày gần nhất theo thứ tự cũ trước, mới sau.
	Week []DayActivity
}

// History là bức tranh dài hạn của một người học, dùng cho màn Tiến độ.
// Snapshot trả lời "hôm nay thế nào", History trả lời "từ trước tới giờ thế nào".
type History struct {
	// Days là dải ngày liên tục kết thúc ở hôm nay, cũ trước mới sau, kể cả
	// ngày không học — biểu đồ cần đủ cột, khoảng trống cũng là thông tin.
	Days []DayActivity

	// TotalLessons đếm trên cả đời chứ không cộng từ Days: Days bị cắt ở
	// DailyActivityLimit ngày.
	TotalLessons int64
	TotalXP      int64

	// ActiveDays và LongestStreak tính trong phạm vi Days.
	ActiveDays    int64
	CurrentStreak int64
	LongestStreak int64

	XPPerLesson int64
	GoalXP      int64
}

// Giới hạn số ngày một lần đọc History trả về.
const (
	DefaultHistoryDays int64 = 30
	MaxHistoryDays     int64 = 365
)

// XPPerLesson: mỗi bài đánh dấu xong được 20 XP. Con số lấy từ chính design —
// mục tiêu 50 XP, đã được 30, và dòng chữ dưới vòng tròn ghi "Còn 20 XP · một
// bài học nữa".
const XPPerLesson int64 = 20

// DefaultGoalXP là mục tiêu mỗi ngày, mặc định trong design. Chưa có màn cài
// đặt nên tạm là hằng số chung; khi nào cho người dùng đổi thì nó thành một cột
// của users.
const DefaultGoalXP int64 = 50

// StreakLocation là múi giờ dùng để cắt ngày. Cố định theo Việt Nam chứ không
// theo giờ máy chủ hay giờ trình duyệt: một chuỗi ngày học phải giữ nguyên dù
// người dùng đi công tác hay máy chủ đổi vùng.
var StreakLocation = time.FixedZone("ICT", 7*60*60)
