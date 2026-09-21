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
	"github.com/nguyensongtai/lingora-api/internal/httpx"
)

// service là những gì Handler cần ở tầng nghiệp vụ.
type service interface {
	Create(ctx context.Context, params CreateParams) (Course, error)
	Get(ctx context.Context, id string) (Course, error)
	GetBySlug(ctx context.Context, slug string) (Course, error)
	List(ctx context.Context, filter ListFilter) (Page, error)
	Update(ctx context.Context, id string, params UpdateParams) (Course, error)
	Delete(ctx context.Context, id string) error
	ListLessons(ctx context.Context, courseID string) ([]Lesson, error)
}

// Handler ánh xạ HTTP sang nghiệp vụ khoá học.
type Handler struct {
	service service
}

// NewHandler nhận service đã sẵn sàng dùng.
func NewHandler(svc service) *Handler {
	return &Handler{service: svc}
}

// Mount gắn route của domain course; adminOnly bọc toàn bộ thao tác ghi.
func (h *Handler) Mount(r chi.Router, adminOnly func(http.Handler) http.Handler) {
	r.Route("/courses", func(r chi.Router) {
		r.Get("/", h.list)
		r.Get("/by-slug/{slug}", h.getBySlug)
		r.Get("/{courseID}", h.get)
		r.Get("/{courseID}/lessons", h.listLessons)

		r.Group(func(r chi.Router) {
			r.Use(adminOnly)
			r.Post("/", h.create)
			r.Patch("/{courseID}", h.update)
			r.Delete("/{courseID}", h.delete)
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
	found, err := h.service.Get(r.Context(), chi.URLParam(r, "courseID"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toAPICourse(found))
}

func (h *Handler) getBySlug(w http.ResponseWriter, r *http.Request) {
	found, err := h.service.GetBySlug(r.Context(), chi.URLParam(r, "slug"))
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

	page, err := h.service.List(r.Context(), filter)
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
	lessons, err := h.service.ListLessons(r.Context(), chi.URLParam(r, "courseID"))
	if err != nil {
		writeError(w, r, err)
		return
	}

	items := make([]api.Lesson, 0, len(lessons))
	for _, lesson := range lessons {
		items = append(items, api.Lesson{
			Id:        lesson.ID,
			CourseId:  lesson.CourseID,
			Slug:      lesson.Slug,
			Title:     lesson.Title,
			Position:  lesson.Position,
			CreatedAt: lesson.CreatedAt,
			UpdatedAt: lesson.UpdatedAt,
		})
	}
	httpx.JSON(w, http.StatusOK, api.LessonList{Items: items})
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
	case errors.Is(err, ErrSlugTaken):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "Slug đã được sử dụng.", map[string]string{
			"slug": "đã tồn tại",
		})
	default:
		slog.ErrorContext(r.Context(), "course request failed", slog.Any("error", err))
		httpx.Error(w, http.StatusInternalServerError, httpx.CodeInternal, "Có lỗi xảy ra, vui lòng thử lại.", nil)
	}
}
