// Package practice dựng phiên luyện tập từ chính vốn từ người học đã mở khoá,
// và chấm câu trả lời.
//
// Không có bảng nào của riêng gói này: câu hỏi sinh ra tại chỗ từ
// vocabulary_entries, còn kết quả thì đổ vào đúng lịch SM-2 đang có. Nhờ vậy
// không có nội dung nào phải soạn thêm, và không có trạng thái nào phải giữ
// đồng bộ với tiến độ học.
package practice

// Kind là dạng câu hỏi.
type Kind string

const (
	// KindMultipleChoice: hiện từ tiếng Anh, chọn nghĩa tiếng Việt đúng.
	KindMultipleChoice Kind = "multiple_choice"
	// KindFillBlank: câu ví dụ bị khoét mất chính từ đó, người học gõ lại.
	KindFillBlank Kind = "fill_blank"
	// KindListenChoose: máy đọc từ lên, chọn đúng từ vừa nghe.
	KindListenChoose Kind = "listen_choose"
	// KindListenWrite: máy đọc từ lên, người học gõ lại đúng chính tả.
	KindListenWrite Kind = "listen_write"
	// KindDictation: máy đọc cả câu ví dụ, người học gõ lại cả câu.
	KindDictation Kind = "dictation"
)

// Valid cho biết giá trị có phải một dạng câu hỏi hay không.
func (k Kind) Valid() bool {
	switch k {
	case KindMultipleChoice, KindFillBlank, KindListenChoose, KindListenWrite, KindDictation:
		return true
	default:
		return false
	}
}

// Question là một câu hỏi đã sẵn sàng hiển thị.
type Question struct {
	EntryID string
	Kind    Kind
	Level   string

	// Word là từ tiếng Anh. Với KindListenChoose, phía trước cần nó để máy đọc
	// lên — nghĩa là đáp án nằm sẵn trong trang. Đó là hệ quả không tránh được
	// khi phát âm bằng speechSynthesis của trình duyệt thay vì file audio:
	// muốn đọc thì phải có chữ. Chấp nhận được vì người gian lận chỉ tự hại
	// mình, và XP vẫn chỉ tới từ việc hoàn thành bài chứ không từ luyện tập.
	Word string

	// Prompt là câu dẫn: từ cần dịch, hoặc câu ví dụ đã khoét chỗ trống.
	Prompt string
	// Hint là gợi ý thêm: nghĩa của từ với KindFillBlank, bản dịch của câu với
	// KindDictation.
	Hint string
	// Options rỗng với các dạng gõ tay.
	Options []string

	// Speak là chữ để máy đọc lên với ba dạng nghe; rỗng với dạng khác. Tách
	// khỏi Prompt vì Prompt là thứ được HIỆN ra — với dạng nghe, hiện chữ lên
	// là đưa luôn đáp án.
	Speak string
}

// Typed cho biết dạng câu hỏi có phải gõ tay hay không.
func (k Kind) Typed() bool {
	return k == KindFillBlank || k == KindListenWrite || k == KindDictation
}

// Result là kết quả chấm một câu.
type Result struct {
	Correct bool
	// Expected luôn được trả về, kể cả khi đúng: người học cần thấy đáp án để
	// đối chiếu, nhất là ở dạng gõ tay.
	Expected string
	// Penalised cho biết câu sai này có đẩy từ về đầu hàng đợi ôn hay không.
	Penalised bool
}

// Blank là chuỗi thay cho từ bị khoét trong câu ví dụ.
const Blank = "____"

// Giới hạn kích thước một phiên luyện tập.
const (
	DefaultSessionSize int = 10
	MaxSessionSize     int = 30

	// MinOptions là số lựa chọn tối thiểu để một câu trắc nghiệm còn có nghĩa.
	// Dưới ngưỡng này thì đoán bừa cũng đúng một nửa.
	MinOptions = 3
	// MaxOptions là số lựa chọn mong muốn.
	MaxOptions = 4
)

// Score là điểm tốt nhất của một người ở bước Luyện tập của một bài.
type Score struct {
	LessonID    string
	BestCorrect int
	Total       int
	Attempts    int
}
