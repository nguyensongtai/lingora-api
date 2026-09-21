// Package course chứa toàn bộ nghiệp vụ khoá học và danh sách bài học của khoá.
package course

import "time"

// Level là trình độ CEFR của khoá học.
type Level string

const (
	LevelA1 Level = "A1"
	LevelA2 Level = "A2"
	LevelB1 Level = "B1"
	LevelB2 Level = "B2"
	LevelC1 Level = "C1"
	LevelC2 Level = "C2"
)

// Valid cho biết giá trị có nằm trong enum course_level của database hay không.
func (l Level) Valid() bool {
	switch l {
	case LevelA1, LevelA2, LevelB1, LevelB2, LevelC1, LevelC2:
		return true
	default:
		return false
	}
}

// Status là trạng thái xuất bản của khoá học.
type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
)

// Valid cho biết giá trị có nằm trong enum course_status của database hay không.
func (s Status) Valid() bool {
	switch s {
	case StatusDraft, StatusPublished:
		return true
	default:
		return false
	}
}

// Course là model nghiệp vụ; id dùng string để tầng trên không phụ thuộc pgtype.
type Course struct {
	ID            string
	Slug          string
	Title         string
	Description   string
	Level         Level
	Status        Status
	CoverImageURL *string
	// Position là thứ tự trong bậc của khoá, do người soạn đặt.
	Position  int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Lesson là bài học thuộc một khoá, sắp xếp theo Position tăng dần.
type Lesson struct {
	ID        string
	CourseID  string
	Slug      string
	Title     string
	Position  int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateParams là dữ liệu đã hợp lệ để tạo khoá học.
type CreateParams struct {
	Slug          string
	Title         string
	Description   string
	Level         Level
	Status        Status
	CoverImageURL *string
}

// UpdateParams là partial update: trường nil nghĩa là giữ nguyên.
// ClearCoverImage tách riêng vì nil không phân biệt được "bỏ qua" với "xoá ảnh".
type UpdateParams struct {
	Slug            *string
	Title           *string
	Description     *string
	Level           *Level
	Status          *Status
	CoverImageURL   *string
	ClearCoverImage bool
}

// Cursor là vị trí keyset của bản ghi cuối trang trước.
type Cursor struct {
	CreatedAt time.Time
	ID        string
}

// ListFilter mô tả một trang danh sách khoá học.
type ListFilter struct {
	Status   *Status
	Level    *Level
	Cursor   *Cursor
	PageSize int32
}

// Page là một trang kết quả; NextCursor nil nghĩa là đã hết dữ liệu.
type Page struct {
	Items      []Course
	NextCursor *Cursor
}

// LessonCreateParams là dữ liệu đã hợp lệ để thêm một bài vào khoá.
// Position không có ở đây: bài mới luôn đứng cuối.
type LessonCreateParams struct {
	Slug  string
	Title string
}

// LessonUpdateParams là partial update; nil nghĩa là giữ nguyên.
type LessonUpdateParams struct {
	Slug  *string
	Title *string
}
