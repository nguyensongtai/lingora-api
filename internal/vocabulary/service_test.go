package vocabulary_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nguyensongtai/lingora-api/internal/vocabulary"
)

const (
	lessonID = "01929f01-0000-7000-8000-000000000001"
	entryID  = "01929f02-0000-7000-8000-000000000001"
	userID   = "01929f03-0000-7000-8000-000000000001"
)

type fakeRepo struct {
	t *testing.T

	createFn    func(context.Context, string, vocabulary.CreateParams) (vocabulary.Entry, error)
	updateFn    func(context.Context, string, vocabulary.UpdateParams) (vocabulary.Entry, error)
	cards       []vocabulary.Card
	unlocked    bool
	newThisWeek int64

	saved *vocabulary.Review
}

func (f *fakeRepo) Create(ctx context.Context, lesson string, params vocabulary.CreateParams) (vocabulary.Entry, error) {
	if f.createFn == nil {
		f.t.Fatal("Create called unexpectedly")
	}
	return f.createFn(ctx, lesson, params)
}

func (f *fakeRepo) Get(context.Context, string) (vocabulary.Entry, error) {
	return vocabulary.Entry{}, vocabulary.ErrNotFound
}

func (f *fakeRepo) ListByLesson(context.Context, string) ([]vocabulary.Entry, error) {
	return nil, nil
}

func (f *fakeRepo) Update(ctx context.Context, id string, params vocabulary.UpdateParams) (vocabulary.Entry, error) {
	if f.updateFn == nil {
		f.t.Fatal("Update called unexpectedly")
	}
	return f.updateFn(ctx, id, params)
}

func (f *fakeRepo) SoftDelete(context.Context, string) error { return nil }

func (f *fakeRepo) ListForUser(context.Context, string) ([]vocabulary.Card, error) {
	return f.cards, nil
}

func (f *fakeRepo) SaveReview(_ context.Context, _, _ string, review vocabulary.Review) (vocabulary.Review, error) {
	f.saved = &review
	return review, nil
}

func (f *fakeRepo) Unlocked(context.Context, string, string) (bool, error) {
	return f.unlocked, nil
}

func (f *fakeRepo) CountNewThisWeek(context.Context, string) (int64, error) {
	return f.newThisWeek, nil
}

type fakeLessons struct {
	exists bool
	calls  int
}

func (f *fakeLessons) LessonExists(context.Context, string) (bool, error) {
	f.calls++
	return f.exists, nil
}

func service(t *testing.T, repo *fakeRepo, lessons *fakeLessons, today string) *vocabulary.Service {
	t.Helper()

	at := day(t, today).Add(15 * time.Hour)
	return vocabulary.NewService(repo, lessons, vocabulary.WithClock(func() time.Time { return at }))
}

func TestCreateTrimsAndRequiresTheLesson(t *testing.T) {
	t.Parallel()

	t.Run("chuẩn hoá khoảng trắng", func(t *testing.T) {
		t.Parallel()

		var got vocabulary.CreateParams
		repo := &fakeRepo{
			t: t,
			createFn: func(_ context.Context, _ string, params vocabulary.CreateParams) (vocabulary.Entry, error) {
				got = params
				return vocabulary.Entry{ID: entryID}, nil
			},
		}

		if _, err := service(t, repo, &fakeLessons{exists: true}, "2026-09-21").
			Create(context.Background(), lessonID, vocabulary.CreateParams{
				Word:    "  reluctant  ",
				Meaning: "  miễn cưỡng  ",
				IPA:     "  /rɪˈlʌktənt/  ",
			}); err != nil {
			t.Fatalf("Create() returned error: %v", err)
		}

		if got.Word != "reluctant" || got.Meaning != "miễn cưỡng" || got.IPA != "/rɪˈlʌktənt/" {
			t.Errorf("params = %+v, want ba field đều được trim", got)
		}
	})

	t.Run("bài không tồn tại", func(t *testing.T) {
		t.Parallel()

		// createFn nil: bài không có thì không được chạm tới kho.
		_, err := service(t, &fakeRepo{t: t}, &fakeLessons{exists: false}, "2026-09-21").
			Create(context.Background(), lessonID, vocabulary.CreateParams{Word: "a", Meaning: "b"})

		if !errors.Is(err, vocabulary.ErrLessonNotFound) {
			t.Fatalf("Create() error = %v, want ErrLessonNotFound", err)
		}
	})

	t.Run("thiếu từ hoặc nghĩa thì không hỏi tới bài", func(t *testing.T) {
		t.Parallel()

		lessons := &fakeLessons{exists: true}
		_, err := service(t, &fakeRepo{t: t}, lessons, "2026-09-21").
			Create(context.Background(), lessonID, vocabulary.CreateParams{Word: "  ", Meaning: ""})

		var validationErr *vocabulary.ValidationError
		if !errors.As(err, &validationErr) {
			t.Fatalf("Create() error = %v, want *ValidationError", err)
		}
		if _, ok := validationErr.Fields["word"]; !ok {
			t.Errorf("Fields = %v, want có key word", validationErr.Fields)
		}
		if _, ok := validationErr.Fields["meaning"]; !ok {
			t.Errorf("Fields = %v, want có key meaning", validationErr.Fields)
		}
		if lessons.calls != 0 {
			t.Errorf("LessonExists gọi %d lần, want 0", lessons.calls)
		}
	})
}

func TestReviewRefusesALockedEntry(t *testing.T) {
	t.Parallel()

	// Chưa học bài chứa từ thì không được ôn: cho phép sẽ vừa hỏng thống kê vừa
	// để người dùng tự đẩy từ vào hàng đợi của mình.
	repo := &fakeRepo{t: t, unlocked: false}

	_, err := service(t, repo, &fakeLessons{exists: true}, "2026-09-21").
		Review(context.Background(), userID, entryID, vocabulary.GradeRemembered)

	if !errors.Is(err, vocabulary.ErrNotUnlocked) {
		t.Fatalf("Review() error = %v, want ErrNotUnlocked", err)
	}
	if repo.saved != nil {
		t.Error("Review() đã ghi lịch cho một từ chưa mở khoá")
	}
}

func TestReviewRejectsAnUnknownGrade(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{t: t, unlocked: true}

	_, err := service(t, repo, &fakeLessons{exists: true}, "2026-09-21").
		Review(context.Background(), userID, entryID, vocabulary.Grade("nho-mo-ho"))

	var validationErr *vocabulary.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Review() error = %v, want *ValidationError", err)
	}
}

func TestReviewContinuesFromTheExistingSchedule(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{
		t:        t,
		unlocked: true,
		cards: []vocabulary.Card{{
			Entry:  vocabulary.Entry{ID: entryID},
			Review: &vocabulary.Review{EaseFactor: 2.5, IntervalDays: 1, Repetitions: 1},
		}},
	}

	got, err := service(t, repo, &fakeLessons{exists: true}, "2026-09-21").
		Review(context.Background(), userID, entryID, vocabulary.GradeRemembered)
	if err != nil {
		t.Fatalf("Review() returned error: %v", err)
	}

	// Lần ôn thứ hai của SM-2 là 6 ngày, không phải quay về 1.
	if got.IntervalDays != 6 || got.Repetitions != 2 {
		t.Errorf("review = %+v, want 6 ngày / 2 lần", got)
	}
	if got.DueOn.Format(time.DateOnly) != "2026-09-27" {
		t.Errorf("DueOn = %s, want 2026-09-27", got.DueOn.Format(time.DateOnly))
	}
}

func TestStatsCountsEachGroupOnce(t *testing.T) {
	t.Parallel()

	today := day(t, "2026-09-21")
	repo := &fakeRepo{
		t:           t,
		newThisWeek: 3,
		cards: []vocabulary.Card{
			{Entry: vocabulary.Entry{ID: "1"}},
			{Entry: vocabulary.Entry{ID: "2"}, Review: &vocabulary.Review{IntervalDays: 30, DueOn: today}},
			{Entry: vocabulary.Entry{ID: "3"}, Review: &vocabulary.Review{IntervalDays: 6, DueOn: day(t, "2026-09-25")}},
			{Entry: vocabulary.Entry{ID: "4"}, Review: &vocabulary.Review{IntervalDays: 40, DueOn: day(t, "2026-10-20")}},
		},
	}

	got, err := service(t, repo, &fakeLessons{exists: true}, "2026-09-21").
		Stats(context.Background(), userID)
	if err != nil {
		t.Fatalf("Stats() returned error: %v", err)
	}

	// Từ số 2 tới hạn hôm nay nên vào nhóm đến hạn, không phải thành thạo.
	want := vocabulary.Stats{Learned: 4, DueToday: 2, Mastered: 1, NewThisWeek: 3}
	if got != want {
		t.Errorf("Stats() = %+v, want %+v", got, want)
	}
}

func TestListFiltersByState(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{
		t: t,
		cards: []vocabulary.Card{
			{Entry: vocabulary.Entry{ID: "1"}},
			{Entry: vocabulary.Entry{ID: "2"}, Review: &vocabulary.Review{IntervalDays: 6, DueOn: day(t, "2026-09-25")}},
		},
	}

	due := vocabulary.StateDue
	got, err := service(t, repo, &fakeLessons{exists: true}, "2026-09-21").
		List(context.Background(), userID, &due)
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}

	if len(got) != 1 || got[0].ID != "1" {
		t.Errorf("List(due) = %+v, want đúng từ chưa ôn lần nào", got)
	}
}
