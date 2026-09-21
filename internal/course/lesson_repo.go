package course

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nguyensongtai/lingora-api/internal/db"
	"github.com/nguyensongtai/lingora-api/internal/platform/postgres"
)

// CreateLesson thêm một bài vào cuối khoá.
func (r *Repo) CreateLesson(ctx context.Context, courseID string, params LessonCreateParams) (Lesson, error) {
	id, err := parseID(courseID)
	if err != nil {
		return Lesson{}, err
	}

	row, err := r.q.CreateLesson(ctx, db.CreateLessonParams{
		CourseID: id,
		Slug:     params.Slug,
		Title:    params.Title,
	})
	if err != nil {
		if isUniqueViolation(err, "lessons_course_id_slug_key") {
			return Lesson{}, fmt.Errorf("create lesson %q: %w", params.Slug, ErrLessonSlugTaken)
		}
		return Lesson{}, fmt.Errorf("create lesson: %w", err)
	}
	return toLesson(row), nil
}

// GetLesson đọc một bài chưa bị xoá mềm.
func (r *Repo) GetLesson(ctx context.Context, lessonID string) (Lesson, error) {
	id, err := parseLessonID(lessonID)
	if err != nil {
		return Lesson{}, err
	}

	row, err := r.q.GetLessonByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Lesson{}, fmt.Errorf("get lesson %s: %w", lessonID, ErrLessonNotFound)
		}
		return Lesson{}, fmt.Errorf("get lesson %s: %w", lessonID, err)
	}
	return toLesson(row), nil
}

// UpdateLesson áp dụng partial update lên một bài.
func (r *Repo) UpdateLesson(ctx context.Context, lessonID string, params LessonUpdateParams) (Lesson, error) {
	id, err := parseLessonID(lessonID)
	if err != nil {
		return Lesson{}, err
	}

	row, err := r.q.UpdateLesson(ctx, db.UpdateLessonParams{
		ID:    id,
		Slug:  params.Slug,
		Title: params.Title,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Lesson{}, fmt.Errorf("update lesson %s: %w", lessonID, ErrLessonNotFound)
		}
		if isUniqueViolation(err, "lessons_course_id_slug_key") {
			return Lesson{}, fmt.Errorf("update lesson %s: %w", lessonID, ErrLessonSlugTaken)
		}
		return Lesson{}, fmt.Errorf("update lesson %s: %w", lessonID, err)
	}
	return toLesson(row), nil
}

// SoftDeleteLesson đánh dấu deleted_at cho một bài.
func (r *Repo) SoftDeleteLesson(ctx context.Context, lessonID string) error {
	id, err := parseLessonID(lessonID)
	if err != nil {
		return err
	}

	affected, err := r.q.SoftDeleteLesson(ctx, id)
	if err != nil {
		return fmt.Errorf("soft delete lesson %s: %w", lessonID, err)
	}
	if affected == 0 {
		return fmt.Errorf("soft delete lesson %s: %w", lessonID, ErrLessonNotFound)
	}
	return nil
}

// ListLessonIDs trả về id các bài của khoá theo thứ tự hiện tại.
func (r *Repo) ListLessonIDs(ctx context.Context, courseID string) ([]string, error) {
	id, err := parseID(courseID)
	if err != nil {
		return nil, err
	}

	rows, err := r.q.ListLessonIDsByCourse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list lesson ids of course %s: %w", courseID, err)
	}

	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, postgres.UUIDString(row))
	}
	return ids, nil
}

// ReorderLessons đánh lại position theo đúng thứ tự của lessonIDs và trả về số
// hàng đã đổi.
func (r *Repo) ReorderLessons(ctx context.Context, courseID string, lessonIDs []string) (int64, error) {
	id, err := parseID(courseID)
	if err != nil {
		return 0, err
	}

	parsed := make([]pgtype.UUID, 0, len(lessonIDs))
	for _, lessonID := range lessonIDs {
		value, err := parseLessonID(lessonID)
		if err != nil {
			return 0, err
		}
		parsed = append(parsed, value)
	}

	affected, err := r.q.ReorderLessons(ctx, db.ReorderLessonsParams{
		CourseID:  id,
		LessonIds: parsed,
	})
	if err != nil {
		return 0, fmt.Errorf("reorder lessons of course %s: %w", courseID, err)
	}
	return affected, nil
}

func parseLessonID(raw string) (pgtype.UUID, error) {
	id, err := postgres.ParseUUID(raw)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("%w: %q", ErrInvalidID, raw)
	}
	return id, nil
}
