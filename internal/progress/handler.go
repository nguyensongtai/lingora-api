package progress

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/nguyensongtai/lingora-api/internal/api"
	"github.com/nguyensongtai/lingora-api/internal/auth"
	"github.com/nguyensongtai/lingora-api/internal/httpx"
)

// service là những gì Handler cần ở tầng nghiệp vụ.
type service interface {
	Complete(ctx context.Context, userID, lessonID string) error
	Uncomplete(ctx context.Context, userID, lessonID string) error
	Snapshot(ctx context.Context, userID string) (Snapshot, error)
	History(ctx context.Context, userID string, days int64) (History, error)
}

// Handler ánh xạ HTTP sang nghiệp vụ tiến độ.
type Handler struct {
	service service
}

func NewHandler(svc service) *Handler {
	return &Handler{service: svc}
}

// Mount gắn route dưới /me: tiến độ luôn là của người đang đăng nhập, nên cả
// nhánh nằm sau authenticated và không có đường nào đọc tiến độ của người khác.
func (h *Handler) Mount(r chi.Router, authenticated func(http.Handler) http.Handler) {
	r.Route("/me/progress", func(r chi.Router) {
		r.Use(authenticated)
		r.Get("/", h.snapshot)
		r.Get("/history", h.history)
		r.Put("/lessons/{lessonId}", h.complete)
		r.Delete("/lessons/{lessonId}", h.uncomplete)
	})
}

func (h *Handler) snapshot(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}

	found, err := h.service.Snapshot(r.Context(), userID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	courses := make([]api.CourseProgress, 0, len(found.Courses))
	for _, course := range found.Courses {
		courses = append(courses, api.CourseProgress{
			CourseId:       course.CourseID,
			LessonCount:    course.LessonCount,
			CompletedCount: course.CompletedCount,
		})
	}

	// Spec khai báo mảng, nên slice nil phải thành [] chứ không phải null:
	// phía trước lặp thẳng trên nó.
	completed := found.CompletedLessonIDs
	if completed == nil {
		completed = []string{}
	}

	week := toAPIDays(found.Week)

	writeJSON(w, r, http.StatusOK, api.ProgressSnapshot{
		Courses:            courses,
		CompletedLessonIds: completed,
		LatestCourseId:     found.LatestCourseID,
		TodayXp:            found.TodayXP,
		GoalXp:             found.GoalXP,
		XpPerLesson:        found.XPPerLesson,
		StreakDays:         found.StreakDays,
		Week:               week,
	})
}

func (h *Handler) complete(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}

	if err := h.service.Complete(r.Context(), userID, chi.URLParam(r, "lessonId")); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) uncomplete(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}

	if err := h.service.Uncomplete(r.Context(), userID, chi.URLParam(r, "lessonId")); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}

	// Tham số hỏng cũng đi tiếp với 0, và service kẹp nó về mặc định: days là
	// tuỳ chọn hiển thị, không phải dữ liệu người dùng nhập.
	days, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("days")), 10, 64)

	found, err := h.service.History(r.Context(), userID, days)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, r, http.StatusOK, api.ProgressHistory{
		Days:          toAPIDays(found.Days),
		TotalLessons:  found.TotalLessons,
		TotalXp:       found.TotalXP,
		ActiveDays:    found.ActiveDays,
		CurrentStreak: found.CurrentStreak,
		LongestStreak: found.LongestStreak,
		XpPerLesson:   found.XPPerLesson,
		GoalXp:        found.GoalXP,
	})
}

// toAPIDays đổi chuỗi ngày sang dạng của spec. Slice nil thành [] chứ không
// phải null: spec khai báo mảng và phía trước lặp thẳng trên nó.
func toAPIDays(days []DayActivity) []api.DayActivity {
	out := make([]api.DayActivity, 0, len(days))
	for _, day := range days {
		out = append(out, api.DayActivity{
			Date:             day.Day.Format(time.DateOnly),
			CompletedLessons: day.Completed,
		})
	}
	return out
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
		slog.ErrorContext(r.Context(), "encode progress response", slog.Any("error", err))
	}
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrInvalidID):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidation, "Dữ liệu không hợp lệ.", map[string]string{
			"id": "phải là UUID hợp lệ",
		})
	case errors.Is(err, ErrLessonNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Không tìm thấy bài học.", nil)
	default:
		slog.ErrorContext(r.Context(), "progress request failed", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternal, "Có lỗi xảy ra, vui lòng thử lại.", nil)
	}
}
