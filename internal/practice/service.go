package practice

import (
	"context"
	"fmt"
	"math/rand/v2"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/nguyensongtai/lingora-api/internal/vocabulary"
)

// words là cửa duy nhất gói này nhìn sang vocabulary: lấy vốn từ đã mở khoá,
// và báo lại khi người học trả lời sai.
//
// List đã lọc sẵn theo lesson_progress, nên quy tắc "chỉ ôn từ của bài đã học
// xong" tự động có hiệu lực ở đây mà không phải chép lại.
type words interface {
	List(ctx context.Context, userID string, state *vocabulary.State) ([]vocabulary.Card, error)
	Review(ctx context.Context, userID, entryID string, grade vocabulary.Grade) (vocabulary.Review, error)
}

// Service dựng và chấm phiên luyện tập.
type Service struct {
	words words

	now func() time.Time
	// shuffle tách ra để test cố định được thứ tự; mặc định là ngẫu nhiên thật.
	shuffle func(n int, swap func(i, j int))
}

// Option điều chỉnh Service lúc dựng.
type Option func(*Service)

// WithClock thay nguồn thời gian; "đến hạn" phụ thuộc hôm nay là ngày nào.
func WithClock(now func() time.Time) Option {
	return func(s *Service) { s.now = now }
}

// WithShuffle thay cách xáo trộn. Test cần thứ tự cố định mới kiểm được nội
// dung câu hỏi.
func WithShuffle(shuffle func(n int, swap func(i, j int))) Option {
	return func(s *Service) { s.shuffle = shuffle }
}

func NewService(words words, opts ...Option) *Service {
	service := &Service{words: words, now: time.Now, shuffle: rand.Shuffle}
	for _, opt := range opts {
		opt(service)
	}
	return service
}

// Session dựng một lượt luyện tập từ vốn từ đã mở khoá.
//
// Trả về phiên rỗng chứ không phải lỗi khi người học chưa đủ từ: "chưa có gì
// để luyện" là một trạng thái bình thường của màn hình, không phải sự cố.
func (s *Service) Session(ctx context.Context, userID string, size int) ([]Question, error) {
	cards, err := s.words.List(ctx, userID, nil)
	if err != nil {
		return nil, fmt.Errorf("build practice session: %w", err)
	}

	// Dưới ngưỡng này thì không dựng nổi một câu trắc nghiệm có nghĩa.
	if len(cards) < MinOptions {
		return []Question{}, nil
	}

	ordered := s.prioritise(cards)
	limit := clampSize(size)
	if limit > len(ordered) {
		limit = len(ordered)
	}

	questions := make([]Question, 0, limit)
	for index, card := range ordered[:limit] {
		questions = append(questions, s.buildQuestion(card, cards, index))
	}
	return questions, nil
}

// prioritise xếp từ đến hạn lên trước, rồi đang học, rồi thành thạo — và xáo
// trộn trong từng nhóm. Luyện tập nên hỏi trước những từ sắp quên, nhưng hỏi
// đúng một thứ tự mỗi lần thì người học thuộc thứ tự chứ không thuộc từ.
func (s *Service) prioritise(cards []vocabulary.Card) []vocabulary.Card {
	today := s.today()
	rank := map[vocabulary.State]int{
		vocabulary.StateDue:      0,
		vocabulary.StateLearning: 1,
		vocabulary.StateMastered: 2,
	}

	ordered := make([]vocabulary.Card, len(cards))
	copy(ordered, cards)
	s.shuffle(len(ordered), func(i, j int) { ordered[i], ordered[j] = ordered[j], ordered[i] })

	sort.SliceStable(ordered, func(i, j int) bool {
		return rank[ordered[i].StateOn(today)] < rank[ordered[j].StateOn(today)]
	})
	return ordered
}

// rotation là thứ tự luân phiên các dạng câu hỏi trong một phiên.
//
// Luân phiên theo vị trí chứ không chọn theo dữ liệu của từng từ. Bản đầu tiên
// làm ngược lại — hễ câu ví dụ dùng được thì ra fill_blank — và với giáo trình
// đầy đủ, nơi từ nào cũng có ví dụ tốt, cả mười câu đều thành gõ tay. Dữ liệu
// mỏng lúc đó che mất chuyện này.
var rotation = [...]Kind{KindMultipleChoice, KindFillBlank, KindListenChoose}

// buildQuestion dựng câu hỏi theo dạng tới lượt, và lùi về trắc nghiệm khi dữ
// liệu của từ không đủ cho dạng đó.
func (s *Service) buildQuestion(card vocabulary.Card, pool []vocabulary.Card, index int) Question {
	question := Question{
		EntryID: card.ID,
		Level:   card.Level,
		Word:    card.Word,
	}

	switch rotation[index%len(rotation)] {
	case KindFillBlank:
		// Điền chỗ trống chỉ dựng được khi câu ví dụ thật sự chứa nguyên từ đó
		// — người soạn có thể viết ví dụ ở dạng chia khác, và khoét nhầm thì
		// câu hỏi vô lý.
		if blanked, ok := blankOut(card.Example, card.Word); ok {
			question.Kind = KindFillBlank
			question.Prompt = blanked
			question.Hint = card.Meaning
			return question
		}

	case KindListenChoose:
		question.Kind = KindListenChoose
		question.Options = s.options(card.Word, pool, func(c vocabulary.Card) string { return c.Word })
		return question
	}

	question.Kind = KindMultipleChoice
	question.Prompt = card.Word
	question.Options = s.options(card.Meaning, pool, func(c vocabulary.Card) string { return c.Meaning })
	return question
}

// options dựng danh sách lựa chọn: đáp án đúng cộng thêm đáp án nhiễu lấy từ
// chính vốn từ của người học, rồi xáo trộn.
func (s *Service) options(answer string, pool []vocabulary.Card, pick func(vocabulary.Card) string) []string {
	seen := map[string]bool{normalise(answer): true}
	options := []string{answer}

	// Nhiễu lấy từ vốn từ của chính người học chứ không sinh bừa: chúng trông
	// hợp lý, và không vô tình dạy sai một từ họ chưa gặp.
	order := make([]int, len(pool))
	for i := range order {
		order[i] = i
	}
	s.shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })

	for _, index := range order {
		if len(options) >= MaxOptions {
			break
		}
		candidate := pick(pool[index])
		key := normalise(candidate)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		options = append(options, candidate)
	}

	s.shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })
	return options
}

// Check chấm một câu và, nếu sai, đẩy từ về đầu hàng đợi ôn.
//
// Chỉ phạt khi sai — trả lời đúng KHÔNG kéo dài khoảng cách ôn. Nếu đúng cũng
// tính thì luyện dồn một buổi sẽ đẩy một từ vừa gặp lên thành "thành thạo",
// trong khi SM-2 dựa trên việc nhớ được sau một quãng nghỉ, không phải nhớ
// được năm lần liên tiếp trong mười phút.
func (s *Service) Check(ctx context.Context, userID, entryID string, kind Kind, answer string) (Result, error) {
	if !kind.Valid() {
		return Result{}, fmt.Errorf("%w: %q", ErrInvalidKind, kind)
	}

	card, err := s.find(ctx, userID, entryID)
	if err != nil {
		return Result{}, err
	}

	expected := card.Word
	if kind == KindMultipleChoice {
		expected = card.Meaning
	}

	if normalise(answer) == normalise(expected) {
		return Result{Correct: true, Expected: expected}, nil
	}

	if _, err := s.words.Review(ctx, userID, entryID, vocabulary.GradeForgot); err != nil {
		return Result{}, fmt.Errorf("penalise %s: %w", entryID, err)
	}
	return Result{Correct: false, Expected: expected, Penalised: true}, nil
}

// find lấy một từ trong đúng vốn từ đã mở khoá của người học. Từ chưa mở khoá
// đơn giản là không có trong danh sách, nên quy tắc mở khoá không phải kiểm lại.
func (s *Service) find(ctx context.Context, userID, entryID string) (vocabulary.Card, error) {
	cards, err := s.words.List(ctx, userID, nil)
	if err != nil {
		return vocabulary.Card{}, fmt.Errorf("check answer: %w", err)
	}
	for _, card := range cards {
		if card.ID == entryID {
			return card, nil
		}
	}
	return vocabulary.Card{}, fmt.Errorf("check answer %s: %w", entryID, ErrNotFound)
}

func (s *Service) today() time.Time {
	now := s.now().In(vocabulary.ReviewLocation)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, vocabulary.ReviewLocation)
}

func clampSize(size int) int {
	switch {
	case size <= 0:
		return DefaultSessionSize
	case size > MaxSessionSize:
		return MaxSessionSize
	default:
		return size
	}
}

// normalise bỏ hoa/thường và khoảng trắng thừa trước khi so sánh. Gõ thừa một
// dấu cách hay viết hoa đầu câu không phải là nhớ sai từ.
func normalise(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}

// wordBoundary dựng biểu thức khớp đúng một từ, không khớp khi nó nằm bên
// trong từ khác — "art" không được khoét mất chữ trong "start".
func wordBoundary(word string) (*regexp.Regexp, error) {
	return regexp.Compile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`)
}

// blankOut thay từ trong câu ví dụ bằng chỗ trống. Trả về false khi câu rỗng
// hoặc không chứa từ đó ở dạng nguyên vẹn.
func blankOut(example, word string) (string, bool) {
	if strings.TrimSpace(example) == "" || strings.TrimSpace(word) == "" {
		return "", false
	}

	pattern, err := wordBoundary(word)
	if err != nil {
		return "", false
	}
	if !pattern.MatchString(example) {
		return "", false
	}
	return pattern.ReplaceAllString(example, Blank), true
}
