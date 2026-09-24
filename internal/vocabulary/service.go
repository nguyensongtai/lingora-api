package vocabulary

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// Giới hạn độ dài, khớp với chỗ hiển thị hẹp nhất trong giao diện.
const (
	maxWordLen    = 100
	maxIPALen     = 100
	maxMeaningLen = 300
	maxExampleLen = 500
)

// repository là những gì Service cần ở tầng lưu trữ. Interface khai báo ở phía
// consumer nên Repo không phải biết đến nó.
type repository interface {
	Create(ctx context.Context, lessonID string, params CreateParams) (Entry, error)
	Get(ctx context.Context, entryID string) (Entry, error)
	ListByLesson(ctx context.Context, lessonID string) ([]Entry, error)
	Update(ctx context.Context, entryID string, params UpdateParams) (Entry, error)
	SoftDelete(ctx context.Context, entryID string) error
	ListForUser(ctx context.Context, userID string) ([]Card, error)
	SaveReview(ctx context.Context, userID, entryID string, review Review) (Review, error)
	Unlocked(ctx context.Context, userID, entryID string) (bool, error)
	CountNewThisWeek(ctx context.Context, userID string) (int64, error)
	CountIntroducedSince(ctx context.Context, userID string, since time.Time) (int64, error)
}

// lessons là cửa duy nhất Service nhìn sang gói course.
type lessons interface {
	LessonExists(ctx context.Context, lessonID string) (bool, error)
}

// Service áp quy tắc nghiệp vụ lên từ vựng và lịch ôn.
type Service struct {
	repo    repository
	lessons lessons

	// now tách ra để test kiểm soát được "hôm nay".
	now func() time.Time
}

// Option điều chỉnh Service lúc dựng.
type Option func(*Service)

// WithClock thay nguồn thời gian; lịch ôn tính bằng ngày nên phải cố định được.
func WithClock(now func() time.Time) Option {
	return func(s *Service) { s.now = now }
}

func NewService(repo repository, lessons lessons, opts ...Option) *Service {
	service := &Service{repo: repo, lessons: lessons, now: time.Now}
	for _, opt := range opts {
		opt(service)
	}
	return service
}

/* ---------- quản lý nội dung ---------- */

// Create thêm một từ vào cuối bài.
func (s *Service) Create(ctx context.Context, lessonID string, params CreateParams) (Entry, error) {
	params = trimCreate(params)

	var v validationBuilder
	validateWord(&v, params.Word)
	validateMeaning(&v, params.Meaning)
	validateOptionalLengths(&v, params.IPA, params.Example, params.ExampleVI)
	if err := v.err(); err != nil {
		return Entry{}, err
	}

	exists, err := s.lessons.LessonExists(ctx, lessonID)
	if err != nil {
		return Entry{}, fmt.Errorf("check lesson %s: %w", lessonID, err)
	}
	if !exists {
		return Entry{}, fmt.Errorf("check lesson %s: %w", lessonID, ErrLessonNotFound)
	}

	created, err := s.repo.Create(ctx, lessonID, params)
	if err != nil {
		return Entry{}, fmt.Errorf("create vocabulary: %w", err)
	}
	return created, nil
}

// ListByLesson trả về từ của một bài.
func (s *Service) ListByLesson(ctx context.Context, lessonID string) ([]Entry, error) {
	entries, err := s.repo.ListByLesson(ctx, lessonID)
	if err != nil {
		return nil, fmt.Errorf("list vocabulary: %w", err)
	}
	return entries, nil
}

// Update áp dụng partial update lên một từ.
func (s *Service) Update(ctx context.Context, entryID string, params UpdateParams) (Entry, error) {
	params = trimUpdate(params)

	var v validationBuilder
	if params.Word != nil {
		validateWord(&v, *params.Word)
	}
	if params.Meaning != nil {
		validateMeaning(&v, *params.Meaning)
	}
	if params.Word == nil && params.IPA == nil && params.Meaning == nil &&
		params.Example == nil && params.ExampleVI == nil {
		v.add("body", "cần ít nhất một field để cập nhật")
	}
	if err := v.err(); err != nil {
		return Entry{}, err
	}

	updated, err := s.repo.Update(ctx, entryID, params)
	if err != nil {
		return Entry{}, fmt.Errorf("update vocabulary: %w", err)
	}
	return updated, nil
}

// Delete xoá mềm một từ.
func (s *Service) Delete(ctx context.Context, entryID string) error {
	if err := s.repo.SoftDelete(ctx, entryID); err != nil {
		return fmt.Errorf("delete vocabulary: %w", err)
	}
	return nil
}

/* ---------- phía người học ---------- */

// List trả về từ đã mở khoá của một người, lọc theo nhóm nếu có. Mỗi thẻ
// mang sẵn State.
func (s *Service) List(ctx context.Context, userID string, state *State) ([]Card, error) {
	cards, err := s.classified(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list vocabulary: %w", err)
	}
	if state == nil {
		return cards, nil
	}

	filtered := make([]Card, 0, len(cards))
	for _, card := range cards {
		if card.State == *state {
			filtered = append(filtered, card)
		}
	}
	return filtered, nil
}

// Stats đếm các ô trên đầu màn Từ vựng.
func (s *Service) Stats(ctx context.Context, userID string) (Stats, error) {
	cards, err := s.classified(ctx, userID)
	if err != nil {
		return Stats{}, fmt.Errorf("read vocabulary stats: %w", err)
	}

	newThisWeek, err := s.repo.CountNewThisWeek(ctx, userID)
	if err != nil {
		return Stats{}, fmt.Errorf("read vocabulary stats: %w", err)
	}

	stats := Stats{Learned: int64(len(cards)), NewThisWeek: newThisWeek}
	for _, card := range cards {
		switch card.State {
		case StateDue:
			stats.DueToday++
		case StateMastered:
			stats.Mastered++
		case StateWaiting:
			stats.Waiting++
		case StateLearning:
		}
	}
	return stats, nil
}

// classified đọc từ đã mở khoá và điền State cho từng thẻ.
//
// Từ đã ôn thì theo lịch SM-2 của nó. Từ chưa ôn lần nào chỉ "đến hạn" khi
// trần từ mới hôm nay còn chỗ; hết chỗ thì "đang chờ". Chỗ đã dùng đếm theo
// số từ được ôn lần đầu từ đầu ngày, nên ôn xong một từ mới thì nó rời nhóm
// chưa ôn và không chiếm thêm chỗ nào — trần là 20 từ mỗi ngày, không phải
// 20 từ mỗi lần mở màn hình.
func (s *Service) classified(ctx context.Context, userID string) ([]Card, error) {
	cards, err := s.repo.ListForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	today := s.today()
	introduced, err := s.repo.CountIntroducedSince(ctx, userID, today)
	if err != nil {
		return nil, err
	}

	// Repo đã xếp từ chưa ôn theo thứ tự học, nên chỗ trống rơi vào đúng
	// những từ học trước.
	budget := max(NewWordsPerDay-introduced, 0)
	for i := range cards {
		switch {
		case cards[i].Review != nil:
			cards[i].State = cards[i].StateOn(today)
		case budget > 0:
			cards[i].State = StateDue
			budget--
		default:
			cards[i].State = StateWaiting
		}
	}
	return cards, nil
}

// Review ghi nhận một lần ôn và trả về lịch mới.
//
// Từ chưa mở khoá bị từ chối: ôn một từ thuộc bài chưa học sẽ làm hỏng thống kê
// và cho phép người dùng tự đẩy từ vào hàng đợi của mình.
func (s *Service) Review(ctx context.Context, userID, entryID string, grade Grade) (Review, error) {
	if !grade.Valid() {
		var v validationBuilder
		v.add("grade", `phải là "remembered" hoặc "forgot"`)
		return Review{}, v.err()
	}

	unlocked, err := s.repo.Unlocked(ctx, userID, entryID)
	if err != nil {
		return Review{}, fmt.Errorf("review vocabulary: %w", err)
	}
	if !unlocked {
		return Review{}, fmt.Errorf("review vocabulary %s: %w", entryID, ErrNotUnlocked)
	}

	cards, err := s.repo.ListForUser(ctx, userID)
	if err != nil {
		return Review{}, fmt.Errorf("review vocabulary: %w", err)
	}

	var current *Review
	for _, card := range cards {
		if card.ID == entryID {
			current = card.Review
			break
		}
	}

	next := Schedule(current, grade, s.today())
	saved, err := s.repo.SaveReview(ctx, userID, entryID, next)
	if err != nil {
		return Review{}, fmt.Errorf("review vocabulary: %w", err)
	}
	return saved, nil
}

// today là hôm nay theo múi giờ dùng để cắt ngày, giống chỗ tính chuỗi ngày học.
func (s *Service) today() time.Time {
	at := s.now().In(ReviewLocation)
	return time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, ReviewLocation)
}

// ReviewLocation cắt ngày theo giờ Việt Nam, cùng quy ước với chuỗi ngày học.
var ReviewLocation = time.FixedZone("ICT", 7*60*60)

/* ---------- validate ---------- */

func trimCreate(params CreateParams) CreateParams {
	params.Word = strings.TrimSpace(params.Word)
	params.IPA = strings.TrimSpace(params.IPA)
	params.Meaning = strings.TrimSpace(params.Meaning)
	params.Example = strings.TrimSpace(params.Example)
	params.ExampleVI = strings.TrimSpace(params.ExampleVI)
	return params
}

func trimUpdate(params UpdateParams) UpdateParams {
	params.Word = trimOptional(params.Word)
	params.IPA = trimOptional(params.IPA)
	params.Meaning = trimOptional(params.Meaning)
	params.Example = trimOptional(params.Example)
	params.ExampleVI = trimOptional(params.ExampleVI)
	return params
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func validateWord(v *validationBuilder, word string) {
	switch {
	case word == "":
		v.add("word", "không được để trống")
	case utf8.RuneCountInString(word) > maxWordLen:
		v.add("word", fmt.Sprintf("tối đa %d ký tự", maxWordLen))
	}
}

func validateMeaning(v *validationBuilder, meaning string) {
	switch {
	case meaning == "":
		v.add("meaning", "không được để trống")
	case utf8.RuneCountInString(meaning) > maxMeaningLen:
		v.add("meaning", fmt.Sprintf("tối đa %d ký tự", maxMeaningLen))
	}
}

func validateOptionalLengths(v *validationBuilder, ipa, example, exampleVI string) {
	if utf8.RuneCountInString(ipa) > maxIPALen {
		v.add("ipa", fmt.Sprintf("tối đa %d ký tự", maxIPALen))
	}
	if utf8.RuneCountInString(example) > maxExampleLen {
		v.add("example", fmt.Sprintf("tối đa %d ký tự", maxExampleLen))
	}
	if utf8.RuneCountInString(exampleVI) > maxExampleLen {
		v.add("example_vi", fmt.Sprintf("tối đa %d ký tự", maxExampleLen))
	}
}
