package course

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Giới hạn độ dài, khớp với ràng buộc của UI và của database.
const (
	maxSlugLen        = 120
	maxTitleLen       = 200
	maxDescriptionLen = 5000
	maxCoverURLLen    = 2048

	// DefaultPageSize được dùng khi client không truyền page size.
	DefaultPageSize int32 = 20
	// MaxPageSize chặn client yêu cầu trang quá lớn.
	MaxPageSize int32 = 100
)

// slugPattern khớp với CHECK courses_slug_format trong migration.
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// repository là những gì Service cần ở tầng lưu trữ. Interface khai báo ở phía
// consumer nên Repo không phải biết đến nó.
type repository interface {
	Create(ctx context.Context, params CreateParams) (Course, error)
	GetByID(ctx context.Context, id string) (Course, error)
	GetBySlug(ctx context.Context, slug string) (Course, error)
	List(ctx context.Context, filter ListFilter) ([]Course, error)
	Update(ctx context.Context, id string, params UpdateParams) (Course, error)
	SoftDelete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
	ListIDsByLevel(ctx context.Context, level Level) ([]string, error)
	ReorderCourses(ctx context.Context, level Level, courseIDs []string) (int64, error)
	ListLessons(ctx context.Context, courseID string) ([]Lesson, error)
	CreateLesson(ctx context.Context, courseID string, params LessonCreateParams) (Lesson, error)
	GetLesson(ctx context.Context, lessonID string) (Lesson, error)
	UpdateLesson(ctx context.Context, lessonID string, params LessonUpdateParams) (Lesson, error)
	SoftDeleteLesson(ctx context.Context, lessonID string) error
	ListLessonIDs(ctx context.Context, courseID string) ([]string, error)
	ReorderLessons(ctx context.Context, courseID string, lessonIDs []string) (int64, error)
}

// Service giữ nghiệp vụ khoá học: chuẩn hoá đầu vào, validate, phân trang.
type Service struct {
	repo repository
}

// NewService nhận repository đã sẵn sàng dùng.
func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

// Create chuẩn hoá và kiểm tra dữ liệu trước khi ghi. Status rỗng mặc định là draft.
func (s *Service) Create(ctx context.Context, params CreateParams) (Course, error) {
	params.Slug = strings.ToLower(strings.TrimSpace(params.Slug))
	params.Title = strings.TrimSpace(params.Title)
	params.Description = strings.TrimSpace(params.Description)
	params.CoverImageURL = trimOptional(params.CoverImageURL)
	if params.Status == "" {
		params.Status = StatusDraft
	}

	var v validationBuilder
	validateSlug(&v, params.Slug)
	validateTitle(&v, params.Title)
	validateDescription(&v, params.Description)
	validateCoverImageURL(&v, params.CoverImageURL)
	if !params.Level.Valid() {
		v.add("level", "phải là một trong A1, A2, B1, B2, C1, C2")
	}
	if !params.Status.Valid() {
		v.add("status", "phải là draft hoặc published")
	}
	if err := v.err(); err != nil {
		return Course{}, err
	}

	created, err := s.repo.Create(ctx, params)
	if err != nil {
		return Course{}, fmt.Errorf("create course: %w", err)
	}
	return created, nil
}

// Get đọc một khoá học theo id.
func (s *Service) Get(ctx context.Context, id string) (Course, error) {
	found, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Course{}, fmt.Errorf("get course: %w", err)
	}
	return found, nil
}

// GetBySlug đọc một khoá học theo slug, dùng cho URL thân thiện phía web.
func (s *Service) GetBySlug(ctx context.Context, slug string) (Course, error) {
	found, err := s.repo.GetBySlug(ctx, strings.ToLower(strings.TrimSpace(slug)))
	if err != nil {
		return Course{}, fmt.Errorf("get course by slug: %w", err)
	}
	return found, nil
}

// List trả về một trang khoá học. Repo được hỏi dư một bản ghi để biết còn
// trang sau hay không mà không cần đếm tổng.
func (s *Service) List(ctx context.Context, filter ListFilter) (Page, error) {
	var v validationBuilder
	if filter.Status != nil && !filter.Status.Valid() {
		v.add("status", "phải là draft hoặc published")
	}
	if filter.Level != nil && !filter.Level.Valid() {
		v.add("level", "phải là một trong A1, A2, B1, B2, C1, C2")
	}
	if err := v.err(); err != nil {
		return Page{}, err
	}

	pageSize := clampPageSize(filter.PageSize)
	query := filter
	query.PageSize = pageSize + 1

	courses, err := s.repo.List(ctx, query)
	if err != nil {
		return Page{}, fmt.Errorf("list courses: %w", err)
	}

	page := Page{Items: courses}
	if int32(len(courses)) > pageSize {
		page.Items = courses[:pageSize]
		last := page.Items[len(page.Items)-1]
		page.NextCursor = &Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return page, nil
}

// Update áp dụng partial update; chỉ những field khác nil mới được kiểm tra.
func (s *Service) Update(ctx context.Context, id string, params UpdateParams) (Course, error) {
	params.Slug = trimLowerOptional(params.Slug)
	params.Title = trimOptional(params.Title)
	params.Description = trimOptional(params.Description)
	params.CoverImageURL = trimOptional(params.CoverImageURL)

	var v validationBuilder
	if params.Slug != nil {
		validateSlug(&v, *params.Slug)
	}
	if params.Title != nil {
		validateTitle(&v, *params.Title)
	}
	if params.Description != nil {
		validateDescription(&v, *params.Description)
	}
	if params.CoverImageURL != nil {
		if params.ClearCoverImage {
			v.add("cover_image_url", "không thể vừa đặt giá trị vừa xoá ảnh bìa")
		}
		validateCoverImageURL(&v, params.CoverImageURL)
	}
	if params.Level != nil && !params.Level.Valid() {
		v.add("level", "phải là một trong A1, A2, B1, B2, C1, C2")
	}
	if params.Status != nil && !params.Status.Valid() {
		v.add("status", "phải là draft hoặc published")
	}
	if isEmptyUpdate(params) {
		v.add("body", "cần ít nhất một field để cập nhật")
	}
	if err := v.err(); err != nil {
		return Course{}, err
	}

	updated, err := s.repo.Update(ctx, id, params)
	if err != nil {
		return Course{}, fmt.Errorf("update course: %w", err)
	}
	return updated, nil
}

// Delete xoá mềm một khoá học.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("delete course: %w", err)
	}
	return nil
}

// ListLessons trả về bài học của một khoá. Khoá không tồn tại là ErrNotFound,
// khác hẳn với khoá tồn tại nhưng chưa có bài nào (danh sách rỗng).
func (s *Service) ListLessons(ctx context.Context, courseID string) ([]Lesson, error) {
	exists, err := s.repo.Exists(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("list lessons: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("list lessons of course %s: %w", courseID, ErrNotFound)
	}

	lessons, err := s.repo.ListLessons(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("list lessons: %w", err)
	}
	return lessons, nil
}

func clampPageSize(size int32) int32 {
	switch {
	case size <= 0:
		return DefaultPageSize
	case size > MaxPageSize:
		return MaxPageSize
	default:
		return size
	}
}

func isEmptyUpdate(params UpdateParams) bool {
	return params.Slug == nil &&
		params.Title == nil &&
		params.Description == nil &&
		params.Level == nil &&
		params.Status == nil &&
		params.CoverImageURL == nil &&
		!params.ClearCoverImage
}

func validateSlug(v *validationBuilder, slug string) {
	switch {
	case slug == "":
		v.add("slug", "không được rỗng")
	case utf8.RuneCountInString(slug) > maxSlugLen:
		v.add("slug", fmt.Sprintf("tối đa %d ký tự", maxSlugLen))
	case !slugPattern.MatchString(slug):
		v.add("slug", "chỉ gồm chữ thường, số và dấu gạch ngang giữa các từ")
	}
}

func validateTitle(v *validationBuilder, title string) {
	switch {
	case title == "":
		v.add("title", "không được rỗng")
	case utf8.RuneCountInString(title) > maxTitleLen:
		v.add("title", fmt.Sprintf("tối đa %d ký tự", maxTitleLen))
	}
}

func validateDescription(v *validationBuilder, description string) {
	if utf8.RuneCountInString(description) > maxDescriptionLen {
		v.add("description", fmt.Sprintf("tối đa %d ký tự", maxDescriptionLen))
	}
}

func validateCoverImageURL(v *validationBuilder, raw *string) {
	if raw == nil || *raw == "" {
		return
	}
	if utf8.RuneCountInString(*raw) > maxCoverURLLen {
		v.add("cover_image_url", fmt.Sprintf("tối đa %d ký tự", maxCoverURLLen))
		return
	}

	parsed, err := url.Parse(*raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		v.add("cover_image_url", "phải là URL http hoặc https hợp lệ")
	}
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func trimLowerOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.ToLower(strings.TrimSpace(*value))
	return &trimmed
}

// ReorderCourses đặt lại thứ tự khoá trong một bậc. Danh sách gửi lên phải là
// đúng tập khoá hiện có của bậc đó: thiếu, thừa hay trùng đều bị từ chối, vì
// một lần sắp xếp thiếu khoá sẽ để lại những khoá mang position cũ lẫn vào giữa.
func (s *Service) ReorderCourses(ctx context.Context, level Level, courseIDs []string) error {
	if !level.Valid() {
		var v validationBuilder
		v.add("level", "phải là một bậc CEFR hợp lệ")
		return v.err()
	}

	current, err := s.repo.ListIDsByLevel(ctx, level)
	if err != nil {
		return fmt.Errorf("reorder courses: %w", err)
	}

	if err := checkSameIDSet(current, courseIDs, "course_ids", "khoá"); err != nil {
		return err
	}

	if _, err := s.repo.ReorderCourses(ctx, level, courseIDs); err != nil {
		return fmt.Errorf("reorder courses: %w", err)
	}
	return nil
}
