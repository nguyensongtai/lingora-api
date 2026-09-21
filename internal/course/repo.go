package course

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/songtai/lingora/api/internal/db"
	"github.com/songtai/lingora/api/internal/platform/postgres"
)

// uniqueViolation là SQLSTATE cho vi phạm unique index.
const uniqueViolation = "23505"

// Repo đọc ghi khoá học trên PostgreSQL thông qua code sqlc sinh ra.
type Repo struct {
	q *db.Queries
}

// NewRepo nhận DBTX nên dùng được cả với pool lẫn transaction.
func NewRepo(dbtx db.DBTX) *Repo {
	return &Repo{q: db.New(dbtx)}
}

// Create thêm một khoá học mới.
func (r *Repo) Create(ctx context.Context, params CreateParams) (Course, error) {
	row, err := r.q.CreateCourse(ctx, db.CreateCourseParams{
		Slug:          params.Slug,
		Title:         params.Title,
		Description:   params.Description,
		Level:         db.CourseLevel(params.Level),
		Status:        db.CourseStatus(params.Status),
		CoverImageUrl: params.CoverImageURL,
	})
	if err != nil {
		if isUniqueViolation(err, "courses_slug_key") {
			return Course{}, fmt.Errorf("create course %q: %w", params.Slug, ErrSlugTaken)
		}
		return Course{}, fmt.Errorf("create course: %w", err)
	}
	return toCourse(row), nil
}

// GetByID đọc một khoá học chưa bị xoá mềm.
func (r *Repo) GetByID(ctx context.Context, id string) (Course, error) {
	courseID, err := parseID(id)
	if err != nil {
		return Course{}, err
	}

	row, err := r.q.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Course{}, fmt.Errorf("get course %s: %w", id, ErrNotFound)
		}
		return Course{}, fmt.Errorf("get course %s: %w", id, err)
	}
	return toCourse(row), nil
}

// GetBySlug đọc một khoá học theo slug.
func (r *Repo) GetBySlug(ctx context.Context, slug string) (Course, error) {
	row, err := r.q.GetCourseBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Course{}, fmt.Errorf("get course by slug %q: %w", slug, ErrNotFound)
		}
		return Course{}, fmt.Errorf("get course by slug %q: %w", slug, err)
	}
	return toCourse(row), nil
}

// List trả về tối đa filter.PageSize khoá học theo keyset (created_at, id) giảm dần.
func (r *Repo) List(ctx context.Context, filter ListFilter) ([]Course, error) {
	params := db.ListCoursesParams{PageSize: filter.PageSize}

	if filter.Status != nil {
		status := db.CourseStatus(*filter.Status)
		params.Status = &status
	}
	if filter.Level != nil {
		level := db.CourseLevel(*filter.Level)
		params.Level = &level
	}
	if filter.Cursor != nil {
		cursorID, err := parseID(filter.Cursor.ID)
		if err != nil {
			return nil, err
		}
		createdAt := filter.Cursor.CreatedAt
		params.CursorCreatedAt = &createdAt
		params.CursorID = cursorID
	}

	rows, err := r.q.ListCourses(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list courses: %w", err)
	}

	courses := make([]Course, 0, len(rows))
	for _, row := range rows {
		courses = append(courses, toCourse(row))
	}
	return courses, nil
}

// Update áp dụng partial update lên một khoá học chưa bị xoá mềm.
func (r *Repo) Update(ctx context.Context, id string, params UpdateParams) (Course, error) {
	courseID, err := parseID(id)
	if err != nil {
		return Course{}, err
	}

	arg := db.UpdateCourseParams{
		ID:              courseID,
		Slug:            params.Slug,
		Title:           params.Title,
		Description:     params.Description,
		CoverImageUrl:   params.CoverImageURL,
		ClearCoverImage: params.ClearCoverImage,
	}
	if params.Level != nil {
		level := db.CourseLevel(*params.Level)
		arg.Level = &level
	}
	if params.Status != nil {
		status := db.CourseStatus(*params.Status)
		arg.Status = &status
	}

	row, err := r.q.UpdateCourse(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Course{}, fmt.Errorf("update course %s: %w", id, ErrNotFound)
		}
		if isUniqueViolation(err, "courses_slug_key") {
			return Course{}, fmt.Errorf("update course %s: %w", id, ErrSlugTaken)
		}
		return Course{}, fmt.Errorf("update course %s: %w", id, err)
	}
	return toCourse(row), nil
}

// SoftDelete đánh dấu deleted_at; khoá đã xoá trước đó trả về ErrNotFound.
func (r *Repo) SoftDelete(ctx context.Context, id string) error {
	courseID, err := parseID(id)
	if err != nil {
		return err
	}

	affected, err := r.q.SoftDeleteCourse(ctx, courseID)
	if err != nil {
		return fmt.Errorf("soft delete course %s: %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("soft delete course %s: %w", id, ErrNotFound)
	}
	return nil
}

// Exists cho biết khoá học có tồn tại và chưa bị xoá mềm hay không.
func (r *Repo) Exists(ctx context.Context, id string) (bool, error) {
	courseID, err := parseID(id)
	if err != nil {
		return false, err
	}

	exists, err := r.q.CourseExists(ctx, courseID)
	if err != nil {
		return false, fmt.Errorf("check course %s exists: %w", id, err)
	}
	return exists, nil
}

// ListLessons trả về bài học của một khoá theo thứ tự position tăng dần.
func (r *Repo) ListLessons(ctx context.Context, courseID string) ([]Lesson, error) {
	id, err := parseID(courseID)
	if err != nil {
		return nil, err
	}

	rows, err := r.q.ListLessonsByCourse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list lessons of course %s: %w", courseID, err)
	}

	lessons := make([]Lesson, 0, len(rows))
	for _, row := range rows {
		lessons = append(lessons, toLesson(row))
	}
	return lessons, nil
}

func parseID(raw string) (pgtype.UUID, error) {
	id, err := postgres.ParseUUID(raw)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("%w: %q", ErrInvalidID, raw)
	}
	return id, nil
}

func isUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == uniqueViolation && pgErr.ConstraintName == constraint
}

func toCourse(row db.Course) Course {
	return Course{
		ID:            postgres.UUIDString(row.ID),
		Slug:          row.Slug,
		Title:         row.Title,
		Description:   row.Description,
		Level:         Level(row.Level),
		Status:        Status(row.Status),
		CoverImageURL: row.CoverImageUrl,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func toLesson(row db.Lesson) Lesson {
	return Lesson{
		ID:        postgres.UUIDString(row.ID),
		CourseID:  postgres.UUIDString(row.CourseID),
		Slug:      row.Slug,
		Title:     row.Title,
		Position:  row.Position,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
