package vocabulary_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nguyensongtai/lingora-api/internal/course"
	"github.com/nguyensongtai/lingora-api/internal/progress"
	"github.com/nguyensongtai/lingora-api/internal/testdb"
	"github.com/nguyensongtai/lingora-api/internal/user"
	"github.com/nguyensongtai/lingora-api/internal/vocabulary"
)

type fixture struct {
	tx       pgx.Tx
	vocab    *vocabulary.Repo
	progress *progress.Repo
	userID   string
	// doneLesson người học đã hoàn thành; lockedLesson thì chưa.
	doneLesson   string
	lockedLesson string
}

func newFixture(t *testing.T) fixture {
	t.Helper()

	tx := testdb.New(t)
	ctx := context.Background()

	account, err := user.NewRepo(tx).Create(ctx, user.CreateParams{
		Email:        "tu-vung-" + shortName(t) + "@lingora.vn",
		PasswordHash: ptrString("$2a$12$x"),
		DisplayName:  "Học viên",
		Role:         user.RoleStudent,
	})
	if err != nil {
		t.Fatalf("tạo người học: %v", err)
	}

	courses := course.NewRepo(tx)
	created, err := courses.Create(ctx, course.CreateParams{
		Slug: "khoa-tu-vung-" + shortName(t), Title: "Khoá từ vựng",
		Level: course.LevelB1, Status: course.StatusPublished,
	})
	if err != nil {
		t.Fatalf("tạo khoá: %v", err)
	}

	done, err := courses.CreateLesson(ctx, created.ID, course.LessonCreateParams{Slug: "bai-da-hoc", Title: "Bài đã học"})
	if err != nil {
		t.Fatalf("tạo bài: %v", err)
	}
	locked, err := courses.CreateLesson(ctx, created.ID, course.LessonCreateParams{Slug: "bai-chua-hoc", Title: "Bài chưa học"})
	if err != nil {
		t.Fatalf("tạo bài: %v", err)
	}

	progressRepo := progress.NewRepo(tx)
	if err := progressRepo.Complete(ctx, account.ID, done.ID); err != nil {
		t.Fatalf("Complete() returned error: %v", err)
	}

	return fixture{
		tx: tx, vocab: vocabulary.NewRepo(tx), progress: progressRepo,
		userID: account.ID, doneLesson: done.ID, lockedLesson: locked.ID,
	}
}

func (f fixture) addWord(t *testing.T, lessonID, word string) vocabulary.Entry {
	t.Helper()

	entry, err := f.vocab.Create(context.Background(), lessonID, vocabulary.CreateParams{
		Word: word, Meaning: "nghĩa của " + word,
	})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	return entry
}

func TestRepoListForUserOnlyReturnsUnlockedWords(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	visible := f.addWord(t, f.doneLesson, "reluctant")
	f.addWord(t, f.lockedLesson, "itinerary")

	cards, err := f.vocab.ListForUser(context.Background(), f.userID)
	if err != nil {
		t.Fatalf("ListForUser() returned error: %v", err)
	}

	// JOIN vào lesson_progress: từ của bài chưa học không xuất hiện, và không
	// có bước ghi nào lúc học xong bài.
	if len(cards) != 1 {
		t.Fatalf("ListForUser() = %d thẻ, want 1", len(cards))
	}
	if cards[0].ID != visible.ID {
		t.Errorf("thẻ = %s, want từ của bài đã học", cards[0].Word)
	}
	if cards[0].Review != nil {
		t.Errorf("Review = %+v, want nil khi chưa ôn lần nào", cards[0].Review)
	}
	if cards[0].Level != "B1" {
		t.Errorf("Level = %q, want B1 lấy từ khoá chứa bài", cards[0].Level)
	}
}

func TestRepoUnlockedFollowsProgress(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	ctx := context.Background()
	open := f.addWord(t, f.doneLesson, "commute")
	shut := f.addWord(t, f.lockedLesson, "deadline")

	for _, tc := range []struct {
		name    string
		entryID string
		want    bool
	}{
		{"từ của bài đã học", open.ID, true},
		{"từ của bài chưa học", shut.ID, false},
	} {
		got, err := f.vocab.Unlocked(ctx, f.userID, tc.entryID)
		if err != nil {
			t.Fatalf("%s: Unlocked() returned error: %v", tc.name, err)
		}
		if got != tc.want {
			t.Errorf("%s: Unlocked() = %v, want %v", tc.name, got, tc.want)
		}
	}

	// Học xong bài kia thì từ của nó mở khoá ngay, không cần thao tác nào khác.
	if err := f.progress.Complete(ctx, f.userID, f.lockedLesson); err != nil {
		t.Fatalf("Complete() returned error: %v", err)
	}
	got, err := f.vocab.Unlocked(ctx, f.userID, shut.ID)
	if err != nil {
		t.Fatalf("Unlocked() returned error: %v", err)
	}
	if !got {
		t.Error("sau khi học xong bài, từ vẫn khoá")
	}
}

func TestRepoSaveReviewUpsertsTheSchedule(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	ctx := context.Background()
	entry := f.addWord(t, f.doneLesson, "negotiate")

	today := time.Date(2026, 9, 21, 0, 0, 0, 0, vocabulary.ReviewLocation)
	first, err := f.vocab.SaveReview(ctx, f.userID, entry.ID, vocabulary.Review{
		EaseFactor: 2.6, IntervalDays: 1, Repetitions: 1, DueOn: today.AddDate(0, 0, 1),
	})
	if err != nil {
		t.Fatalf("SaveReview() returned error: %v", err)
	}
	if first.IntervalDays != 1 || first.Repetitions != 1 {
		t.Fatalf("lần đầu = %+v, want 1 ngày / 1 lần", first)
	}

	// Lần hai phải UPDATE chứ không INSERT thêm hàng — khoá chính là (user, entry).
	second, err := f.vocab.SaveReview(ctx, f.userID, entry.ID, vocabulary.Review{
		EaseFactor: 2.7, IntervalDays: 6, Repetitions: 2, DueOn: today.AddDate(0, 0, 6),
	})
	if err != nil {
		t.Fatalf("SaveReview() lần hai returned error: %v", err)
	}
	if second.IntervalDays != 6 || second.Repetitions != 2 {
		t.Errorf("lần hai = %+v, want 6 ngày / 2 lần", second)
	}

	cards, err := f.vocab.ListForUser(ctx, f.userID)
	if err != nil {
		t.Fatalf("ListForUser() returned error: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("ListForUser() = %d thẻ, want 1", len(cards))
	}
	if cards[0].Review == nil || cards[0].Review.Repetitions != 2 {
		t.Errorf("Review = %+v, want lần ôn thứ hai", cards[0].Review)
	}
	if cards[0].Review.DueOn.Format(time.DateOnly) != "2026-09-27" {
		t.Errorf("DueOn = %s, want 2026-09-27", cards[0].Review.DueOn.Format(time.DateOnly))
	}
}

func TestRepoSoftDeletedWordLeavesTheQueue(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	ctx := context.Background()
	entry := f.addWord(t, f.doneLesson, "itinerary")

	if err := f.vocab.SoftDelete(ctx, entry.ID); err != nil {
		t.Fatalf("SoftDelete() returned error: %v", err)
	}

	cards, err := f.vocab.ListForUser(ctx, f.userID)
	if err != nil {
		t.Fatalf("ListForUser() returned error: %v", err)
	}
	if len(cards) != 0 {
		t.Errorf("ListForUser() = %d thẻ, want 0 sau khi xoá mềm", len(cards))
	}

	unlocked, err := f.vocab.Unlocked(ctx, f.userID, entry.ID)
	if err != nil {
		t.Fatalf("Unlocked() returned error: %v", err)
	}
	if unlocked {
		t.Error("từ đã xoá mềm vẫn mở khoá, want đóng")
	}
}

func TestRepoPositionsFollowInsertOrder(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	words := []string{"alpha", "bravo", "charlie"}
	for _, word := range words {
		f.addWord(t, f.doneLesson, word)
	}

	entries, err := f.vocab.ListByLesson(context.Background(), f.doneLesson)
	if err != nil {
		t.Fatalf("ListByLesson() returned error: %v", err)
	}
	if len(entries) != len(words) {
		t.Fatalf("ListByLesson() = %d từ, want %d", len(entries), len(words))
	}
	// Câu INSERT lấy max(position)+1 nên từ mới luôn đứng cuối, không có lỗ.
	for index, entry := range entries {
		if entry.Word != words[index] || entry.Position != int32(index) {
			t.Errorf("vị trí %d là %q (position %d), want %q", index, entry.Word, entry.Position, words[index])
		}
	}
}

func ptrString(value string) *string { return &value }

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

// Trần từ mới lấy những từ chưa ôn đứng đầu danh sách, nên danh sách phải xếp
// theo thứ tự người học đã học chúng — bài xong trước đứng trước — chứ không
// theo lúc người soạn tạo từ.
func TestRepoNewWordsFollowTheOrderLessonsWereCompleted(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	ctx := context.Background()

	// Từ của bài "đã học" được tạo trước, nhưng bài kia lại xong sớm hơn.
	first := f.addWord(t, f.doneLesson, "alpha")
	second := f.addWord(t, f.doneLesson, "beta")
	earlier := f.addWord(t, f.lockedLesson, "gamma")
	if err := f.progress.Complete(ctx, f.userID, f.lockedLesson); err != nil {
		t.Fatalf("Complete() returned error: %v", err)
	}
	if _, err := f.tx.Exec(ctx,
		`UPDATE lesson_progress SET completed_at = now() - interval '1 day'
		 WHERE user_id = $1 AND lesson_id = $2`, f.userID, f.lockedLesson); err != nil {
		t.Fatalf("lùi mốc hoàn thành: %v", err)
	}

	cards, err := f.vocab.ListForUser(ctx, f.userID)
	if err != nil {
		t.Fatalf("ListForUser() returned error: %v", err)
	}

	got := make([]string, 0, len(cards))
	for _, card := range cards {
		got = append(got, card.ID)
	}
	want := []string{earlier.ID, first.ID, second.ID}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Errorf("thứ tự = %v, want gamma (bài xong trước), rồi alpha, beta theo vị trí trong bài", got)
	}
}

func TestRepoCountIntroducedSinceCountsFirstReviews(t *testing.T) {
	t.Parallel()

	f := newFixture(t)
	ctx := context.Background()
	today := time.Date(2026, 9, 24, 0, 0, 0, 0, vocabulary.ReviewLocation)

	for _, word := range []string{"one", "two"} {
		entry := f.addWord(t, f.doneLesson, word)
		if _, err := f.vocab.SaveReview(ctx, f.userID, entry.ID, vocabulary.Review{
			EaseFactor: 2.5, IntervalDays: 1, Repetitions: 1, DueOn: today.AddDate(0, 0, 1),
		}); err != nil {
			t.Fatalf("SaveReview() returned error: %v", err)
		}
	}

	// created_at là now() của transaction, nên một mốc trước đó đếm được cả
	// hai, còn một mốc sau đó thì không đếm được gì.
	var now time.Time
	if err := f.tx.QueryRow(ctx, "SELECT now()").Scan(&now); err != nil {
		t.Fatalf("đọc now(): %v", err)
	}
	for since, want := range map[time.Time]int64{now.Add(-time.Hour): 2, now.Add(time.Hour): 0} {
		got, err := f.vocab.CountIntroducedSince(ctx, f.userID, since)
		if err != nil {
			t.Fatalf("CountIntroducedSince() returned error: %v", err)
		}
		if got != want {
			t.Errorf("CountIntroducedSince(%s) = %d, want %d", since, got, want)
		}
	}
}
