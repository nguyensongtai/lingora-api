package progress_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nguyensongtai/lingora-api/internal/progress"
)

// fixedToday cố định "hôm nay" để chuỗi ngày học kiểm được.
func fixedToday(t *testing.T, day string) progress.Option {
	t.Helper()

	parsed, err := time.ParseInLocation(time.DateOnly, day, progress.StreakLocation)
	if err != nil {
		t.Fatalf("parse %q: %v", day, err)
	}
	return progress.WithClock(func() time.Time { return parsed.Add(15 * time.Hour) })
}

// activityOn dựng lịch sử "mỗi ngày trong danh sách hoàn thành n bài".
func activityOn(t *testing.T, days map[string]int64) func(context.Context, string) ([]progress.DayActivity, error) {
	t.Helper()

	entries := make([]progress.DayActivity, 0, len(days))
	for day, count := range days {
		parsed, err := time.ParseInLocation(time.DateOnly, day, progress.StreakLocation)
		if err != nil {
			t.Fatalf("parse %q: %v", day, err)
		}
		entries = append(entries, progress.DayActivity{Day: parsed, Completed: count})
	}
	return func(context.Context, string) ([]progress.DayActivity, error) { return entries, nil }
}

const (
	lessonID = "01929f00-0000-7000-8000-00000000aaaa"
	userID   = "01929f00-0000-7000-8000-00000000bbbb"
)

type fakeRepo struct {
	completeFn   func(ctx context.Context, userID, lessonID string) error
	uncompleteFn func(ctx context.Context, userID, lessonID string) (bool, error)
	snapshotFn   func(ctx context.Context, userID string) (progress.Snapshot, error)
	dailyFn      func(ctx context.Context, userID string) ([]progress.DayActivity, error)
}

func (f *fakeRepo) DailyActivity(ctx context.Context, userID string) ([]progress.DayActivity, error) {
	if f.dailyFn == nil {
		return nil, nil
	}
	return f.dailyFn(ctx, userID)
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

func TestSnapshotDerivesTodayXP(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{
		snapshotFn: func(context.Context, string) (progress.Snapshot, error) {
			return progress.Snapshot{}, nil
		},
		dailyFn: activityOn(t, map[string]int64{
			"2026-09-21": 3,
			"2026-09-20": 5,
		}),
	}

	got, err := progress.NewService(repo, aliveLessons(), fixedToday(t, "2026-09-21")).
		Snapshot(context.Background(), userID)
	if err != nil {
		t.Fatalf("Snapshot() returned error: %v", err)
	}

	if want := 3 * progress.XPPerLesson; got.TodayXP != want {
		t.Errorf("TodayXP = %d, want %d", got.TodayXP, want)
	}
	if got.GoalXP != progress.DefaultGoalXP {
		t.Errorf("GoalXP = %d, want %d", got.GoalXP, progress.DefaultGoalXP)
	}
}

func TestSnapshotCountsTheStreak(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		days map[string]int64
		want int64
	}{
		"ba ngày liên tiếp tính cả hôm nay": {
			map[string]int64{"2026-09-21": 1, "2026-09-20": 2, "2026-09-19": 1}, 3,
		},
		"hôm nay chưa học thì chuỗi vẫn sống": {
			// Người học còn nguyên hôm nay để giữ chuỗi, nên chưa được tính mất.
			map[string]int64{"2026-09-20": 2, "2026-09-19": 1}, 2,
		},
		"đứt từ hôm kia": {
			map[string]int64{"2026-09-19": 5, "2026-09-18": 5}, 0,
		},
		"ngày có hàng nhưng không bài nào": {
			map[string]int64{"2026-09-21": 0, "2026-09-20": 3}, 1,
		},
		"chưa học gì": {map[string]int64{}, 0},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := &fakeRepo{
				snapshotFn: func(context.Context, string) (progress.Snapshot, error) {
					return progress.Snapshot{}, nil
				},
				dailyFn: activityOn(t, tc.days),
			}

			got, err := progress.NewService(repo, aliveLessons(), fixedToday(t, "2026-09-21")).
				Snapshot(context.Background(), userID)
			if err != nil {
				t.Fatalf("Snapshot() returned error: %v", err)
			}

			if got.StreakDays != tc.want {
				t.Errorf("StreakDays = %d, want %d", got.StreakDays, tc.want)
			}
		})
	}
}

func TestSnapshotAlwaysReturnsSevenDays(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{
		snapshotFn: func(context.Context, string) (progress.Snapshot, error) {
			return progress.Snapshot{}, nil
		},
		dailyFn: activityOn(t, map[string]int64{"2026-09-21": 2, "2026-09-17": 1}),
	}

	got, err := progress.NewService(repo, aliveLessons(), fixedToday(t, "2026-09-21")).
		Snapshot(context.Background(), userID)
	if err != nil {
		t.Fatalf("Snapshot() returned error: %v", err)
	}

	if len(got.Week) != 7 {
		t.Fatalf("len(Week) = %d, want 7", len(got.Week))
	}
	// Cũ trước, mới sau: chấm cuối cùng là hôm nay.
	if last := got.Week[6]; last.Day.Format(time.DateOnly) != "2026-09-21" || last.Completed != 2 {
		t.Errorf("Week[6] = %+v, want 2026-09-21 với 2 bài", last)
	}
	if first := got.Week[0]; first.Day.Format(time.DateOnly) != "2026-09-15" || first.Completed != 0 {
		t.Errorf("Week[0] = %+v, want 2026-09-15 với 0 bài", first)
	}
	if mid := got.Week[2]; mid.Completed != 1 {
		t.Errorf("Week[2] = %+v, want ngày 17 có 1 bài", mid)
	}
}
