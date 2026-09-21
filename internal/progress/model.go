// Package progress ghi nhận việc người học đánh dấu hoàn thành từng bài.
package progress

// CourseProgress là tỉ lệ hoàn thành của một khoá. LessonCount là mẫu số nên
// vẫn có mặt cả khi khoá chưa có bài nào — lúc đó bằng 0.
type CourseProgress struct {
	CourseID       string
	LessonCount    int64
	CompletedCount int64
}

// Snapshot là toàn bộ tiến độ của một người học, đủ để dựng mọi màn hiện có
// mà không phải gọi thêm lần nào.
type Snapshot struct {
	Courses            []CourseProgress
	CompletedLessonIDs []string
	// LatestCourseID là khoá được đụng tới gần đây nhất; nil khi chưa học gì.
	LatestCourseID *string
}
