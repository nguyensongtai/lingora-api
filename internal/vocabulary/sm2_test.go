package vocabulary_test

import (
	"testing"
	"time"

	"github.com/nguyensongtai/lingora-api/internal/vocabulary"
)

func day(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.ParseInLocation(time.DateOnly, value, vocabulary.ReviewLocation)
	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}
	return parsed
}

func TestScheduleFollowsTheSM2Ladder(t *testing.T) {
	t.Parallel()

	today := day(t, "2026-09-21")

	// Ba lần nhớ liên tiếp: 1 ngày, 6 ngày, rồi 6 × E-Factor.
	first := vocabulary.Schedule(nil, vocabulary.GradeRemembered, today)
	if first.IntervalDays != 1 || first.Repetitions != 1 {
		t.Fatalf("lần 1 = %+v, want 1 ngày / 1 lần", first)
	}
	if first.DueOn.Format(time.DateOnly) != "2026-09-22" {
		t.Errorf("DueOn = %s, want 2026-09-22", first.DueOn.Format(time.DateOnly))
	}

	second := vocabulary.Schedule(&first, vocabulary.GradeRemembered, day(t, "2026-09-22"))
	if second.IntervalDays != 6 || second.Repetitions != 2 {
		t.Fatalf("lần 2 = %+v, want 6 ngày / 2 lần", second)
	}

	third := vocabulary.Schedule(&second, vocabulary.GradeRemembered, day(t, "2026-09-28"))
	// E-Factor sau hai lần nhớ là 2.5 + 0.1 + 0.1 = 2.7; 6 × 2.7 = 16.2 → 16.
	if third.IntervalDays != 16 {
		t.Errorf("lần 3 = %d ngày, want 16 (6 × 2.7 làm tròn)", third.IntervalDays)
	}
	if third.Repetitions != 3 {
		t.Errorf("Repetitions = %d, want 3", third.Repetitions)
	}
}

func TestScheduleResetsOnFailureWithoutTouchingEase(t *testing.T) {
	t.Parallel()

	today := day(t, "2026-09-21")
	mature := vocabulary.Review{
		EaseFactor:   2.8,
		IntervalDays: 30,
		Repetitions:  6,
		DueOn:        today,
	}

	got := vocabulary.Schedule(&mature, vocabulary.GradeForgot, today)

	// SM-2 gốc: trượt thì đếm lại từ đầu nhưng *không* đổi E-Factor.
	if got.Repetitions != 0 {
		t.Errorf("Repetitions = %d, want 0", got.Repetitions)
	}
	if got.IntervalDays != 1 {
		t.Errorf("IntervalDays = %d, want 1", got.IntervalDays)
	}
	if got.EaseFactor != 2.8 {
		t.Errorf("EaseFactor = %v, want giữ nguyên 2.8", got.EaseFactor)
	}
	if got.DueOn.Format(time.DateOnly) != "2026-09-22" {
		t.Errorf("DueOn = %s, want ngày mai", got.DueOn.Format(time.DateOnly))
	}
}

func TestScheduleKeepsEaseAboveTheFloor(t *testing.T) {
	t.Parallel()

	// Hàng cũ có thể mang E-Factor sát sàn; lần nhớ tiếp theo không được đẩy nó
	// xuống dưới 1.3.
	low := vocabulary.Review{EaseFactor: 1.3, IntervalDays: 4, Repetitions: 3}

	got := vocabulary.Schedule(&low, vocabulary.GradeRemembered, day(t, "2026-09-21"))

	if got.EaseFactor < 1.3 {
		t.Errorf("EaseFactor = %v, want >= 1.3", got.EaseFactor)
	}
	if got.IntervalDays != 5 {
		t.Errorf("IntervalDays = %d, want 5 (4 × 1.3 làm tròn)", got.IntervalDays)
	}
}

func TestCardStateFollowsTheSchedule(t *testing.T) {
	t.Parallel()

	today := day(t, "2026-09-21")

	for name, tc := range map[string]struct {
		review *vocabulary.Review
		want   vocabulary.State
	}{
		"chưa ôn lần nào": {nil, vocabulary.StateDue},
		"tới hạn hôm nay": {
			&vocabulary.Review{IntervalDays: 30, DueOn: today}, vocabulary.StateDue,
		},
		"quá hạn": {
			&vocabulary.Review{IntervalDays: 30, DueOn: day(t, "2026-09-01")}, vocabulary.StateDue,
		},
		"đang học": {
			&vocabulary.Review{IntervalDays: 6, DueOn: day(t, "2026-09-25")}, vocabulary.StateLearning,
		},
		"thành thạo": {
			&vocabulary.Review{IntervalDays: 21, DueOn: day(t, "2026-10-10")}, vocabulary.StateMastered,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			card := vocabulary.Card{Review: tc.review}
			if got := card.StateOn(today); got != tc.want {
				t.Errorf("StateOn() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFamiliarityCapsAtFive(t *testing.T) {
	t.Parallel()

	// Vạch trong design có đúng 5 đốt, nên số lần ôn lớn hơn cũng chỉ sáng 5.
	for repetitions, want := range map[int32]int32{0: 0, 3: 3, 5: 5, 12: 5} {
		card := vocabulary.Card{Review: &vocabulary.Review{Repetitions: repetitions}}
		if got := card.Familiarity(); got != want {
			t.Errorf("Familiarity(%d lần) = %d, want %d", repetitions, got, want)
		}
	}
	if got := (vocabulary.Card{}).Familiarity(); got != 0 {
		t.Errorf("Familiarity() của từ chưa ôn = %d, want 0", got)
	}
}
