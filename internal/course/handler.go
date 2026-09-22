package course

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
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
	Create(ctx context.Context, params CreateParams) (Course, error)
	Get(ctx context.Context, viewer Viewer, id string) (Course, error)
	GetBySlug(ctx context.Context, viewer Viewer, slug string) (Course, error)
	List(ctx context.Context, viewer Viewer, filter ListFilter) (Page, error)
	Update(ctx context.Context, id string, params UpdateParams) (Course, error)
	Delete(ctx context.Context, id string) error
	ListLessons(ctx context.Context, viewer Viewer, courseID string) ([]Lesson, error)
	CreateLesson(ctx context.Context, courseID string, params LessonCreateParams) (Lesson, error)
	GetLesson(ctx context.Context, viewer Viewer, lessonID string) (Lesson, error)
	UpdateLesson(ctx context.Context, lessonID string, params LessonUpdateParams) (Lesson, error)
	DeleteLesson(ctx context.Context, lessonID string) error
	ReorderLessons(ctx context.Context, courseID string, lessonIDs []string) error
	ReorderCourses(ctx context.Context, level Level, courseIDs []string) error
}

// Handler ánh xạ HTTP sang nghiệp vụ khoá học.
type Handler struct {
	service service
}

// NewHandler nhận service đã sẵn sàng dùng.
func NewHandler(svc service) *Handler {
	return &Handler{service: svc}
}

// Mount gắn route của domain course. adminOnly bọc toàn bộ thao tác ghi;
// optionalAuth bọc các route đọc, vốn công khai nhưng trả nội dung khác nhau
// cho khách và cho admin.
func (h *Handler) Mount(r chi.Router, adminOnly, optionalAuth func(http.Handler) http.Handler) {
	r.Route("/courses", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(optionalAuth)
			r.Get("/", h.list)
			r.Get("/by-slug/{slug}", h.getBySlug)
			r.Get("/{courseID}", h.get)
		})

		r.Group(func(r chi.Router) {
			r.Use(adminOnly)
			r.Post("/", h.create)
			// Đặt trước /{courseID} là thừa với chi (nó khớp literal trước
			// param), nhưng để cạnh nhau cho dễ đọc.
			r.Put("/order", h.reorderCourses)
			r.Patch("/{courseID}", h.update)
			r.Delete("/{courseID}", h.delete)
		})

		// Bộ sưu tập bài học lồng dưới khoá.
		r.Route("/{courseID}/lessons", func(r chi.Router) {
			r.With(optionalAuth).Get("/", h.listLessons)

			r.Group(func(r chi.Router) {
				r.Use(adminOnly)
				r.Post("/", h.createLesson)
				r.Put("/order", h.reorderLessons)
			})
		})
	})

	// Thao tác trên từng bài nằm phẳng: sửa hay xoá không cần biết khoá nào.
	r.Route("/lessons", func(r chi.Router) {
		r.With(optionalAuth).Get("/{lessonID}", h.getLesson)

		r.Group(func(r chi.Router) {
			r.Use(adminOnly)
			r.Patch("/{lessonID}", h.updateLesson)
			r.Delete("/{lessonID}", h.deleteLesson)
		})
	})
}

/* ---------- DTO ---------- */

// Response và body tạo mới dùng thẳng type sinh từ openapi.yaml: spec lệch code
// là build gãy.
//
// Riêng body PATCH phải viết tay: oapi-codegen sinh *string cho cover_image_url
// nên không phân biệt được "không gửi field" với "gửi null", mà đó lại đúng là
// khác biệt giữa giữ nguyên ảnh bìa và xoá nó.
type updateCourseRequest struct {
	Slug          *string                `json:"slug"`
	Title         *string                `json:"title"`
	Description   *string                `json:"description"`
	Level         *string                `json:"level"`
	Status        *string                `json:"status"`
	CoverImageURL httpx.Nullable[string] `json:"cover_image_url"`
}

/* ---------- handler ---------- */

func (h *Handler) reorderCourses(w http.ResponseWriter, r *http.Request) {
	var body api.CourseOrder
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeMalformed(w, err)
		return
	}

	if err := h.service.ReorderCourses(r.Context(), Level(body.Level), body.CourseIds); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var body api.CourseCreate
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeMalformed(w, err)
		return
	}

	params := CreateParams{
		Slug:          body.Slug,
		Title:         body.Title,
		Level:         Level(body.Level),
		CoverImageURL: body.CoverImageUrl,
	}
	if body.Description != nil {
		params.Description = *body.Description
	}
	if body.Status != nil {
		params.Status = Status(*body.Status)
	}

	created, err := h.service.Create(r.Context(), params)
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Set("Location", "/v1/courses/"+created.ID)
	httpx.JSON(w, http.StatusCreated, toAPICourse(created))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	found, err := h.service.Get(r.Context(), viewerFrom(r), chi.URLParam(r, "courseID"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toAPICourse(found))
}

func (h *Handler) getBySlug(w http.ResponseWriter, r *http.Request) {
	found, err := h.service.GetBySlug(r.Context(), viewerFrom(r), chi.URLParam(r, "slug"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toAPICourse(found))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseListFilter(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	page, err := h.service.List(r.Context(), viewerFrom(r), filter)
	if err != nil {
		writeError(w, r, err)
		return
	}

	items := make([]api.Course, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, toAPICourse(item))
	}

	response := api.CourseList{Items: items}
	if page.NextCursor != nil {
		encoded := encodeCursor(*page.NextCursor)
		response.NextCursor = &encoded
	}
	httpx.JSON(w, http.StatusOK, response)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var body updateCourseRequest
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeMalformed(w, err)
		return
	}

	params := UpdateParams{
		Slug:            body.Slug,
		Title:           body.Title,
		Description:     body.Description,
		CoverImageURL:   body.CoverImageURL.Value,
		ClearCoverImage: body.CoverImageURL.IsNull(),
	}
	if body.Level != nil {
		level := Level(*body.Level)
		params.Level = &level
	}
	if body.Status != nil {
		status := Status(*body.Status)
		params.Status = &status
	}

	updated, err := h.service.Update(r.Context(), chi.URLParam(r, "courseID"), params)
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toAPICourse(updated))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), chi.URLParam(r, "courseID")); err != nil {
		writeError(w, r, err)
		return
	}
	httpx.NoContent(w)
}

func (h *Handler) listLessons(w http.ResponseWriter, r *http.Request) {
	lessons, err := h.service.ListLessons(r.Context(), viewerFrom(r), chi.URLParam(r, "courseID"))
	if err != nil {
		writeError(w, r, err)
		return
	}

	items := make([]api.Lesson, 0, len(lessons))
	for _, lesson := range lessons {
		items = append(items, toAPILesson(lesson))
	}
	httpx.JSON(w, http.StatusOK, api.LessonList{Items: items})
}

// viewerFrom dựng người đọc từ claims mà optionalAuth gắn vào. Không có claims
// nghĩa là khách vãng lai, đó là trường hợp bình thường ở đây chứ không phải lỗi.
func viewerFrom(r *http.Request) Viewer {
	claims, ok := auth.ClaimsFrom(r.Context())
	return Viewer{IsAdmin: ok && claims.Role == auth.RoleAdmin}
}

/* ---------- query, cursor, lỗi ---------- */

func parseListFilter(r *http.Request) (ListFilter, error) {
	query := r.URL.Query()
	var v validationBuilder
	filter := ListFilter{}

	if raw := strings.TrimSpace(query.Get("status")); raw != "" {
		status := Status(raw)
		filter.Status = &status
	}
	if raw := strings.TrimSpace(query.Get("level")); raw != "" {
		level := Level(raw)
		filter.Level = &level
	}
	if raw := strings.TrimSpace(query.Get("page_size")); raw != "" {
		size, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			v.add("page_size", "phải là số nguyên")
		} else {
			filter.PageSize = int32(size)
		}
	}
	if raw := strings.TrimSpace(query.Get("cursor")); raw != "" {
		cursor, err := decodeCursor(raw)
		if err != nil {
			v.add("cursor", "không hợp lệ")
		} else {
			filter.Cursor = &cursor
		}
	}

	if err := v.err(); err != nil {
		return ListFilter{}, err
	}
	return filter, nil
}

// encodeCursor gói (created_at, id) thành chuỗi mờ để client không tự chế biến.
func encodeCursor(cursor Cursor) string {
	raw := cursor.CreatedAt.UTC().Format(time.RFC3339Nano) + "|" + cursor.ID
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(raw string) (Cursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return Cursor{}, fmt.Errorf("decode cursor: %w", err)
	}

	timestamp, id, found := strings.Cut(string(decoded), "|")
	if !found || id == "" {
		return Cursor{}, errors.New("decode cursor: expected <timestamp>|<id>")
	}

	createdAt, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return Cursor{}, fmt.Errorf("decode cursor timestamp: %w", err)
	}
	return Cursor{CreatedAt: createdAt, ID: id}, nil
}

func toAPICourse(item Course) api.Course {
	return api.Course{
		Id:            item.ID,
		Slug:          item.Slug,
		Title:         item.Title,
		Description:   item.Description,
		Level:         api.CourseLevel(item.Level),
		Status:        api.CourseStatus(item.Status),
		Position:      item.Position,
		CoverImageUrl: item.CoverImageURL,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func writeMalformed(w http.ResponseWriter, err error) {
	httpx.Error(w, http.StatusBadRequest, httpx.CodeMalformed, "Body không đọc được.", map[string]string{
		"body": err.Error(),
	})
}

// writeError dịch sentinel của domain sang mã HTTP; lỗi lạ thành 500 và chỉ lộ
// chi tiết trong log.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr *ValidationError

	switch {
	case errors.As(err, &validationErr):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidation, "Dữ liệu không hợp lệ.", validationErr.Fields)
	case errors.Is(err, ErrInvalidID):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeValidation, "Dữ liệu không hợp lệ.", map[string]string{
			"id": "phải là UUID hợp lệ",
		})
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Không tìm thấy khoá học.", nil)
	case errors.Is(err, ErrLessonNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Không tìm thấy bài học.", nil)
	case errors.Is(err, ErrLessonSlugTaken):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "Slug bài học đã được dùng trong khoá này.", map[string]string{
			"slug": "đã tồn tại trong khoá",
		})
	case errors.Is(err, ErrSlugTaken):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "Slug đã được sử dụng.", map[string]string{
			"slug": "đã tồn tại",
		})
	default:
		slog.ErrorContext(r.Context(), "course request failed", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternal, "Có lỗi xảy ra, vui lòng thử lại.", nil)
	}
}
