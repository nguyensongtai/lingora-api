package progress

import (
	"context"
	"fmt"
	"regexp"
	"time"
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
	DailyActivity(ctx context.Context, userID string) ([]DayActivity, error)
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

	// now tách ra để test kiểm soát được "hôm nay".
	now func() time.Time
}

// Option điều chỉnh Service lúc dựng.
type Option func(*Service)

// WithClock thay nguồn thời gian. Chuỗi ngày học phụ thuộc vào "hôm nay là
// ngày nào", nên phải có cách cố định nó lại mà kiểm.
func WithClock(now func() time.Time) Option {
	return func(s *Service) { s.now = now }
}

func NewService(repo repository, lessons lessons, opts ...Option) *Service {
	service := &Service{repo: repo, lessons: lessons, now: time.Now}
	for _, opt := range opts {
		opt(service)
	}
	return service
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

// Snapshot đọc toàn bộ tiến độ của một người học, kèm XP hôm nay, chuỗi ngày
// học liên tiếp và bảy ngày gần nhất.
func (s *Service) Snapshot(ctx context.Context, userID string) (Snapshot, error) {
	found, err := s.repo.Snapshot(ctx, userID)
	if err != nil {
		return Snapshot{}, fmt.Errorf("read progress: %w", err)
	}

	days, err := s.repo.DailyActivity(ctx, userID)
	if err != nil {
		return Snapshot{}, fmt.Errorf("read progress: %w", err)
	}

	today := s.today()
	found.GoalXP = DefaultGoalXP
	found.XPPerLesson = XPPerLesson
	found.TodayXP = completionsOn(days, today) * XPPerLesson
	found.StreakDays = streakEndingAt(days, today)
	found.Week = weekEndingAt(days, today)
	return found, nil
}

// today là hôm nay theo múi giờ dùng để cắt ngày.
func (s *Service) today() time.Time {
	return truncateDay(s.now().In(StreakLocation))
}

func truncateDay(at time.Time) time.Time {
	return time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, StreakLocation)
}

func completionsOn(days []DayActivity, day time.Time) int64 {
	for _, entry := range days {
		if sameDay(entry.Day, day) {
			return entry.Completed
		}
	}
	return 0
}

// streakEndingAt đếm số ngày học liên tiếp. Chuỗi vẫn còn sống khi hôm nay chưa
// học nhưng hôm qua có: người học vẫn còn nguyên ngày hôm nay để giữ nó.
func streakEndingAt(days []DayActivity, today time.Time) int64 {
	active := make(map[string]bool, len(days))
	for _, entry := range days {
		if entry.Completed > 0 {
			active[dayKey(entry.Day)] = true
		}
	}

	cursor := today
	if !active[dayKey(cursor)] {
		cursor = cursor.AddDate(0, 0, -1)
	}

	var streak int64
	for active[dayKey(cursor)] {
		streak++
		cursor = cursor.AddDate(0, 0, -1)
	}
	return streak
}

// weekEndingAt trả về bảy ngày gần nhất theo thứ tự cũ trước, mới sau, kể cả
// ngày không học — phía trước vẽ đủ bảy chấm.
func weekEndingAt(days []DayActivity, today time.Time) []DayActivity {
	counts := make(map[string]int64, len(days))
	for _, entry := range days {
		counts[dayKey(entry.Day)] = entry.Completed
	}

	week := make([]DayActivity, 0, 7)
	for offset := 6; offset >= 0; offset-- {
		day := today.AddDate(0, 0, -offset)
		week = append(week, DayActivity{Day: day, Completed: counts[dayKey(day)]})
	}
	return week
}

func dayKey(at time.Time) string {
	return at.In(StreakLocation).Format(time.DateOnly)
}

func sameDay(a, b time.Time) bool {
	return dayKey(a) == dayKey(b)
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
