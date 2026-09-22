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
)

// Valid cho biết giá trị có phải một dạng câu hỏi hay không.
func (k Kind) Valid() bool {
	switch k {
	case KindMultipleChoice, KindFillBlank, KindListenChoose:
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
	// Hint là gợi ý thêm; hiện chỉ dùng cho KindFillBlank để hiện nghĩa.
	Hint string
	// Options rỗng với KindFillBlank vì dạng đó gõ tay.
	Options []string
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
