package course

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/nguyensongtai/lingora-api/internal/api"
	"github.com/nguyensongtai/lingora-api/internal/httpx"
)

func (h *Handler) createLesson(w http.ResponseWriter, r *http.Request) {
	var body api.LessonCreate
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeMalformed(w, err)
		return
	}

	courseID := chi.URLParam(r, "courseID")
	created, err := h.service.CreateLesson(r.Context(), courseID, LessonCreateParams{
		Slug:  body.Slug,
		Title: body.Title,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Set("Location", "/v1/lessons/"+created.ID)
	httpx.JSON(w, http.StatusCreated, toAPILesson(created))
}

func (h *Handler) getLesson(w http.ResponseWriter, r *http.Request) {
	found, err := h.service.GetLesson(r.Context(), chi.URLParam(r, "lessonID"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toAPILesson(found))
}

func (h *Handler) updateLesson(w http.ResponseWriter, r *http.Request) {
	var body api.LessonUpdate
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeMalformed(w, err)
		return
	}

	updated, err := h.service.UpdateLesson(r.Context(), chi.URLParam(r, "lessonID"), LessonUpdateParams{
		Slug:  body.Slug,
		Title: body.Title,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toAPILesson(updated))
}

func (h *Handler) deleteLesson(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteLesson(r.Context(), chi.URLParam(r, "lessonID")); err != nil {
		writeError(w, r, err)
		return
	}
	httpx.NoContent(w)
}

func (h *Handler) reorderLessons(w http.ResponseWriter, r *http.Request) {
	var body api.LessonOrder
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		writeMalformed(w, err)
		return
	}

	if err := h.service.ReorderLessons(r.Context(), chi.URLParam(r, "courseID"), body.LessonIds); err != nil {
		writeError(w, r, err)
		return
	}
	httpx.NoContent(w)
}

func toAPILesson(item Lesson) api.Lesson {
	return api.Lesson{
		Id:        item.ID,
		CourseId:  item.CourseID,
		Slug:      item.Slug,
		Title:     item.Title,
		Position:  item.Position,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
