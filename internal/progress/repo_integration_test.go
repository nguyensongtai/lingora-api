package progress_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nguyensongtai/lingora-api/internal/course"
	"github.com/nguyensongtai/lingora-api/internal/progress"
	"github.com/nguyensongtai/lingora-api/internal/testdb"
	"github.com/nguyensongtai/lingora-api/internal/user"
)

// fixture dựng một người học, một khoá và ba bài trong cùng transaction.
type fixture struct {
	tx        pgx.Tx
	progress  *progress.Repo
	courses   *course.Repo
	userID    string
	courseID  string
	lessonIDs []string
}

func newFixture(t *testing.T) fixture {
	t.Helper()

	tx := testdb.New(t)
	ctx := context.Background()

	account, err := user.NewRepo(tx).Create(ctx, user.CreateParams{
		Email:        "hoc-vien-" + t.Name() + "@lingora.vn",
		PasswordHash: ptrString("$2a$12$x"),
		DisplayName:  "Học viên",
		Role:         user.RoleStudent,
	})
	if err != nil {
		t.Fatalf("tạo người học: %v", err)
	}

	courses := course.NewRepo(tx)
	created, err := courses.Create(ctx, course.CreateParams{
		Slug: "khoa-tien-do-" + shortName(t), Title: "Khoá tiến độ",
		Level: course.LevelA1, Status: course.StatusPublished,
	})
	if err != nil {
		t.Fatalf("tạo khoá: %v", err)
	}

	var lessonIDs []string
	for _, slug := range []string{"bai-mot", "bai-hai", "bai-ba"} {
		lesson, err := courses.CreateLesson(ctx, created.ID, course.LessonCreateParams{Slug: slug, Title: slug})
		if err != nil {
			t.Fatalf("tạo bài: %v", err)
		}
		lessonIDs = append(lessonIDs, lesson.ID)
	}

	return fixture{
		tx: tx, progress: progress.NewRepo(tx), courses: courses,
		userID: account.ID, courseID: created.ID, lessonIDs: lessonIDs,
	}
}

func TestRepoCompleteIsIdempotent(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	ctx := context.Background()

	for range 3 {
		if err := f.progress.Complete(ctx, f.userID, f.lessonIDs[0]); err != nil {
			t.Fatalf("Complete() returned error: %v", err)
		}
	}

	// ON CONFLICT DO NOTHING: ba lần bấm chỉ để lại một hàng.
	snapshot, err := f.progress.Snapshot(ctx, f.userID)
	if err != nil {
		t.Fatalf("Snapshot() returned error: %v", err)
	}
	if len(snapshot.CompletedLessonIDs) != 1 {
		t.Errorf("CompletedLessonIDs = %v, want đúng một bài", snapshot.CompletedLessonIDs)
	}
	if got := courseProgress(t, snapshot, f.courseID); got.CompletedCount != 1 || got.LessonCount != 3 {
		t.Errorf("tiến độ khoá = %+v, want 1/3", got)
	}
}

func TestRepoSnapshotDropsSoftDeletedLessons(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	ctx := context.Background()

	for _, id := range f.lessonIDs[:2] {
		if err := f.progress.Complete(ctx, f.userID, id); err != nil {
			t.Fatalf("Complete() returned error: %v", err)
		}
	}
	if err := f.courses.SoftDeleteLesson(ctx, f.lessonIDs[0]); err != nil {
		t.Fatalf("SoftDeleteLesson() returned error: %v", err)
	}

	snapshot, err := f.progress.Snapshot(ctx, f.userID)
	if err != nil {
		t.Fatalf("Snapshot() returned error: %v", err)
	}

	// Bài xoá mềm rơi khỏi cả tử số lẫn mẫu số, dù hàng tiến độ vẫn còn trong bảng.
	if len(snapshot.CompletedLessonIDs) != 1 {
		t.Errorf("CompletedLessonIDs = %v, want đúng một bài còn sống", snapshot.CompletedLessonIDs)
	}
	if got := courseProgress(t, snapshot, f.courseID); got.CompletedCount != 1 || got.LessonCount != 2 {
		t.Errorf("tiến độ khoá = %+v, want 1/2", got)
	}
}

func TestRepoSnapshotKeepsCoursesWithoutLessons(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	ctx := context.Background()

	empty, err := f.courses.Create(ctx, course.CreateParams{
		Slug: "khoa-rong-" + shortName(t), Title: "Khoá rỗng",
		Level: course.LevelA1, Status: course.StatusDraft,
	})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	snapshot, err := f.progress.Snapshot(ctx, f.userID)
	if err != nil {
		t.Fatalf("Snapshot() returned error: %v", err)
	}

	// LEFT JOIN giữ khoá chưa có bài để phía trước còn mẫu số mà vẽ tỉ lệ.
	got := courseProgress(t, snapshot, empty.ID)
	if got.LessonCount != 0 || got.CompletedCount != 0 {
		t.Errorf("khoá rỗng = %+v, want 0/0", got)
	}
}

func TestRepoUncompleteRemovesTheRow(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	ctx := context.Background()

	if err := f.progress.Complete(ctx, f.userID, f.lessonIDs[0]); err != nil {
		t.Fatalf("Complete() returned error: %v", err)
	}

	removed, err := f.progress.Uncomplete(ctx, f.userID, f.lessonIDs[0])
	if err != nil {
		t.Fatalf("Uncomplete() returned error: %v", err)
	}
	if !removed {
		t.Error("Uncomplete() = false, want true ở lần đầu")
	}

	// Gỡ lần hai không có gì để xoá, nhưng cũng không phải lỗi.
	again, err := f.progress.Uncomplete(ctx, f.userID, f.lessonIDs[0])
	if err != nil {
		t.Fatalf("Uncomplete() lần hai returned error: %v", err)
	}
	if again {
		t.Error("Uncomplete() lần hai = true, want false")
	}
}

func TestRepoCompleteRejectsAnUnknownLesson(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	ctx := context.Background()

	// Khoá ngoại là chốt cuối: không có câu SELECT nào đứng trước nên không có
	// khe hở giữa lúc kiểm tra và lúc ghi.
	err := testdb.Attempt(t, f.tx, func(nested pgx.Tx) error {
		return progress.NewRepo(nested).Complete(ctx, f.userID, "01929f01-0000-7000-8000-0000000fffff")
	})
	if !errors.Is(err, progress.ErrLessonNotFound) {
		t.Fatalf("error = %v, want ErrLessonNotFound", err)
	}
}

func TestRepoDailyActivityGroupsByDay(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	ctx := context.Background()

	for _, id := range f.lessonIDs {
		if err := f.progress.Complete(ctx, f.userID, id); err != nil {
			t.Fatalf("Complete() returned error: %v", err)
		}
	}

	days, err := f.progress.DailyActivity(ctx, f.userID)
	if err != nil {
		t.Fatalf("DailyActivity() returned error: %v", err)
	}

	// Cả ba lần hoàn thành nằm trong một transaction nên cùng một mốc now(),
	// tức là cùng một ngày.
	if len(days) != 1 {
		t.Fatalf("DailyActivity() = %d ngày, want 1", len(days))
	}
	if days[0].Completed != 3 {
		t.Errorf("ngày đó có %d bài, want 3", days[0].Completed)
	}
}

func courseProgress(t *testing.T, snapshot progress.Snapshot, courseID string) progress.CourseProgress {
	t.Helper()

	for _, item := range snapshot.Courses {
		if item.CourseID == courseID {
			return item
		}
	}
	t.Fatalf("không thấy khoá %s trong snapshot", courseID)
	return progress.CourseProgress{}
}

func ptrString(value string) *string { return &value }

// shortName cắt tên test thành hậu tố hợp lệ cho slug.
func shortName(t *testing.T) string {
	t.Helper()

	name := make([]rune, 0, len(t.Name()))
	for _, r := range t.Name() {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			name = append(name, r)
		case r >= 'A' && r <= 'Z':
			name = append(name, r+('a'-'A'))
		}
	}
	return string(name)
}
