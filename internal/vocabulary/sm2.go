package vocabulary

import (
	"math"
	"time"
)

// Grade là kết quả một lần ôn. Design chỉ cho hai nút nên chỉ có hai mức, ánh
// xạ sang thang 0–5 của SM-2: nhớ là 5, không nhớ là 2 (dưới 3 nghĩa là trượt).
type Grade string

const (
	GradeForgot     Grade = "forgot"
	GradeRemembered Grade = "remembered"
)

// Valid cho biết giá trị có nằm trong enum hay không.
func (g Grade) Valid() bool {
	return g == GradeForgot || g == GradeRemembered
}

const (
	// initialEaseFactor là E-Factor khởi điểm SM-2 quy định.
	initialEaseFactor = 2.5
	// minEaseFactor là sàn SM-2 quy định; dưới mức này khoảng cách ôn co lại quá nhanh.
	minEaseFactor = 1.3

	firstIntervalDays  int32 = 1
	secondIntervalDays int32 = 6
)

// Schedule tính trạng thái SM-2 tiếp theo. current nil nghĩa là lần ôn đầu tiên.
//
// Bám đúng SM-2 gốc ở hai điểm hay bị làm sai: trượt thì đưa repetitions về 0
// và hẹn lại sau một ngày, nhưng *không* đổi E-Factor; và E-Factor có sàn 1.3.
func Schedule(current *Review, grade Grade, today time.Time) Review {
	ease := initialEaseFactor
	var repetitions int32
	if current != nil {
		ease = current.EaseFactor
		repetitions = current.Repetitions
	}

	if grade == GradeForgot {
		return Review{
			EaseFactor:     ease,
			IntervalDays:   firstIntervalDays,
			Repetitions:    0,
			DueOn:          addDays(today, firstIntervalDays),
			LastReviewedAt: today,
		}
	}

	repetitions++

	var interval int32
	switch {
	case repetitions == 1:
		interval = firstIntervalDays
	case repetitions == 2:
		interval = secondIntervalDays
	default:
		previous := firstIntervalDays
		if current != nil && current.IntervalDays > 0 {
			previous = current.IntervalDays
		}
		interval = int32(math.Round(float64(previous) * ease))
	}

	// q = 5: EF + (0.1 - 0*(0.08 + 0*0.02)) = EF + 0.1
	ease = math.Max(minEaseFactor, ease+0.1)

	return Review{
		EaseFactor:     ease,
		IntervalDays:   interval,
		Repetitions:    repetitions,
		DueOn:          addDays(today, interval),
		LastReviewedAt: today,
	}
}

func addDays(day time.Time, days int32) time.Time {
	return day.AddDate(0, 0, int(days))
}
