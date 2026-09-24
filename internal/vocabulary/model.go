// Package vocabulary quản lý từ vựng của từng bài và lịch ôn theo SM-2.
package vocabulary

import "time"

// Entry là một từ trong bài học.
type Entry struct {
	ID        string
	LessonID  string
	Word      string
	IPA       string
	Meaning   string
	Example   string
	ExampleVI string
	Position  int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Review là trạng thái SM-2 của một người với một từ.
type Review struct {
	EaseFactor     float64
	IntervalDays   int32
	Repetitions    int32
	DueOn          time.Time
	LastReviewedAt time.Time
}

// Card là một từ kèm lịch ôn của người đang đăng nhập. Review nil nghĩa là
// người đó chưa ôn từ này lần nào — từ mới, và luôn đến hạn.
type Card struct {
	Entry
	// Level là bậc CEFR của khoá chứa bài, tiện cho phía trước khỏi tra thêm.
	Level  string
	Review *Review
	// State do Service.List điền. Nó không suy ra được từ riêng một thẻ: từ
	// chưa ôn lần nào là "đến hạn" hay "đang chờ" tuỳ trần từ mới hôm nay còn
	// bao nhiêu chỗ.
	State State
}

// State là ba nhóm trong design: đến hạn, đang học, thành thạo.
type State string

const (
	StateDue      State = "due"
	StateLearning State = "learning"
	StateMastered State = "mastered"
	// StateWaiting là từ mới đã mở khoá nhưng vượt trần từ mới của hôm nay.
	// Nó không mất: sang ngày sau sẽ lần lượt vào hàng đợi.
	StateWaiting State = "waiting"
)

// NewWordsPerDay là số từ mới tối đa vào hàng đợi ôn mỗi ngày — mặc định của
// Anki, người dùng đã chốt. Không có trần thì học xong ba khoá trong một buổi
// là sáng hôm sau có hơn trăm từ đến hạn cùng lúc.
const NewWordsPerDay int64 = 20

// MasteredIntervalDays là mốc coi một từ đã thành thạo. 21 ngày là ngưỡng quen
// thuộc của các bộ SRS dựa trên SM-2.
const MasteredIntervalDays int32 = 21

// StateOn phân nhóm một từ tại một ngày cho trước.
func (c Card) StateOn(day time.Time) State {
	switch {
	case c.Review == nil:
		return StateDue
	case !c.Review.DueOn.After(day):
		return StateDue
	case c.Review.IntervalDays >= MasteredIntervalDays:
		return StateMastered
	default:
		return StateLearning
	}
}

// Familiarity là số đốt sáng trên vạch 5 đốt của design.
func (c Card) Familiarity() int32 {
	if c.Review == nil {
		return 0
	}
	if c.Review.Repetitions > 5 {
		return 5
	}
	return c.Review.Repetitions
}

// CreateParams là dữ liệu đã hợp lệ để thêm một từ. Position không có ở đây:
// từ mới luôn đứng cuối bài.
type CreateParams struct {
	Word      string
	IPA       string
	Meaning   string
	Example   string
	ExampleVI string
}

// UpdateParams là partial update; nil nghĩa là giữ nguyên.
type UpdateParams struct {
	Word      *string
	IPA       *string
	Meaning   *string
	Example   *string
	ExampleVI *string
}

// Stats là bốn ô thống kê trên đầu màn Từ vựng.
type Stats struct {
	Learned     int64
	DueToday    int64
	Mastered    int64
	NewThisWeek int64
	// Waiting là số từ mới đang chờ tới lượt vào hàng đợi.
	Waiting int64
}
