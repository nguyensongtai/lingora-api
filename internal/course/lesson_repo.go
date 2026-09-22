package course

import (
	"context"
	"encoding/json"
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
		ID:      id,
		Slug:    params.Slug,
		Title:   params.Title,
		Summary: params.Summary,
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

// LessonExists cho biết bài còn sống hay không. Tách khỏi GetLesson để phía gọi
// chỉ hỏi sự tồn tại không phải kéo theo cả nội dung bài.
func (r *Repo) LessonExists(ctx context.Context, lessonID string) (bool, error) {
	if _, err := r.GetLesson(ctx, lessonID); err != nil {
		if errors.Is(err, ErrLessonNotFound) || errors.Is(err, ErrInvalidID) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListBlocks đọc nội dung của một bài theo đúng thứ tự hiển thị.
func (r *Repo) ListBlocks(ctx context.Context, lessonID string) ([]Block, error) {
	id, err := parseLessonID(lessonID)
	if err != nil {
		return nil, err
	}

	rows, err := r.q.ListLessonBlocks(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list blocks of lesson %s: %w", lessonID, err)
	}

	blocks := make([]Block, 0, len(rows))
	for _, row := range rows {
		blocks = append(blocks, Block{
			ID:      postgres.UUIDString(row.ID),
			Kind:    BlockKind(row.Kind),
			Body:    row.Body,
			TextEN:  row.TextEn,
			TextVI:  row.TextVi,
			Speaker: row.Speaker,
		})
	}
	return blocks, nil
}

// ReplaceBlocks thay toàn bộ nội dung của một bài. Danh sách rỗng là xoá sạch.
func (r *Repo) ReplaceBlocks(ctx context.Context, lessonID string, blocks []Block) error {
	id, err := parseLessonID(lessonID)
	if err != nil {
		return err
	}

	// Query nhận jsonb; khoá ở đây phải khớp tên cột mà câu lệnh đọc ra.
	payload := make([]map[string]string, 0, len(blocks))
	for _, block := range blocks {
		payload = append(payload, map[string]string{
			"kind":    string(block.Kind),
			"body":    block.Body,
			"text_en": block.TextEN,
			"text_vi": block.TextVI,
			"speaker": block.Speaker,
		})
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode blocks of lesson %s: %w", lessonID, err)
	}

	if _, err := r.q.ReplaceLessonBlocks(ctx, db.ReplaceLessonBlocksParams{
		LessonID: id,
		Blocks:   encoded,
	}); err != nil {
		return fmt.Errorf("replace blocks of lesson %s: %w", lessonID, err)
	}
	return nil
}
