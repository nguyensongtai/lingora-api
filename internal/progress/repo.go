package progress

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nguyensongtai/lingora-api/internal/db"
	"github.com/nguyensongtai/lingora-api/internal/platform/postgres"
)

// foreignKeyViolation là SQLSTATE khi lesson_id trỏ tới bài không tồn tại.
const foreignKeyViolation = "23503"

// Repo đọc ghi tiến độ trên PostgreSQL thông qua code sqlc sinh ra.
type Repo struct {
	q *db.Queries
}

// NewRepo nhận DBTX nên dùng được cả với pool lẫn transaction.
func NewRepo(dbtx db.DBTX) *Repo {
	return &Repo{q: db.New(dbtx)}
}

// Complete đánh dấu hoàn thành. Gọi lại lần nữa không đổi gì và cũng không lỗi.
func (r *Repo) Complete(ctx context.Context, userID, lessonID string) error {
	user, lesson, err := parsePair(userID, lessonID)
	if err != nil {
		return err
	}

	err = r.q.CompleteLesson(ctx, db.CompleteLessonParams{
		UserID:   user,
		LessonID: lesson,
	})
	if err != nil {
		// Bài không tồn tại chỉ lộ ra ở khoá ngoại: không có câu SELECT nào
		// đứng trước, nên cũng không có khe hở giữa kiểm tra và ghi.
		if isForeignKeyViolation(err, "lesson_progress_lesson_id_fkey") {
			return fmt.Errorf("complete lesson %s: %w", lessonID, ErrLessonNotFound)
		}
		return fmt.Errorf("complete lesson %s: %w", lessonID, err)
	}
	return nil
}

// Uncomplete gỡ đánh dấu và cho biết có hàng nào bị xoá hay không.
func (r *Repo) Uncomplete(ctx context.Context, userID, lessonID string) (bool, error) {
	user, lesson, err := parsePair(userID, lessonID)
	if err != nil {
		return false, err
	}

	affected, err := r.q.UncompleteLesson(ctx, db.UncompleteLessonParams{
		UserID:   user,
		LessonID: lesson,
	})
	if err != nil {
		return false, fmt.Errorf("uncomplete lesson %s: %w", lessonID, err)
	}
	return affected > 0, nil
}

// Snapshot đọc toàn bộ tiến độ của một người học.
func (r *Repo) Snapshot(ctx context.Context, userID string) (Snapshot, error) {
	user, err := parseID(userID)
	if err != nil {
		return Snapshot{}, err
	}

	rows, err := r.q.ListCourseProgress(ctx, user)
	if err != nil {
		return Snapshot{}, fmt.Errorf("list course progress of %s: %w", userID, err)
	}
	courses := make([]CourseProgress, 0, len(rows))
	for _, row := range rows {
		courses = append(courses, CourseProgress{
			CourseID:       postgres.UUIDString(row.CourseID),
			LessonCount:    row.LessonCount,
			CompletedCount: row.CompletedCount,
		})
	}

	lessonIDs, err := r.q.ListCompletedLessonIDs(ctx, user)
	if err != nil {
		return Snapshot{}, fmt.Errorf("list completed lessons of %s: %w", userID, err)
	}
	completed := make([]string, 0, len(lessonIDs))
	for _, id := range lessonIDs {
		completed = append(completed, postgres.UUIDString(id))
	}

	snapshot := Snapshot{Courses: courses, CompletedLessonIDs: completed}

	latest, err := r.q.LatestCompletedCourse(ctx, user)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		// Chưa học bài nào — không phải lỗi, chỉ là chưa có khoá gần nhất.
	case err != nil:
		return Snapshot{}, fmt.Errorf("latest course of %s: %w", userID, err)
	default:
		id := postgres.UUIDString(latest)
		snapshot.LatestCourseID = &id
	}

	return snapshot, nil
}

// DailyActivity trả về số bài hoàn thành theo từng ngày, mới nhất trước. Việc
// biến nó thành XP, chuỗi ngày hay bảy chấm trong tuần là của service.
func (r *Repo) DailyActivity(ctx context.Context, userID string) ([]DayActivity, error) {
	user, err := parseID(userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.q.ListDailyCompletions(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("list daily completions of %s: %w", userID, err)
	}

	days := make([]DayActivity, 0, len(rows))
	for _, row := range rows {
		if !row.Day.Valid {
			continue
		}
		days = append(days, DayActivity{Day: row.Day.Time, Completed: row.Completions})
	}
	return days, nil
}

// CompletedCount đếm tổng số bài đã hoàn thành, không giới hạn thời gian.
func (r *Repo) CompletedCount(ctx context.Context, userID string) (int64, error) {
	user, err := parseID(userID)
	if err != nil {
		return 0, err
	}

	total, err := r.q.CountCompletedLessons(ctx, user)
	if err != nil {
		return 0, fmt.Errorf("count completed lessons of %s: %w", userID, err)
	}
	return total, nil
}

func parseID(raw string) (pgtype.UUID, error) {
	id, err := postgres.ParseUUID(raw)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("%w: %q", ErrInvalidID, raw)
	}
	return id, nil
}

func parsePair(userID, lessonID string) (pgtype.UUID, pgtype.UUID, error) {
	user, err := parseID(userID)
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, err
	}
	lesson, err := parseID(lessonID)
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, err
	}
	return user, lesson, nil
}

func isForeignKeyViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == foreignKeyViolation && pgErr.ConstraintName == constraint
}
