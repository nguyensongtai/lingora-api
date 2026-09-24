package practice_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nguyensongtai/lingora-api/internal/course"
	"github.com/nguyensongtai/lingora-api/internal/practice"
	"github.com/nguyensongtai/lingora-api/internal/vocabulary"
)

const userID = "01929f00-0000-7000-8000-00000000bbbb"

// noShuffle giữ nguyên thứ tự để nội dung câu hỏi kiểm được. Xáo trộn thật đã
// có test riêng ở dưới.
func noShuffle(int, func(i, j int)) {}

type fakeWords struct {
	cards []vocabulary.Card
	err   error

	// lessonEntries là từ của từng bài, tra theo lesson id. Cố ý tách khỏi
	// cards: luyện trong bài không được phụ thuộc vào việc đã mở khoá.
	lessonEntries map[string][]vocabulary.Entry
	lessonAsked   bool

	// graded ghi lại những lần Service chấm điểm sang SM-2.
	graded []struct {
		entryID string
		grade   vocabulary.Grade
	}
}

func (f *fakeWords) List(context.Context, string, *vocabulary.State) ([]vocabulary.Card, error) {
	return f.cards, f.err
}

func (f *fakeWords) Review(_ context.Context, _, entryID string, grade vocabulary.Grade) (vocabulary.Review, error) {
	f.graded = append(f.graded, struct {
		entryID string
		grade   vocabulary.Grade
	}{entryID, grade})
	return vocabulary.Review{}, nil
}

func (f *fakeWords) ListByLesson(_ context.Context, lessonID string) ([]vocabulary.Entry, error) {
	f.lessonAsked = true
	return f.lessonEntries[lessonID], nil
}

// fakeLessons có đúng một bài, "lesson-1", nằm trong khoá B1 mang status cho
// trước. Nó áp đúng quy tắc của course: khoá nháp thì chỉ admin thấy.
type fakeLessons struct {
	status course.Status
}

func (f fakeLessons) GetLesson(_ context.Context, viewer course.Viewer, lessonID string) (course.Lesson, error) {
	if lessonID != "lesson-1" {
		return course.Lesson{}, course.ErrLessonNotFound
	}
	if !viewer.CanSee(f.status) {
		return course.Lesson{}, course.ErrNotFound
	}
	return course.Lesson{ID: lessonID, CourseID: "course-1"}, nil
}

func (f fakeLessons) Get(_ context.Context, viewer course.Viewer, _ string) (course.Course, error) {
	if !viewer.CanSee(f.status) {
		return course.Course{}, course.ErrNotFound
	}
	return course.Course{ID: "course-1", Level: course.LevelB1, Status: f.status}, nil
}

func card(id, word, meaning, example string) vocabulary.Card {
	return vocabulary.Card{
		Entry: vocabulary.Entry{
			ID:      id,
			Word:    word,
			Meaning: meaning,
			Example: example,
		},
		Level: "A1",
	}
}

// pool là vốn từ đủ để dựng trắc nghiệm; chỉ từ đầu có câu ví dụ dùng được.
func pool() []vocabulary.Card {
	return []vocabulary.Card{
		card("id-1", "reluctant", "miễn cưỡng", "She was reluctant to admit her mistake."),
		card("id-2", "commute", "đi lại (đi làm)", ""),
		card("id-3", "deadline", "hạn chót", ""),
		card("id-4", "thorough", "kỹ lưỡng", ""),
		card("id-5", "spare", "rảnh rỗi", ""),
	}
}

func newService(words *fakeWords) *practice.Service {
	return newLessonService(words, course.StatusPublished)
}

func newLessonService(words *fakeWords, status course.Status) *practice.Service {
	return practice.NewService(words, fakeLessons{status: status},
		practice.WithShuffle(noShuffle),
		practice.WithClock(func() time.Time { return time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC) }),
	)
}

func sessionOf(t *testing.T, words *fakeWords, size int) []practice.Question {
	t.Helper()

	questions, err := newService(words).Session(context.Background(), userID, size)
	if err != nil {
		t.Fatalf("Session() returned error: %v", err)
	}
	return questions
}

/* ---------- dựng phiên ---------- */

// Câu thứ hai trong vòng luân phiên là dạng điền từ.
func TestSessionBlanksTheWordInItsExample(t *testing.T) {
	t.Parallel()

	// "reluctant" là từ duy nhất trong pool có câu ví dụ; đặt nó ở vị trí thứ
	// hai, đúng lượt của dạng điền từ.
	cards := []vocabulary.Card{pool()[1], pool()[0], pool()[2], pool()[3]}
	questions := sessionOf(t, &fakeWords{cards: cards}, 4)

	second := questions[1]
	if second.Kind != practice.KindFillBlank {
		t.Fatalf("kind = %q, want fill_blank ở vị trí thứ hai", second.Kind)
	}
	if second.Prompt != "She was "+practice.Blank+" to admit her mistake." {
		t.Errorf("prompt = %q, chỗ trống đặt sai", second.Prompt)
	}
	if second.Hint != "miễn cưỡng" {
		t.Errorf("hint = %q, want nghĩa tiếng Việt", second.Hint)
	}
	if len(second.Options) != 0 {
		t.Errorf("options = %v, want rỗng: dạng này gõ tay", second.Options)
	}
}

/**
 * Một phiên phải có cả ba dạng chứ không toàn một kiểu.
 *
 * Bản đầu tiên chọn dạng theo dữ liệu của từng từ — hễ câu ví dụ dùng được thì
 * ra fill_blank. Với giáo trình đầy đủ, nơi từ nào cũng có ví dụ tốt, cả mười
 * câu đều thành gõ tay. Test này khoá lại chuyện đó.
 */
func TestSessionRotatesThroughEveryKind(t *testing.T) {
	t.Parallel()

	// Vốn từ mà MỌI từ đều có câu ví dụ dùng được — đúng hình dạng dữ liệu
	// thật, và là đúng trường hợp bản cũ hỏng.
	cards := []vocabulary.Card{
		card("id-1", "reluctant", "miễn cưỡng", "She was reluctant to admit it."),
		card("id-2", "commute", "đi lại", "I commute by bus."),
		card("id-3", "deadline", "hạn chót", "We missed the deadline."),
		card("id-4", "thorough", "kỹ lưỡng", "He is thorough in his work."),
		card("id-5", "spare", "rảnh rỗi", "I read in my spare time."),
		card("id-6", "fresh", "tươi", "We buy fresh fish."),
	}

	questions := sessionOf(t, &fakeWords{cards: cards}, 6)

	seen := map[practice.Kind]int{}
	for _, question := range questions {
		seen[question.Kind]++
	}

	for _, kind := range []practice.Kind{
		practice.KindMultipleChoice,
		practice.KindFillBlank,
		practice.KindListenChoose,
		practice.KindDictation,
		practice.KindListenWrite,
	} {
		if seen[kind] == 0 {
			t.Errorf("không có câu nào dạng %q; phân bố nhận được: %v", kind, seen)
		}
	}
}

// Từ không có câu ví dụ dùng được vẫn phải ra một câu hỏi, chỉ là dạng khác.
func TestSessionFallsBackWhenAKindDoesNotFit(t *testing.T) {
	t.Parallel()

	// Không từ nào có câu ví dụ, nên lượt của fill_blank phải lùi về trắc nghiệm.
	cards := []vocabulary.Card{
		card("id-1", "commute", "đi lại", ""),
		card("id-2", "deadline", "hạn chót", ""),
		card("id-3", "thorough", "kỹ lưỡng", ""),
		card("id-4", "spare", "rảnh rỗi", ""),
	}

	cards = append(cards, card("id-5", "fresh", "tươi", ""))
	questions := sessionOf(t, &fakeWords{cards: cards}, 5)

	if len(questions) != 5 {
		t.Fatalf("len(questions) = %d, want 5", len(questions))
	}
	for i, question := range questions {
		if question.Kind == practice.KindFillBlank || question.Kind == practice.KindDictation {
			t.Errorf("câu %d ra %s dù từ không có câu ví dụ", i, question.Kind)
		}
		if question.Kind == practice.KindMultipleChoice && len(question.Options) == 0 {
			t.Errorf("câu %d là trắc nghiệm nhưng không có lựa chọn nào", i)
		}
	}
}

// "art" không được khoét mất chữ trong "start": khoét nửa từ ra một câu vô lý.
func TestSessionIgnoresAnExampleThatOnlyContainsTheWordInside(t *testing.T) {
	t.Parallel()

	// "art" đặt ở vị trí thứ hai, tức lượt của dạng điền từ.
	words := &fakeWords{cards: []vocabulary.Card{
		card("id-1", "commute", "đi lại", ""),
		card("id-2", "art", "nghệ thuật", "We will start tomorrow."),
		card("id-3", "deadline", "hạn chót", ""),
	}}

	questions := sessionOf(t, words, 2)

	if questions[1].Kind == practice.KindFillBlank {
		t.Errorf("kind = fill_blank, nhưng câu ví dụ không chứa nguyên từ %q", "art")
	}
}

func TestSessionOffersMeaningsForMultipleChoice(t *testing.T) {
	t.Parallel()

	// rotation bắt đầu bằng trắc nghiệm, nên câu đầu luôn là dạng đó.
	words := &fakeWords{cards: pool()[1:]}

	questions := sessionOf(t, words, 1)

	first := questions[0]
	if first.Kind != practice.KindMultipleChoice {
		t.Fatalf("kind = %q, want multiple_choice", first.Kind)
	}
	if first.Prompt != "commute" {
		t.Errorf("prompt = %q, want chính từ đó", first.Prompt)
	}
	if !contains(first.Options, "đi lại (đi làm)") {
		t.Errorf("options = %v, thiếu đáp án đúng", first.Options)
	}
	if len(first.Options) != practice.MaxOptions {
		t.Errorf("len(options) = %d, want %d", len(first.Options), practice.MaxOptions)
	}
}

// Đáp án nhiễu phải lấy từ chính vốn từ của người học: chúng trông hợp lý, và
// không vô tình dạy sai một từ họ chưa gặp.
func TestSessionDrawsDistractorsFromTheLearnersOwnWords(t *testing.T) {
	t.Parallel()

	words := &fakeWords{cards: pool()[1:]}
	meanings := map[string]bool{}
	for _, c := range words.cards {
		meanings[c.Meaning] = true
	}

	questions := sessionOf(t, words, 1)

	for _, option := range questions[0].Options {
		if !meanings[option] {
			t.Errorf("option %q không nằm trong vốn từ của người học", option)
		}
	}
}

func TestSessionNeverRepeatsAnOption(t *testing.T) {
	t.Parallel()

	// Hai từ khác nhau nhưng trùng nghĩa: chỉ một được xuất hiện.
	words := &fakeWords{cards: []vocabulary.Card{
		card("id-1", "commute", "đi lại", ""),
		card("id-2", "travel", "đi lại", ""),
		card("id-3", "deadline", "hạn chót", ""),
		card("id-4", "spare", "rảnh rỗi", ""),
	}}

	questions := sessionOf(t, words, 1)

	seen := map[string]bool{}
	for _, option := range questions[0].Options {
		if seen[option] {
			t.Errorf("option %q xuất hiện hai lần trong %v", option, questions[0].Options)
		}
		seen[option] = true
	}
}

// Chưa đủ từ là trạng thái bình thường của màn hình, không phải sự cố.
func TestSessionIsEmptyWhenThereAreTooFewWords(t *testing.T) {
	t.Parallel()

	words := &fakeWords{cards: pool()[:practice.MinOptions-1]}

	questions := sessionOf(t, words, 10)

	if len(questions) != 0 {
		t.Errorf("len(questions) = %d, want 0", len(questions))
	}
}

func TestSessionClampsItsSize(t *testing.T) {
	t.Parallel()

	cases := map[string]struct{ asked, want int }{
		"không truyền":   {asked: 0, want: 5},
		"số âm":          {asked: -3, want: 5},
		"quá trần":       {asked: 999, want: 5},
		"nhỏ hơn vốn từ": {asked: 2, want: 2},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Vốn từ chỉ có 5, nên mọi yêu cầu lớn hơn đều dừng ở 5.
			questions := sessionOf(t, &fakeWords{cards: pool()}, tc.asked)

			if len(questions) != tc.want {
				t.Errorf("len(questions) = %d, want %d", len(questions), tc.want)
			}
		})
	}
}

// Từ đến hạn nên được hỏi trước: luyện tập là để cứu những từ sắp quên.
func TestSessionAsksDueWordsFirst(t *testing.T) {
	t.Parallel()

	mastered := card("id-mastered", "spare", "rảnh rỗi", "")
	mastered.Review = &vocabulary.Review{
		IntervalDays: vocabulary.MasteredIntervalDays + 5,
		DueOn:        time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC),
	}
	due := card("id-due", "deadline", "hạn chót", "")

	words := &fakeWords{cards: []vocabulary.Card{
		mastered,
		due,
		card("id-3", "commute", "đi lại", ""),
	}}

	questions := sessionOf(t, words, 3)

	if questions[len(questions)-1].EntryID != "id-mastered" {
		t.Errorf("từ đã thành thạo phải nằm cuối, thứ tự nhận được: %v", idsOf(questions))
	}
}

/* ---------- chấm bài ---------- */

func checkAnswer(t *testing.T, words *fakeWords, entryID string, kind practice.Kind, answer string) practice.Result {
	t.Helper()

	result, err := newService(words).Check(context.Background(), userID, entryID, kind, answer)
	if err != nil {
		t.Fatalf("Check() returned error: %v", err)
	}
	return result
}

func TestCheckAcceptsTheRightAnswer(t *testing.T) {
	t.Parallel()

	words := &fakeWords{cards: pool()}

	result := checkAnswer(t, words, "id-1", practice.KindFillBlank, "reluctant")

	if !result.Correct {
		t.Error("Correct = false, want true")
	}
	if result.Expected != "reluctant" {
		t.Errorf("Expected = %q, want reluctant", result.Expected)
	}
	// Đây là quyết định trung tâm: đúng thì KHÔNG đụng vào lịch ôn.
	if len(words.graded) != 0 {
		t.Errorf("trả lời đúng đã chạm SM-2: %v", words.graded)
	}
	if result.Penalised {
		t.Error("Penalised = true cho câu đúng")
	}
}

// Gõ thừa dấu cách hay viết hoa không phải là nhớ sai từ.
func TestCheckIgnoresCaseAndSpacing(t *testing.T) {
	t.Parallel()

	for _, answer := range []string{"Reluctant", "  reluctant  ", "RELUCTANT"} {
		t.Run(answer, func(t *testing.T) {
			t.Parallel()

			words := &fakeWords{cards: pool()}

			if !checkAnswer(t, words, "id-1", practice.KindFillBlank, answer).Correct {
				t.Errorf("%q bị chấm sai", answer)
			}
		})
	}
}

func TestCheckPushesAWrongWordBackIntoTheQueue(t *testing.T) {
	t.Parallel()

	words := &fakeWords{cards: pool()}

	result := checkAnswer(t, words, "id-1", practice.KindFillBlank, "reluctent")

	if result.Correct {
		t.Error("Correct = true cho câu sai")
	}
	if !result.Penalised {
		t.Error("Penalised = false: câu sai phải đẩy từ về đầu hàng đợi")
	}
	if len(words.graded) != 1 {
		t.Fatalf("số lần chấm SM-2 = %d, want 1", len(words.graded))
	}
	if words.graded[0].entryID != "id-1" || words.graded[0].grade != vocabulary.GradeForgot {
		t.Errorf("chấm %v, want id-1 với GradeForgot", words.graded[0])
	}
}

func TestCheckComparesAgainstTheMeaningForMultipleChoice(t *testing.T) {
	t.Parallel()

	words := &fakeWords{cards: pool()}

	if !checkAnswer(t, words, "id-1", practice.KindMultipleChoice, "miễn cưỡng").Correct {
		t.Error("chọn đúng nghĩa lại bị chấm sai")
	}
	// Dạng trắc nghiệm hỏi nghĩa, nên gõ lại chính từ đó là sai.
	if checkAnswer(t, &fakeWords{cards: pool()}, "id-1", practice.KindMultipleChoice, "reluctant").Correct {
		t.Error("đáp án là từ tiếng Anh lại được chấp nhận ở dạng hỏi nghĩa")
	}
}

// Từ chưa mở khoá không có trong danh sách, nên nó không phân biệt được với từ
// không tồn tại — cố ý, vì nói "có từ này nhưng bạn chưa được học" là tiết lộ
// nội dung bài chưa học.
func TestCheckHidesWordsTheLearnerHasNotUnlocked(t *testing.T) {
	t.Parallel()

	_, err := newService(&fakeWords{cards: pool()}).
		Check(context.Background(), userID, "id-khong-co", practice.KindFillBlank, "gì đó")

	if !errors.Is(err, practice.ErrNotFound) {
		t.Fatalf("error = %v, want it to match ErrNotFound", err)
	}
}

func TestCheckRejectsAnUnknownKind(t *testing.T) {
	t.Parallel()

	_, err := newService(&fakeWords{cards: pool()}).
		Check(context.Background(), userID, "id-1", practice.Kind("nghe-roi-doan"), "reluctant")

	if !errors.Is(err, practice.ErrInvalidKind) {
		t.Fatalf("error = %v, want it to match ErrInvalidKind", err)
	}
}

/* ---------- tiện ích ---------- */

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func idsOf(questions []practice.Question) []string {
	ids := make([]string, 0, len(questions))
	for _, question := range questions {
		ids = append(ids, question.EntryID)
	}
	return ids
}

/* ---------- luyện tập trong bài ---------- */

// lessonWords là sáu từ của "lesson-1", mỗi từ có câu ví dụ dùng được.
func lessonWords() map[string][]vocabulary.Entry {
	entries := []vocabulary.Entry{}
	for _, c := range []vocabulary.Card{
		card("l-1", "hello", "xin chào", "Say hello to your new neighbour."),
		card("l-2", "meet", "gặp", "Nice to meet you."),
		card("l-3", "name", "tên", "What is your name?"),
		card("l-4", "spell", "đánh vần", "Could you spell it for me?"),
		card("l-5", "job", "công việc", "I like my job."),
		card("l-6", "live", "sống", "Where do you live?"),
	} {
		entries = append(entries, c.Entry)
	}
	return map[string][]vocabulary.Entry{"lesson-1": entries}
}

func TestLessonSessionAsksEveryWordOfTheLessonEvenBeforeItIsUnlocked(t *testing.T) {
	t.Parallel()

	// cards rỗng: người học chưa mở khoá từ nào, vì chưa bấm "Đã xong".
	words := &fakeWords{lessonEntries: lessonWords()}

	questions, err := newService(words).LessonSession(context.Background(), course.Viewer{}, "lesson-1")
	if err != nil {
		t.Fatalf("LessonSession() returned error: %v", err)
	}
	if len(questions) != 6 {
		t.Fatalf("len = %d, want đủ 6 từ của bài", len(questions))
	}

	kinds := map[practice.Kind]int{}
	for _, question := range questions {
		kinds[question.Kind]++
		if question.Level != string(course.LevelB1) {
			t.Errorf("Level = %q, want bậc của khoá chứa bài", question.Level)
		}
	}
	// Mọi từ đều có ví dụ tốt — đúng tình huống từng làm cả phiên thành gõ tay.
	if len(kinds) != 5 {
		t.Errorf("kinds = %v, want đủ năm dạng", kinds)
	}
}

func TestLessonSessionHidesDraftLessonsFromLearners(t *testing.T) {
	t.Parallel()

	words := &fakeWords{lessonEntries: lessonWords()}

	_, err := newLessonService(words, course.StatusDraft).LessonSession(context.Background(), course.Viewer{}, "lesson-1")
	if !errors.Is(err, course.ErrNotFound) {
		t.Fatalf("error = %v, want course.ErrNotFound", err)
	}
	if words.lessonAsked {
		t.Error("đã đọc từ của một bài mà người này không được thấy")
	}

	// Admin thì thấy, để còn thử bài trước khi xuất bản.
	if _, err := newLessonService(words, course.StatusDraft).
		LessonSession(context.Background(), course.Viewer{IsAdmin: true}, "lesson-1"); err != nil {
		t.Errorf("admin: error = %v, want nil", err)
	}
}

func TestLessonSessionIsEmptyWhenTheLessonHasTooFewWords(t *testing.T) {
	t.Parallel()

	words := &fakeWords{lessonEntries: map[string][]vocabulary.Entry{
		"lesson-1": {{ID: "l-1", Word: "hello", Meaning: "xin chào"}, {ID: "l-2", Word: "meet", Meaning: "gặp"}},
	}}

	questions, err := newService(words).LessonSession(context.Background(), course.Viewer{}, "lesson-1")
	if err != nil {
		t.Fatalf("LessonSession() returned error: %v", err)
	}
	if len(questions) != 0 {
		t.Errorf("len = %d, want 0", len(questions))
	}
}

func TestCheckLessonNeverTouchesTheReviewQueue(t *testing.T) {
	t.Parallel()

	words := &fakeWords{lessonEntries: lessonWords()}
	service := newService(words)

	wrong, err := service.CheckLesson(context.Background(), course.Viewer{}, "lesson-1", "l-1", practice.KindFillBlank, "goodbye")
	if err != nil {
		t.Fatalf("CheckLesson() returned error: %v", err)
	}
	if wrong.Correct || wrong.Penalised || wrong.Expected != "hello" {
		t.Errorf("sai: result = %+v, want Correct=false, Penalised=false, Expected=hello", wrong)
	}

	right, err := service.CheckLesson(context.Background(), course.Viewer{}, "lesson-1", "l-2", practice.KindMultipleChoice, "gặp")
	if err != nil {
		t.Fatalf("CheckLesson() returned error: %v", err)
	}
	if !right.Correct {
		t.Errorf("đúng: result = %+v, want Correct=true", right)
	}

	if len(words.graded) != 0 {
		t.Errorf("graded = %v, want không chấm gì sang SM-2", words.graded)
	}
}

func TestCheckLessonOnlyAcceptsWordsOfThatLesson(t *testing.T) {
	t.Parallel()

	// id-1 là từ đã mở khoá, nhưng không thuộc lesson-1.
	words := &fakeWords{cards: pool(), lessonEntries: lessonWords()}

	_, err := newService(words).CheckLesson(context.Background(), course.Viewer{}, "lesson-1", "id-1", practice.KindMultipleChoice, "miễn cưỡng")
	if !errors.Is(err, practice.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestCheckLessonRejectsAnUnknownKindBeforeReadingAnything(t *testing.T) {
	t.Parallel()

	words := &fakeWords{lessonEntries: lessonWords()}

	_, err := newService(words).CheckLesson(context.Background(), course.Viewer{}, "lesson-1", "l-1", practice.Kind("essay"), "x")
	if !errors.Is(err, practice.ErrInvalidKind) {
		t.Fatalf("error = %v, want ErrInvalidKind", err)
	}
	if words.lessonAsked {
		t.Error("đã đọc từ vựng dù dạng câu hỏi không hợp lệ")
	}
}

/* ---------- nghe rồi viết ---------- */

// lượt thứ tư và thứ năm của vòng luân phiên là chép chính tả và nghe-viết.
func listeningCards() []vocabulary.Card {
	cards := []vocabulary.Card{
		card("id-1", "reluctant", "miễn cưỡng", "She was reluctant to admit it."),
		card("id-2", "commute", "đi lại", "I commute by bus."),
		card("id-3", "deadline", "hạn chót", "We missed the deadline."),
		card("id-4", "umbrella", "cái ô", "Take an umbrella with you."),
		card("id-5", "spare", "rảnh rỗi", "I read in my spare time."),
	}
	cards[3].ExampleVI = "Mang theo ô nhé."
	return cards
}

func TestSessionAsksToWriteWhatIsHeard(t *testing.T) {
	t.Parallel()

	questions := sessionOf(t, &fakeWords{cards: listeningCards()}, 5)

	dictation := questions[3]
	if dictation.Kind != practice.KindDictation {
		t.Fatalf("kind = %q, want dictation ở vị trí thứ tư", dictation.Kind)
	}
	if dictation.Speak != "Take an umbrella with you." || dictation.Hint != "Mang theo ô nhé." {
		t.Errorf("speak = %q, hint = %q, want cả câu và bản dịch", dictation.Speak, dictation.Hint)
	}

	write := questions[4]
	if write.Kind != practice.KindListenWrite || write.Speak != "spare" {
		t.Errorf("câu thứ năm = %+v, want listen_write đọc \"spare\"", write)
	}

	// Với dạng nghe, chữ chỉ được nằm ở Speak: hiện Prompt hay Options ra là
	// đưa luôn đáp án cho một câu chính tả.
	for _, question := range []practice.Question{dictation, write} {
		if question.Prompt != "" || len(question.Options) != 0 {
			t.Errorf("%s: prompt = %q, options = %v, want rỗng", question.Kind, question.Prompt, question.Options)
		}
	}
}

func TestDictationIgnoresCaseAndPunctuationButNotSpelling(t *testing.T) {
	t.Parallel()

	words := &fakeWords{cards: listeningCards()}
	service := newService(words)

	for answer, want := range map[string]bool{
		"take an umbrella with you":     true,
		"Take an umbrella, with you!!":  true,
		"  take   an umbrella with you": true,
		"Take an umbrela with you.":     false,
		"Take the umbrella with you.":   false,
	} {
		result, err := service.Check(context.Background(), userID, "id-4", practice.KindDictation, answer)
		if err != nil {
			t.Fatalf("Check(%q) returned error: %v", answer, err)
		}
		if result.Correct != want {
			t.Errorf("Check(%q).Correct = %v, want %v", answer, result.Correct, want)
		}
		if result.Expected != "Take an umbrella with you." {
			t.Errorf("Expected = %q, want nguyên câu gốc", result.Expected)
		}
	}
}

// Dấu nháy đơn là chính tả, không phải dấu câu: "Nams" không phải "Nam's".
func TestDictationKeepsApostrophes(t *testing.T) {
	t.Parallel()

	cards := listeningCards()
	cards[0].Example = "Did you go to Nam's party?"
	service := newService(&fakeWords{cards: cards})

	right, _ := service.Check(context.Background(), userID, "id-1", practice.KindDictation, "did you go to Nam’s party")
	wrong, _ := service.Check(context.Background(), userID, "id-1", practice.KindDictation, "did you go to Nams party")
	if !right.Correct || wrong.Correct {
		t.Errorf("nháy cong: %v, bỏ nháy: %v — want true, false", right.Correct, wrong.Correct)
	}
}

// Từ không có câu ví dụ thì không có câu chép chính tả nào để chấm, và càng
// không được phạt vào lịch ôn vì nó.
func TestDictationOfAWordWithoutAnExampleIsRejectedNotPenalised(t *testing.T) {
	t.Parallel()

	words := &fakeWords{cards: pool()} // id-2 không có câu ví dụ

	_, err := newService(words).Check(context.Background(), userID, "id-2", practice.KindDictation, "anything")
	if !errors.Is(err, practice.ErrInvalidKind) {
		t.Fatalf("error = %v, want ErrInvalidKind", err)
	}
	if len(words.graded) != 0 {
		t.Errorf("graded = %v, want không phạt gì", words.graded)
	}
}

func TestListenWriteIsGradedAgainstTheWord(t *testing.T) {
	t.Parallel()

	service := newService(&fakeWords{cards: listeningCards()})

	result, err := service.Check(context.Background(), userID, "id-5", practice.KindListenWrite, " Spare ")
	if err != nil {
		t.Fatalf("Check() returned error: %v", err)
	}
	if !result.Correct || result.Expected != "spare" {
		t.Errorf("result = %+v, want đúng, Expected=spare", result)
	}
}
