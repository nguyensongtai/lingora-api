package practice

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/nguyensongtai/lingora-api/internal/api"
	"github.com/nguyensongtai/lingora-api/internal/auth"
	"github.com/nguyensongtai/lingora-api/internal/course"
	"github.com/nguyensongtai/lingora-api/internal/httpx"
	"github.com/nguyensongtai/lingora-api/internal/vocabulary"
)

// service là những gì Handler cần ở tầng nghiệp vụ.
type service interface {
	Session(ctx context.Context, userID string, size int) ([]Question, error)
	Check(ctx context.Context, userID, entryID string, kind Kind, answer string) (Result, error)
	LessonSession(ctx context.Context, viewer course.Viewer, lessonID string) ([]Question, error)
	CheckLesson(ctx context.Context, viewer course.Viewer, lessonID, entryID string, kind Kind, answer string) (Result, error)
	RecordLessonResult(ctx context.Context, viewer course.Viewer, userID, lessonID string, correct, total int) (Score, error)
	LessonScores(ctx context.Context, userID string) ([]Score, error)
}

// Handler ánh xạ HTTP sang nghiệp vụ luyện tập.
type Handler struct {
	service service
}

func NewHandler(svc service) *Handler {
	return &Handler{service: svc}
}

// Mount gắn nhánh /me/practice; toàn bộ là thứ của riêng người đang đăng nhập.
func (h *Handler) Mount(r chi.Router, authenticated func(http.Handler) http.Handler) {
	r.Route("/me/practice", func(r chi.Router) {
		r.Use(authenticated)
		r.Get("/session", h.session)
		r.Post("/answers", h.check)
		r.Get("/scores", h.scores)
		r.Put("/lessons/{lessonID}/score", h.recordScore)
	})
}

func (h *Handler) session(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}

	var questions []Question
	var err error
	if lessonID := strings.TrimSpace(r.URL.Query().Get("lesson_id")); lessonID != "" {
		questions, err = h.service.LessonSession(r.Context(), course.ViewerFrom(r), lessonID)
	} else {
		// Như days của lịch sử: size là tuỳ chọn hiển thị, giá trị hỏng đi
		// tiếp với 0 và service kẹp về mặc định.
		size, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("size")))
		questions, err = h.service.Session(r.Context(), userID, size)
	}
	if err != nil {
		writeError(w, r, err)
		return
	}

	items := make([]api.PracticeQuestion, 0, len(questions))
	for _, question := range questions {
		// Options nil phải thành [] vì spec khai báo mảng và phía trước lặp
		// thẳng trên nó — fill_blank luôn rơi vào trường hợp này.
		options := question.Options
		if options == nil {
			options = []string{}
		}
		items = append(items, api.PracticeQuestion{
			EntryId: question.EntryID,
			Kind:    api.PracticeKind(question.Kind),
			Level:   api.CourseLevel(question.Level),
			Word:    question.Word,
			Prompt:  question.Prompt,
			Hint:    question.Hint,
			Options: options,
			Speak:   question.Speak,
		})
	}

	writeJSON(w, r, http.StatusOK, api.PracticeSession{Questions: items})
}

func (h *Handler) check(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}

	var body api.PracticeAnswer
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeMalformed, "Body không phải JSON hợp lệ.", nil)
		return
	}

	var result Result
	var err error
	if body.LessonId != nil && strings.TrimSpace(*body.LessonId) != "" {
		result, err = h.service.CheckLesson(r.Context(), course.ViewerFrom(r),
			strings.TrimSpace(*body.LessonId), body.EntryId, Kind(body.Kind), body.Answer)
	} else {
		result, err = h.service.Check(r.Context(), userID, body.EntryId, Kind(body.Kind), body.Answer)
	}
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, api.PracticeResult{
		Correct:   result.Correct,
		Expected:  result.Expected,
		Penalised: result.Penalised,
	})
}

func (h *Handler) scores(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}

	found, err := h.service.LessonScores(r.Context(), userID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	items := make([]api.LessonPracticeScore, 0, len(found))
	for _, score := range found {
		items = append(items, toAPIScore(score))
	}
	writeJSON(w, r, http.StatusOK, api.LessonPracticeScoreList{Items: items})
}

// recordScore là PUT chứ không phải POST: gửi lại cùng một kết quả không làm
// điểm tốt nhất đổi đi. attempts thì có tăng, nhưng đó chỉ là con số phụ.
func (h *Handler) recordScore(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}

	var body api.LessonPracticeResult
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeMalformed, "Body không phải JSON hợp lệ.", nil)
		return
	}

	score, err := h.service.RecordLessonResult(r.Context(), course.ViewerFrom(r), userID,
		chi.URLParam(r, "lessonID"), body.Correct, body.Total)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, r, http.StatusOK, toAPIScore(score))
}

func toAPIScore(score Score) api.LessonPracticeScore {
	return api.LessonPracticeScore{
		LessonId:    score.LessonID,
		BestCorrect: score.BestCorrect,
		Total:       score.Total,
		Attempts:    score.Attempts,
	}
}

/* ---------- lỗi và tiện ích ---------- */

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInvalidKind):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidation, "Dữ liệu không hợp lệ.",
			map[string]string{"kind": "phải là một dạng câu hỏi hợp lệ, và dùng được cho từ này"})
	case errors.Is(err, ErrNotFound), errors.Is(err, vocabulary.ErrNotFound), errors.Is(err, vocabulary.ErrNotUnlocked):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Không tìm thấy câu hỏi này.", nil)
	case errors.Is(err, ErrInvalidResult):
		var resultErr *ResultError
		details := map[string]string{}
		if errors.As(err, &resultErr) {
			details = resultErr.Fields
		}
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidation, "Dữ liệu không hợp lệ.", details)
	case errors.Is(err, course.ErrNotFound), errors.Is(err, course.ErrLessonNotFound):
		// Bài của khoá nháp cũng rơi vào đây, cố ý giống hệt bài không tồn tại.
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Không tìm thấy bài học này.", nil)
	case errors.Is(err, course.ErrInvalidID), errors.Is(err, ErrInvalidID):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidation, "Dữ liệu không hợp lệ.",
			map[string]string{"lesson_id": "phải là UUID hợp lệ"})
	case errors.Is(err, vocabulary.ErrInvalidID):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidation, "Dữ liệu không hợp lệ.",
			map[string]string{"entry_id": "phải là UUID hợp lệ"})
	default:
		slog.ErrorContext(r.Context(), "practice request failed", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternal, "Có lỗi xảy ra, thử lại sau.", nil)
	}
}

// currentUser đọc subject mà middleware đã gắn. Thiếu claims nghĩa là route bị
// gắn sai chỗ — trả 401 chứ không chạy tiếp với userID rỗng.
func currentUser(w http.ResponseWriter, r *http.Request) (string, bool) {
	claims, ok := auth.ClaimsFrom(r.Context())
	if !ok || claims.Subject == "" {
		httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "Thiếu access token hợp lệ.", nil)
		return "", false
	}
	return claims.Subject, true
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.ErrorContext(r.Context(), "encode practice response", slog.Any("error", err))
	}
}
