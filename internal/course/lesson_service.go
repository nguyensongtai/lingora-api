package course

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
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
func (s *Service) GetLesson(ctx context.Context, viewer Viewer, lessonID string) (Lesson, error) {
	found, err := s.repo.GetLesson(ctx, lessonID)
	if err != nil {
		return Lesson{}, fmt.Errorf("get lesson: %w", err)
	}
	// Bài không có status riêng; nó thừa hưởng của khoá. Đọc thẳng bằng id sẽ
	// đi vòng qua khoá nháp nếu không hỏi lại đúng một lần nữa.
	if _, err := s.Get(ctx, viewer, found.CourseID); err != nil {
		return Lesson{}, fmt.Errorf("get lesson %s: %w", lessonID, err)
	}
	return found, nil
}

// UpdateLesson áp dụng partial update lên một bài.
func (s *Service) UpdateLesson(ctx context.Context, lessonID string, params LessonUpdateParams) (Lesson, error) {
	params.Slug = trimLowerOptional(params.Slug)
	params.Title = trimOptional(params.Title)
	params.Summary = trimOptional(params.Summary)

	var v validationBuilder
	if params.Slug != nil {
		validateSlug(&v, *params.Slug)
	}
	if params.Title != nil {
		validateTitle(&v, *params.Title)
	}
	if params.Summary != nil && utf8.RuneCountInString(*params.Summary) > maxSummaryLen {
		v.add("summary", fmt.Sprintf("tối đa %d ký tự", maxSummaryLen))
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

// GetLessonDetail đọc một bài kèm nội dung của nó.
func (s *Service) GetLessonDetail(ctx context.Context, viewer Viewer, lessonID string) (LessonDetail, error) {
	lesson, err := s.GetLesson(ctx, viewer, lessonID)
	if err != nil {
		return LessonDetail{}, err
	}

	blocks, err := s.repo.ListBlocks(ctx, lessonID)
	if err != nil {
		return LessonDetail{}, fmt.Errorf("get lesson detail: %w", err)
	}
	return LessonDetail{Lesson: lesson, Blocks: blocks}, nil
}

// ReplaceBlocks thay toàn bộ nội dung của một bài. Gửi danh sách rỗng là xoá
// sạch nội dung — đó là cách duy nhất để làm việc đó, và nó có chủ đích.
//
// Validate lặp lại đúng ràng buộc của database, không phải vì thừa mà để lỗi
// trả về là 400 kèm tên field, thay vì 500 kèm một câu SQLSTATE.
func (s *Service) ReplaceBlocks(ctx context.Context, lessonID string, blocks []Block) error {
	var v validationBuilder
	if len(blocks) > maxBlocksPerLesson {
		v.add("blocks", fmt.Sprintf("tối đa %d khối mỗi bài", maxBlocksPerLesson))
	}
	for index, block := range blocks {
		validateBlock(&v, index, block)
	}
	if err := v.err(); err != nil {
		return err
	}

	// Bài phải tồn tại: không có bước này thì ghi vào một id bịa ra sẽ âm thầm
	// thành công với 0 dòng.
	if _, err := s.repo.GetLesson(ctx, lessonID); err != nil {
		return fmt.Errorf("replace blocks: %w", err)
	}

	if err := s.repo.ReplaceBlocks(ctx, lessonID, normaliseBlocks(blocks)); err != nil {
		return fmt.Errorf("replace blocks: %w", err)
	}
	return nil
}

func normaliseBlocks(blocks []Block) []Block {
	out := make([]Block, 0, len(blocks))
	for _, block := range blocks {
		out = append(out, Block{
			Kind:    block.Kind,
			Body:    strings.TrimSpace(block.Body),
			TextEN:  strings.TrimSpace(block.TextEN),
			TextVI:  strings.TrimSpace(block.TextVI),
			Speaker: strings.TrimSpace(block.Speaker),
		})
	}
	return out
}

func validateBlock(v *validationBuilder, index int, block Block) {
	field := fmt.Sprintf("blocks[%d]", index)

	if !block.Kind.Valid() {
		v.add(field+".kind", "phải là note, example hoặc dialogue")
		return
	}

	body := strings.TrimSpace(block.Body)
	textEN := strings.TrimSpace(block.TextEN)
	speaker := strings.TrimSpace(block.Speaker)

	// Mỗi dạng bắt buộc đúng phần của nó, và phần thừa phải rỗng — nếu không,
	// dữ liệu lọt qua đây sẽ bị database chặn và biến thành lỗi 500.
	switch block.Kind {
	case BlockNote:
		if body == "" {
			v.add(field+".body", "không được rỗng với khối note")
		}
		if textEN != "" || strings.TrimSpace(block.TextVI) != "" || speaker != "" {
			v.add(field+".body", "khối note chỉ dùng body")
		}
	case BlockExample:
		if textEN == "" {
			v.add(field+".text_en", "không được rỗng với khối example")
		}
		if body != "" || speaker != "" {
			v.add(field+".text_en", "khối example chỉ dùng text_en và text_vi")
		}
	case BlockDialogue:
		if textEN == "" {
			v.add(field+".text_en", "không được rỗng với khối dialogue")
		}
		if speaker == "" {
			v.add(field+".speaker", "không được rỗng với khối dialogue")
		}
		if body != "" {
			v.add(field+".text_en", "khối dialogue không dùng body")
		}
	}

	if utf8.RuneCountInString(body) > maxBlockTextLen ||
		utf8.RuneCountInString(textEN) > maxBlockTextLen ||
		utf8.RuneCountInString(block.TextVI) > maxBlockTextLen {
		v.add(field, fmt.Sprintf("mỗi đoạn tối đa %d ký tự", maxBlockTextLen))
	}
	if utf8.RuneCountInString(speaker) > maxSpeakerLen {
		v.add(field+".speaker", fmt.Sprintf("tối đa %d ký tự", maxSpeakerLen))
	}
}
