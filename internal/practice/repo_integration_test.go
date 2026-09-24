package practice_test

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nguyensongtai/lingora-api/internal/course"
	"github.com/nguyensongtai/lingora-api/internal/practice"
	"github.com/nguyensongtai/lingora-api/internal/testdb"
	"github.com/nguyensongtai/lingora-api/internal/user"
)

type scoreFixture struct {
	tx       pgx.Tx
	repo     *practice.Repo
	userID   string
	lessonID string
}

func newScoreFixture(t *testing.T) scoreFixture {
	t.Helper()

	tx := testdb.New(t)
	ctx := context.Background()
	name := strings.ToLower(strings.NewReplacer("/", "-", "_", "-").Replace(t.Name()))
	if len(name) > 40 {
		name = name[len(name)-40:]
	}

	hash := "$2a$12$x"
	account, err := user.NewRepo(tx).Create(ctx, user.CreateParams{
		Email: "diem-" + name + "@lingora.vn", PasswordHash: &hash, Role: user.RoleStudent,
	})
	if err != nil {
		t.Fatalf("tạo người học: %v", err)
	}
	courses := course.NewRepo(tx)
	created, err := courses.Create(ctx, course.CreateParams{
		Slug: "khoa-diem-" + strings.Trim(name, "-"), Title: "Khoá", Level: course.LevelA1, Status: course.StatusPublished,
	})
	if err != nil {
		t.Fatalf("tạo khoá: %v", err)
	}
	lesson, err := courses.CreateLesson(ctx, created.ID, course.LessonCreateParams{Slug: "bai-diem", Title: "Bài"})
	if err != nil {
		t.Fatalf("tạo bài: %v", err)
	}
	return scoreFixture{tx: tx, repo: practice.NewRepo(tx), userID: account.ID, lessonID: lesson.ID}
}

func TestRepoRecordScoreKeepsTheBestOfTheSameLesson(t *testing.T) {
	t.Parallel()

	f := newScoreFixture(t)
	ctx := context.Background()

	steps := []struct{ correct, total, wantBest, wantTotal, wantAttempts int }{
		{4, 6, 4, 6, 1},
		{2, 6, 4, 6, 2}, // lượt kém hơn không kéo điểm xuống
		{6, 6, 6, 6, 3},
		// Bài đổi số từ: điểm cũ tính trên bộ từ khác, nên bị thay chứ không so.
		{3, 7, 3, 7, 4},
	}
	for i, step := range steps {
		got, err := f.repo.RecordScore(ctx, f.userID, f.lessonID, step.correct, step.total)
		if err != nil {
			t.Fatalf("lượt %d: RecordScore() returned error: %v", i+1, err)
		}
		if got.BestCorrect != step.wantBest || got.Total != step.wantTotal || got.Attempts != step.wantAttempts {
			t.Errorf("lượt %d = %+v, want %d/%d sau %d lượt", i+1, got, step.wantBest, step.wantTotal, step.wantAttempts)
		}
	}

	list, err := f.repo.ListScores(ctx, f.userID)
	if err != nil {
		t.Fatalf("ListScores() returned error: %v", err)
	}
	if len(list) != 1 || list[0].LessonID != f.lessonID || list[0].BestCorrect != 3 {
		t.Errorf("ListScores() = %+v, want đúng một điểm 3/7", list)
	}
}

// Ràng buộc trong database là lưới cuối nếu service để lọt một kết quả vô lý.
func TestRepoRejectsAnImpossibleScore(t *testing.T) {
	t.Parallel()

	f := newScoreFixture(t)
	if _, err := f.repo.RecordScore(context.Background(), f.userID, f.lessonID, 7, 6); err == nil {
		t.Fatal("RecordScore(7/6) returned nil error, want vi phạm ràng buộc")
	}
}

func TestRepoScoresOfADeletedLessonDisappear(t *testing.T) {
	t.Parallel()

	f := newScoreFixture(t)
	ctx := context.Background()
	if _, err := f.repo.RecordScore(ctx, f.userID, f.lessonID, 5, 6); err != nil {
		t.Fatalf("RecordScore() returned error: %v", err)
	}
	if err := course.NewRepo(f.tx).SoftDeleteLesson(ctx, f.lessonID); err != nil {
		t.Fatalf("SoftDeleteLesson() returned error: %v", err)
	}

	list, err := f.repo.ListScores(ctx, f.userID)
	if err != nil {
		t.Fatalf("ListScores() returned error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("ListScores() = %+v, want rỗng khi bài đã xoá", list)
	}
}
