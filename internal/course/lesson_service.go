package course

import (
	"context"
	"fmt"
	"strings"
)

// CreateLesson thêm một bài vào cuối khoá. Khoá không tồn tại là ErrNotFound,
// không phải lỗi validate: client đang trỏ vào một tài nguyên không có thật.
func (s *Service) CreateLesson(ctx context.Context, courseID string, params LessonCreateParams) (Lesson, error) {
	params.Slug = strings.ToLower(strings.TrimSpace(params.Slug))
	params.Title = strings.TrimSpace(params.Title)

	var v validationBuilder
	validateSlug(&v, params.Slug)
	validateTitle(&v, params.Title)
	if err := v.err(); err != nil {
		return Lesson{}, err
	}

	if err := s.requireCourse(ctx, courseID); err != nil {
		return Lesson{}, err
	}

	created, err := s.repo.CreateLesson(ctx, courseID, params)
	if err != nil {
		return Lesson{}, fmt.Errorf("create lesson: %w", err)
	}
	return created, nil
}

// GetLesson đọc một bài theo id.
func (s *Service) GetLesson(ctx context.Context, lessonID string) (Lesson, error) {
	found, err := s.repo.GetLesson(ctx, lessonID)
	if err != nil {
		return Lesson{}, fmt.Errorf("get lesson: %w", err)
	}
	return found, nil
}

// UpdateLesson áp dụng partial update lên một bài.
func (s *Service) UpdateLesson(ctx context.Context, lessonID string, params LessonUpdateParams) (Lesson, error) {
	params.Slug = trimLowerOptional(params.Slug)
	params.Title = trimOptional(params.Title)

	var v validationBuilder
	if params.Slug != nil {
		validateSlug(&v, *params.Slug)
	}
	if params.Title != nil {
		validateTitle(&v, *params.Title)
	}
	if params.Slug == nil && params.Title == nil {
		v.add("body", "cần ít nhất một field để cập nhật")
	}
	if err := v.err(); err != nil {
		return Lesson{}, err
	}

	updated, err := s.repo.UpdateLesson(ctx, lessonID, params)
	if err != nil {
		return Lesson{}, fmt.Errorf("update lesson: %w", err)
	}
	return updated, nil
}

// DeleteLesson xoá mềm một bài.
func (s *Service) DeleteLesson(ctx context.Context, lessonID string) error {
	if err := s.repo.SoftDeleteLesson(ctx, lessonID); err != nil {
		return fmt.Errorf("delete lesson: %w", err)
	}
	return nil
}

// ReorderLessons đặt lại thứ tự bài trong khoá. Danh sách gửi lên phải là đúng
// tập bài hiện có: thiếu, thừa hay trùng đều bị từ chối, vì một lần sắp xếp
// thiếu bài sẽ để lại những bài mang position cũ lẫn vào giữa.
func (s *Service) ReorderLessons(ctx context.Context, courseID string, lessonIDs []string) error {
	if err := s.requireCourse(ctx, courseID); err != nil {
		return err
	}

	current, err := s.repo.ListLessonIDs(ctx, courseID)
	if err != nil {
		return fmt.Errorf("reorder lessons: %w", err)
	}

	if err := checkSameLessonSet(current, lessonIDs); err != nil {
		return err
	}

	if _, err := s.repo.ReorderLessons(ctx, courseID, lessonIDs); err != nil {
		return fmt.Errorf("reorder lessons: %w", err)
	}
	return nil
}

func (s *Service) requireCourse(ctx context.Context, courseID string) error {
	exists, err := s.repo.Exists(ctx, courseID)
	if err != nil {
		return fmt.Errorf("check course: %w", err)
	}
	if !exists {
		return fmt.Errorf("course %s: %w", courseID, ErrNotFound)
	}
	return nil
}

func checkSameLessonSet(current, requested []string) error {
	return checkSameIDSet(current, requested, "lesson_ids", "bài")
}

// checkSameIDSet đòi danh sách gửi lên đúng bằng tập hiện có. Dùng chung cho cả
// sắp xếp bài trong khoá lẫn sắp xếp khoá trong bậc.
func checkSameIDSet(current, requested []string, field, noun string) error {
	var v validationBuilder

	seen := make(map[string]bool, len(requested))
	for _, id := range requested {
		if seen[id] {
			v.add(field, "có id bị lặp")
			return v.err()
		}
		seen[id] = true
	}

	if len(requested) != len(current) {
		v.add(field, fmt.Sprintf("phải liệt kê đủ %d %s, đang có %d", len(current), noun, len(requested)))
		return v.err()
	}

	for _, id := range current {
		if !seen[id] {
			v.add(field, "thiếu "+noun+" "+id)
			return v.err()
		}
	}
	return nil
}
