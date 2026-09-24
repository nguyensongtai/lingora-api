package vocabulary

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/nguyensongtai/lingora-api/internal/api"
	"github.com/nguyensongtai/lingora-api/internal/auth"
	"github.com/nguyensongtai/lingora-api/internal/httpx"
)

// service là những gì Handler cần ở tầng nghiệp vụ.
type service interface {
	Create(ctx context.Context, lessonID string, params CreateParams) (Entry, error)
	ListByLesson(ctx context.Context, lessonID string) ([]Entry, error)
	Update(ctx context.Context, entryID string, params UpdateParams) (Entry, error)
	Delete(ctx context.Context, entryID string) error
	List(ctx context.Context, userID string, state *State) ([]Card, error)
	Stats(ctx context.Context, userID string) (Stats, error)
	Review(ctx context.Context, userID, entryID string, grade Grade) (Review, error)
}

// Handler ánh xạ HTTP sang nghiệp vụ từ vựng.
type Handler struct {
	service service
}

func NewHandler(svc service) *Handler {
	return &Handler{service: svc}
}

// Mount gắn hai nhánh: /vocabulary là công cụ soạn nội dung (toàn bộ cần
// admin), /me/vocabulary là thứ của riêng người đang đăng nhập.
func (h *Handler) Mount(r chi.Router, adminOnly, authenticated func(http.Handler) http.Handler) {
	// Cả nhánh /vocabulary là công cụ soạn nội dung. Người học không đọc ở đây
	// mà ở /me/vocabulary, nơi hàng đợi được lọc theo bài họ đã học xong —
	// mở /vocabulary cho người ngoài là phát không toàn bộ từ và nghĩa, và vô
	// hiệu hoá luôn quy tắc mở khoá đó.
	r.Route("/vocabulary", func(r chi.Router) {
		r.Use(adminOnly)
		r.Get("/", h.listByLesson)
		r.Post("/", h.create)
		r.Patch("/{entryId}", h.update)
		r.Delete("/{entryId}", h.delete)
	})

	r.Route("/me/vocabulary", func(r chi.Router) {
		r.Use(authenticated)
		r.Get("/", h.listMine)
		r.Get("/stats", h.stats)
		r.Post("/{entryId}/review", h.review)
	})
}

/* ---------- nội dung ---------- */

func (h *Handler) listByLesson(w http.ResponseWriter, r *http.Request) {
	lessonID := r.URL.Query().Get("lesson_id")
	if lessonID == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidation, "Dữ liệu không hợp lệ.", map[string]string{
			"lesson_id": "không được rỗng",
		})
		return
	}

	entries, err := h.service.ListByLesson(r.Context(), lessonID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	items := make([]api.VocabularyEntry, 0, len(entries))
	for _, entry := range entries {
		items = append(items, ToAPIEntry(entry))
	}
	httpx.JSON(w, http.StatusOK, api.VocabularyEntryList{Items: items})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var body api.VocabularyEntryCreate
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeMalformed(w, err)
		return
	}

	created, err := h.service.Create(r.Context(), body.LessonId, CreateParams{
		Word:      body.Word,
		IPA:       derefString(body.Ipa),
		Meaning:   body.Meaning,
		Example:   derefString(body.Example),
		ExampleVI: derefString(body.ExampleVi),
	})
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Set("Location", "/v1/vocabulary/"+created.ID)
	httpx.JSON(w, http.StatusCreated, ToAPIEntry(created))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var body api.VocabularyEntryUpdate
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeMalformed(w, err)
		return
	}

	updated, err := h.service.Update(r.Context(), chi.URLParam(r, "entryId"), UpdateParams{
		Word:      body.Word,
		IPA:       body.Ipa,
		Meaning:   body.Meaning,
		Example:   body.Example,
		ExampleVI: body.ExampleVi,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, ToAPIEntry(updated))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), chi.URLParam(r, "entryId")); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

/* ---------- phía người học ---------- */

func (h *Handler) listMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}

	state, err := parseState(r.URL.Query().Get("state"))
	if err != nil {
		writeError(w, r, err)
		return
	}

	cards, err := h.service.List(r.Context(), userID, state)
	if err != nil {
		writeError(w, r, err)
		return
	}

	items := make([]api.VocabularyCard, 0, len(cards))
	for _, card := range cards {
		items = append(items, api.VocabularyCard{
			Entry:       ToAPIEntry(card.Entry),
			Level:       api.CourseLevel(card.Level),
			State:       api.VocabularyState(card.State),
			Familiarity: card.Familiarity(),
			Review:      toAPIReview(card.Review),
		})
	}
	httpx.JSON(w, http.StatusOK, api.VocabularyCardList{Items: items})
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}

	stats, err := h.service.Stats(r.Context(), userID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	httpx.JSON(w, http.StatusOK, api.VocabularyStats{
		Learned:     stats.Learned,
		DueToday:    stats.DueToday,
		Mastered:    stats.Mastered,
		NewThisWeek: stats.NewThisWeek,
		Waiting:     stats.Waiting,
	})
}

func (h *Handler) review(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}

	var body api.VocabularyReviewRequest
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeMalformed(w, err)
		return
	}

	review, err := h.service.Review(r.Context(), userID, chi.URLParam(r, "entryId"), Grade(body.Grade))
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, *toAPIReview(&review))
}

/* ---------- ánh xạ ---------- */

func parseState(raw string) (*State, error) {
	if raw == "" {
		return nil, nil
	}
	switch State(raw) {
	case StateDue, StateLearning, StateMastered, StateWaiting:
		state := State(raw)
		return &state, nil
	default:
		var v validationBuilder
		v.add("state", `phải là "due", "learning", "mastered" hoặc "waiting"`)
		return nil, v.err()
	}
}

func ToAPIEntry(entry Entry) api.VocabularyEntry {
	return api.VocabularyEntry{
		Id:        entry.ID,
		LessonId:  entry.LessonID,
		Word:      entry.Word,
		Ipa:       entry.IPA,
		Meaning:   entry.Meaning,
		Example:   entry.Example,
		ExampleVi: entry.ExampleVI,
		Position:  entry.Position,
		CreatedAt: entry.CreatedAt,
		UpdatedAt: entry.UpdatedAt,
	}
}

func toAPIReview(review *Review) *api.VocabularyReview {
	if review == nil {
		return nil
	}
	return &api.VocabularyReview{
		EaseFactor:     review.EaseFactor,
		IntervalDays:   review.IntervalDays,
		Repetitions:    review.Repetitions,
		DueOn:          review.DueOn.Format(time.DateOnly),
		LastReviewedAt: review.LastReviewedAt,
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// currentUser đọc subject mà middleware đã gắn.
func currentUser(w http.ResponseWriter, r *http.Request) (string, bool) {
	claims, ok := auth.ClaimsFrom(r.Context())
	if !ok || claims.Subject == "" {
		httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "Thiếu access token hợp lệ.", nil)
		return "", false
	}
	return claims.Subject, true
}

func writeMalformed(w http.ResponseWriter, err error) {
	httpx.Error(w, http.StatusBadRequest, httpx.CodeMalformed, "Body không đọc được.", map[string]string{
		"body": err.Error(),
	})
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr *ValidationError

	switch {
	case errors.As(err, &validationErr):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidation, "Dữ liệu không hợp lệ.", validationErr.Fields)
	case errors.Is(err, ErrInvalidID):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidation, "Dữ liệu không hợp lệ.", map[string]string{
			"id": "phải là UUID hợp lệ",
		})
	case errors.Is(err, ErrNotUnlocked):
		httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "Bạn cần hoàn thành bài chứa từ này trước khi ôn.", nil)
	case errors.Is(err, ErrLessonNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Không tìm thấy bài học.", nil)
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Không tìm thấy từ vựng.", nil)
	case errors.Is(err, ErrWordTaken):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "Từ này đã có trong bài.", map[string]string{
			"word": "đã tồn tại trong bài",
		})
	default:
		slog.ErrorContext(r.Context(), "vocabulary request failed", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternal, "Có lỗi xảy ra, vui lòng thử lại.", nil)
	}
}
