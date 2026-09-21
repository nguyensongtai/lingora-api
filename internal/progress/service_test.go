package progress_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nguyensongtai/lingora-api/internal/progress"
)

const (
	lessonID = "01929f00-0000-7000-8000-00000000aaaa"
	userID   = "01929f00-0000-7000-8000-00000000bbbb"
)

type fakeRepo struct {
	completeFn   func(ctx context.Context, userID, lessonID string) error
	uncompleteFn func(ctx context.Context, userID, lessonID string) (bool, error)
	snapshotFn   func(ctx context.Context, userID string) (progress.Snapshot, error)
}

func (f *fakeRepo) Complete(ctx context.Context, userID, lessonID string) error {
	if f.completeFn == nil {
		return nil
	}
	return f.completeFn(ctx, userID, lessonID)
}

func (f *fakeRepo) Uncomplete(ctx context.Context, userID, lessonID string) (bool, error) {
	if f.uncompleteFn == nil {
		return true, nil
	}
	return f.uncompleteFn(ctx, userID, lessonID)
}

func (f *fakeRepo) Snapshot(ctx context.Context, userID string) (progress.Snapshot, error) {
	if f.snapshotFn == nil {
		return progress.Snapshot{}, nil
	}
	return f.snapshotFn(ctx, userID)
}

type fakeLessons struct {
	exists bool
	err    error
	calls  int
}

func (f *fakeLessons) LessonExists(context.Context, string) (bool, error) {
	f.calls++
	return f.exists, f.err
}

func aliveLessons() *fakeLessons { return &fakeLessons{exists: true} }

func TestCompleteRequiresALiveLesson(t *testing.T) {
	t.Parallel()

	t.Run("bài đã xoá mềm", func(t *testing.T) {
		t.Parallel()

		repo := &fakeRepo{
			completeFn: func(context.Context, string, string) error {
				t.Fatal("Complete() không được ghi khi bài không còn")
				return nil
			},
		}

		err := progress.NewService(repo, &fakeLessons{exists: false}).
			Complete(context.Background(), userID, lessonID)

		if !errors.Is(err, progress.ErrLessonNotFound) {
			t.Fatalf("Complete() error = %v, want ErrLessonNotFound", err)
		}
	})

	t.Run("id sai cú pháp không chạm tới kho", func(t *testing.T) {
		t.Parallel()

		lessons := aliveLessons()
		err := progress.NewService(&fakeRepo{}, lessons).
			Complete(context.Background(), userID, "không-phải-uuid")

		if !errors.Is(err, progress.ErrInvalidID) {
			t.Fatalf("Complete() error = %v, want ErrInvalidID", err)
		}
		if lessons.calls != 0 {
			t.Errorf("LessonExists gọi %d lần, want 0", lessons.calls)
		}
	})
}

func TestCompleteIsIdempotent(t *testing.T) {
	t.Parallel()

	calls := 0
	repo := &fakeRepo{
		completeFn: func(context.Context, string, string) error {
			calls++
			return nil
		},
	}
	service := progress.NewService(repo, aliveLessons())

	for range 2 {
		if err := service.Complete(context.Background(), userID, lessonID); err != nil {
			t.Fatalf("Complete() returned error: %v", err)
		}
	}

	if calls != 2 {
		t.Errorf("repo.Complete gọi %d lần, want 2", calls)
	}
}

func TestUncompleteAcceptsALessonThatWasNeverCompleted(t *testing.T) {
	t.Parallel()

	// Kết quả người dùng muốn — bài không còn được tính — đã đúng sẵn, nên
	// không có gì để xoá cũng là thành công.
	repo := &fakeRepo{
		uncompleteFn: func(context.Context, string, string) (bool, error) {
			return false, nil
		},
	}

	if err := progress.NewService(repo, aliveLessons()).
		Uncomplete(context.Background(), userID, lessonID); err != nil {
		t.Fatalf("Uncomplete() returned error: %v", err)
	}
}

func TestUncompleteRequiresALiveLesson(t *testing.T) {
	t.Parallel()

	err := progress.NewService(&fakeRepo{}, &fakeLessons{exists: false}).
		Uncomplete(context.Background(), userID, lessonID)

	if !errors.Is(err, progress.ErrLessonNotFound) {
		t.Fatalf("Uncomplete() error = %v, want ErrLessonNotFound", err)
	}
}

func TestSnapshotPassesThroughAndWrapsErrors(t *testing.T) {
	t.Parallel()

	t.Run("trả nguyên dữ liệu của kho", func(t *testing.T) {
		t.Parallel()

		latest := "01929f00-0000-7000-8000-00000000cccc"
		want := progress.Snapshot{
			Courses: []progress.CourseProgress{
				{CourseID: latest, LessonCount: 6, CompletedCount: 3},
			},
			CompletedLessonIDs: []string{lessonID},
			LatestCourseID:     &latest,
		}
		repo := &fakeRepo{
			snapshotFn: func(context.Context, string) (progress.Snapshot, error) {
				return want, nil
			},
		}

		got, err := progress.NewService(repo, aliveLessons()).
			Snapshot(context.Background(), userID)
		if err != nil {
			t.Fatalf("Snapshot() returned error: %v", err)
		}

		if len(got.Courses) != 1 || got.Courses[0].CompletedCount != 3 {
			t.Errorf("Courses = %+v, want một khoá 3/6", got.Courses)
		}
		if got.LatestCourseID == nil || *got.LatestCourseID != latest {
			t.Errorf("LatestCourseID = %v, want %q", got.LatestCourseID, latest)
		}
	})

	t.Run("lỗi kho được bọc lại", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("kết nối rớt")
		repo := &fakeRepo{
			snapshotFn: func(context.Context, string) (progress.Snapshot, error) {
				return progress.Snapshot{}, boom
			},
		}

		_, err := progress.NewService(repo, aliveLessons()).
			Snapshot(context.Background(), userID)

		if !errors.Is(err, boom) {
			t.Fatalf("Snapshot() error = %v, want nó bọc %v", err, boom)
		}
	})
}
