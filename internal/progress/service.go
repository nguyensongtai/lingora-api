package progress

import (
	"context"
	"fmt"
	"regexp"
)

// uuidPattern chỉ kiểm cú pháp id. Service tự chặn id rác để lỗi trả về là 400
// chứ không phải 404 — giống mọi endpoint khác nhận id trên đường dẫn.
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// repository là những gì Service cần ở tầng lưu trữ. Interface khai báo ở phía
// consumer nên Repo không phải biết đến nó.
type repository interface {
	Complete(ctx context.Context, userID, lessonID string) error
	Uncomplete(ctx context.Context, userID, lessonID string) (bool, error)
	Snapshot(ctx context.Context, userID string) (Snapshot, error)
}

// lessons là cửa duy nhất Service nhìn sang gói course: chỉ cần biết một bài
// còn sống hay không, không cần biết nội dung bài.
type lessons interface {
	LessonExists(ctx context.Context, lessonID string) (bool, error)
}

// Service áp quy tắc nghiệp vụ lên tiến độ học.
type Service struct {
	repo    repository
	lessons lessons
}

func NewService(repo repository, lessons lessons) *Service {
	return &Service{repo: repo, lessons: lessons}
}

// Complete đánh dấu người học đã xong một bài. Gọi lại lần nữa là hợp lệ và
// không đổi mốc hoàn thành — nút "Đã xong" bị bấm hai lần không phải lỗi.
func (s *Service) Complete(ctx context.Context, userID, lessonID string) error {
	if err := s.requireLesson(ctx, lessonID); err != nil {
		return err
	}

	if err := s.repo.Complete(ctx, userID, lessonID); err != nil {
		return fmt.Errorf("complete lesson: %w", err)
	}
	return nil
}

// Uncomplete gỡ đánh dấu. Gỡ một bài chưa từng đánh dấu cũng là hợp lệ: kết
// quả người dùng muốn — bài không còn được tính — đã đúng sẵn.
func (s *Service) Uncomplete(ctx context.Context, userID, lessonID string) error {
	if err := s.requireLesson(ctx, lessonID); err != nil {
		return err
	}

	if _, err := s.repo.Uncomplete(ctx, userID, lessonID); err != nil {
		return fmt.Errorf("uncomplete lesson: %w", err)
	}
	return nil
}

// Snapshot đọc toàn bộ tiến độ của một người học.
func (s *Service) Snapshot(ctx context.Context, userID string) (Snapshot, error) {
	found, err := s.repo.Snapshot(ctx, userID)
	if err != nil {
		return Snapshot{}, fmt.Errorf("read progress: %w", err)
	}
	return found, nil
}

// requireLesson chặn id sai cú pháp và bài đã xoá mềm — khoá ngoại không thấy
// được chuyện xoá mềm vì hàng vẫn còn nguyên trong bảng lessons.
func (s *Service) requireLesson(ctx context.Context, lessonID string) error {
	if !uuidPattern.MatchString(lessonID) {
		return fmt.Errorf("%w: %q", ErrInvalidID, lessonID)
	}

	exists, err := s.lessons.LessonExists(ctx, lessonID)
	if err != nil {
		return fmt.Errorf("check lesson %s: %w", lessonID, err)
	}
	if !exists {
		return fmt.Errorf("check lesson %s: %w", lessonID, ErrLessonNotFound)
	}
	return nil
}
