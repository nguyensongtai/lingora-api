package course_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/nguyensongtai/lingora-api/internal/auth"
	"github.com/nguyensongtai/lingora-api/internal/course"
)

type fakeService struct {
	t *testing.T

	// lastViewer giữ viewer của lần đọc gần nhất, để test kiểm handler có
	// thực sự dựng nó từ token hay không.
	lastViewer course.Viewer

	createFn      func(context.Context, course.CreateParams) (course.Course, error)
	getFn         func(context.Context, string) (course.Course, error)
	getBySlugFn   func(context.Context, string) (course.Course, error)
	listFn        func(context.Context, course.ListFilter) (course.Page, error)
	updateFn      func(context.Context, string, course.UpdateParams) (course.Course, error)
	deleteFn      func(context.Context, string) error
	listLessonsFn func(context.Context, string) ([]course.Lesson, error)

	createLessonFn   func(context.Context, string, course.LessonCreateParams) (course.Lesson, error)
	getLessonFn      func(context.Context, string) (course.Lesson, error)
	updateLessonFn   func(context.Context, string, course.LessonUpdateParams) (course.Lesson, error)
	deleteLessonFn   func(context.Context, string) error
	reorderLessonsFn func(context.Context, string, []string) error
	reorderCoursesFn func(context.Context, course.Level, []string) error

	reorderedIDs []string
}

func (f *fakeService) Create(ctx context.Context, params course.CreateParams) (course.Course, error) {
	if f.createFn == nil {
		f.t.Fatal("Create called unexpectedly")
	}
	return f.createFn(ctx, params)
}

func (f *fakeService) Get(ctx context.Context, viewer course.Viewer, id string) (course.Course, error) {
	f.lastViewer = viewer
	if f.getFn == nil {
		f.t.Fatal("Get called unexpectedly")
	}
	return f.getFn(ctx, id)
}

func (f *fakeService) GetBySlug(ctx context.Context, viewer course.Viewer, slug string) (course.Course, error) {
	f.lastViewer = viewer
	if f.getBySlugFn == nil {
		f.t.Fatal("GetBySlug called unexpectedly")
	}
	return f.getBySlugFn(ctx, slug)
}

func (f *fakeService) List(ctx context.Context, viewer course.Viewer, filter course.ListFilter) (course.Page, error) {
	f.lastViewer = viewer
	if f.listFn == nil {
		f.t.Fatal("List called unexpectedly")
	}
	return f.listFn(ctx, filter)
}

func (f *fakeService) Update(ctx context.Context, id string, params course.UpdateParams) (course.Course, error) {
	if f.updateFn == nil {
		f.t.Fatal("Update called unexpectedly")
	}
	return f.updateFn(ctx, id, params)
}

func (f *fakeService) Delete(ctx context.Context, id string) error {
	if f.deleteFn == nil {
		f.t.Fatal("Delete called unexpectedly")
	}
	return f.deleteFn(ctx, id)
}

func (f *fakeService) ListLessons(ctx context.Context, viewer course.Viewer, courseID string) ([]course.Lesson, error) {
	f.lastViewer = viewer
	if f.listLessonsFn == nil {
		f.t.Fatal("ListLessons called unexpectedly")
	}
	return f.listLessonsFn(ctx, courseID)
}

// newServer gắn route thật; adminOnly là no-op vì quyền được test riêng ở package auth.
func newServer(svc *fakeService) http.Handler {
	router := chi.NewRouter()
	noop := func(next http.Handler) http.Handler { return next }
	course.NewHandler(svc).Mount(router, noop, noop)
	return router
}

func do(t *testing.T, svc *fakeService, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()

	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, target, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	newServer(svc).ServeHTTP(recorder, req)
	return recorder
}

// decodeBody đọc JSON response vào map để kiểm tra đúng tên field trên dây.
func decodeBody(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not a JSON object: %v (body: %s)", err, recorder.Body.String())
	}
	return body
}

func sampleCourse() course.Course {
	return course.Course{
		ID:        "018f3a9c-7b2e-7c31-9a55-0f1d2e3a4b5c",
		Slug:      "ngu-phap-co-ban",
		Title:     "Ngữ pháp cơ bản",
		Level:     course.LevelA1,
		Status:    course.StatusDraft,
		CreatedAt: time.Date(2026, time.September, 21, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, time.September, 21, 10, 0, 0, 0, time.UTC),
	}
}

func TestCreateReturns201WithLocation(t *testing.T) {
	t.Parallel()

	svc := &fakeService{t: t, createFn: func(context.Context, course.CreateParams) (course.Course, error) {
		return sampleCourse(), nil
	}}

	recorder := do(t, svc, http.MethodPost, "/courses", `{"slug":"ngu-phap-co-ban","title":"Ngữ pháp cơ bản","level":"A1"}`)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body: %s)", recorder.Code, recorder.Body.String())
	}
	if want := "/v1/courses/" + sampleCourse().ID; recorder.Header().Get("Location") != want {
		t.Errorf("Location = %q, want %q", recorder.Header().Get("Location"), want)
	}
	if body := decodeBody(t, recorder); body["slug"] != "ngu-phap-co-ban" {
		t.Errorf("body slug = %v, want ngu-phap-co-ban", body["slug"])
	}
}

func TestErrorMapping(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err        error
		wantStatus int
		wantCode   string
	}{
		"không tìm thấy":   {course.ErrNotFound, http.StatusNotFound, "not_found"},
		"id sai định dạng": {course.ErrInvalidID, http.StatusBadRequest, "validation_error"},
		"trùng slug":       {course.ErrSlugTaken, http.StatusConflict, "conflict"},
		"lỗi lạ":           {errors.New("connection reset by peer"), http.StatusInternalServerError, "internal_error"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := &fakeService{t: t, getFn: func(context.Context, string) (course.Course, error) {
				return course.Course{}, tc.err
			}}

			recorder := do(t, svc, http.MethodGet, "/courses/id-1", "")
			if recorder.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tc.wantStatus)
			}

			body := decodeBody(t, recorder)
			if body["code"] != tc.wantCode {
				t.Errorf("code = %v, want %q", body["code"], tc.wantCode)
			}
			if body["message"] == "" || body["message"] == nil {
				t.Error("message is empty, want a human readable message")
			}
		})
	}
}

func TestInternalErrorDoesNotLeakDetails(t *testing.T) {
	t.Parallel()

	svc := &fakeService{t: t, getFn: func(context.Context, string) (course.Course, error) {
		return course.Course{}, errors.New("pq: password authentication failed for user \"lingora\"")
	}}

	recorder := do(t, svc, http.MethodGet, "/courses/id-1", "")
	if strings.Contains(recorder.Body.String(), "password") {
		t.Fatalf("response leaks the underlying error: %s", recorder.Body.String())
	}
}

func TestValidationErrorBecomesFieldDetails(t *testing.T) {
	t.Parallel()

	svc := &fakeService{t: t, createFn: func(context.Context, course.CreateParams) (course.Course, error) {
		return course.Course{}, &course.ValidationError{Fields: map[string]string{"slug": "không được rỗng"}}
	}}

	recorder := do(t, svc, http.MethodPost, "/courses", `{"title":"Tiêu đề","level":"A1"}`)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}

	body := decodeBody(t, recorder)
	details, ok := body["details"].(map[string]any)
	if !ok {
		t.Fatalf("details = %v, want an object", body["details"])
	}
	if details["slug"] != "không được rỗng" {
		t.Errorf("details.slug = %v, want the field message", details["slug"])
	}
}

func TestUnknownFieldInBodyIsRejected(t *testing.T) {
	t.Parallel()

	// createFn để nil: body hỏng thì không được chạm tới service.
	recorder := do(t, &fakeService{t: t}, http.MethodPost, "/courses",
		`{"slug":"a","title":"b","level":"A1","titel":"typo"}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
	if code := decodeBody(t, recorder)["code"]; code != "malformed_body" {
		t.Errorf("code = %v, want malformed_body", code)
	}
}

func TestListEncodesNextCursorAndDecodesItBack(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.September, 21, 10, 0, 0, 0, time.UTC)

	var encoded string
	t.Run("encode", func(t *testing.T) {
		svc := &fakeService{t: t, listFn: func(context.Context, course.ListFilter) (course.Page, error) {
			return course.Page{
				Items:      []course.Course{sampleCourse()},
				NextCursor: &course.Cursor{CreatedAt: createdAt, ID: "id-9"},
			}, nil
		}}

		recorder := do(t, svc, http.MethodGet, "/courses", "")
		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", recorder.Code)
		}

		next, ok := decodeBody(t, recorder)["next_cursor"].(string)
		if !ok {
			t.Fatal("next_cursor is not a string")
		}
		if _, err := base64.RawURLEncoding.DecodeString(next); err != nil {
			t.Errorf("next_cursor is not base64url: %v", err)
		}
		encoded = next
	})

	t.Run("decode", func(t *testing.T) {
		if encoded == "" {
			t.Skip("encode subtest did not run")
		}

		var got course.ListFilter
		svc := &fakeService{t: t, listFn: func(_ context.Context, filter course.ListFilter) (course.Page, error) {
			got = filter
			return course.Page{}, nil
		}}

		if recorder := do(t, svc, http.MethodGet, "/courses?cursor="+encoded, ""); recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", recorder.Code)
		}
		if got.Cursor == nil {
			t.Fatal("filter.Cursor = nil, want the decoded cursor")
		}
		if got.Cursor.ID != "id-9" || !got.Cursor.CreatedAt.Equal(createdAt) {
			t.Errorf("cursor = %+v, want id-9 at %v", got.Cursor, createdAt)
		}
	})
}

func TestListRejectsBadQueryParams(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		target    string
		wantField string
	}{
		"page_size không phải số":   {"/courses?page_size=abc", "page_size"},
		"cursor không giải mã được": {"/courses?cursor=!!!", "cursor"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			recorder := do(t, &fakeService{t: t}, http.MethodGet, tc.target, "")
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", recorder.Code)
			}

			details, ok := decodeBody(t, recorder)["details"].(map[string]any)
			if !ok {
				t.Fatal("details is not an object")
			}
			if _, ok := details[tc.wantField]; !ok {
				t.Errorf("details = %v, want an entry for %q", details, tc.wantField)
			}
		})
	}
}

func TestListPassesFiltersThrough(t *testing.T) {
	t.Parallel()

	var got course.ListFilter
	svc := &fakeService{t: t, listFn: func(_ context.Context, filter course.ListFilter) (course.Page, error) {
		got = filter
		return course.Page{}, nil
	}}

	if recorder := do(t, svc, http.MethodGet, "/courses?status=published&level=B2&page_size=5", ""); recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	if got.Status == nil || *got.Status != course.StatusPublished {
		t.Errorf("status filter = %v, want published", got.Status)
	}
	if got.Level == nil || *got.Level != course.LevelB2 {
		t.Errorf("level filter = %v, want B2", got.Level)
	}
	if got.PageSize != 5 {
		t.Errorf("page size = %d, want 5", got.PageSize)
	}
}

func TestListReturnsEmptyArrayNotNull(t *testing.T) {
	t.Parallel()

	svc := &fakeService{t: t, listFn: func(context.Context, course.ListFilter) (course.Page, error) {
		return course.Page{}, nil
	}}

	recorder := do(t, svc, http.MethodGet, "/courses", "")
	if !strings.Contains(recorder.Body.String(), `"items":[]`) {
		t.Errorf("body = %s, want items to be an empty array", recorder.Body.String())
	}
	if next := decodeBody(t, recorder)["next_cursor"]; next != nil {
		t.Errorf("next_cursor = %v, want null", next)
	}
}

func TestUpdateDistinguishesAbsentNullAndValueForCoverImage(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		body      string
		wantClear bool
		wantValue *string
	}{
		"vắng mặt thì giữ nguyên": {`{"title":"Tiêu đề mới"}`, false, nil},
		"null thì xoá ảnh":        {`{"cover_image_url":null}`, true, nil},
		"có giá trị thì đặt mới":  {`{"cover_image_url":"https://cdn.example.com/a.png"}`, false, ptr("https://cdn.example.com/a.png")},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got course.UpdateParams
			svc := &fakeService{t: t, updateFn: func(_ context.Context, _ string, params course.UpdateParams) (course.Course, error) {
				got = params
				return sampleCourse(), nil
			}}

			if recorder := do(t, svc, http.MethodPatch, "/courses/id-1", tc.body); recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", recorder.Code)
			}

			if got.ClearCoverImage != tc.wantClear {
				t.Errorf("ClearCoverImage = %v, want %v", got.ClearCoverImage, tc.wantClear)
			}
			switch {
			case tc.wantValue == nil && got.CoverImageURL != nil:
				t.Errorf("CoverImageURL = %q, want nil", *got.CoverImageURL)
			case tc.wantValue != nil && (got.CoverImageURL == nil || *got.CoverImageURL != *tc.wantValue):
				t.Errorf("CoverImageURL = %v, want %q", got.CoverImageURL, *tc.wantValue)
			}
		})
	}
}

func TestDeleteReturns204WithoutBody(t *testing.T) {
	t.Parallel()

	svc := &fakeService{t: t, deleteFn: func(context.Context, string) error { return nil }}

	recorder := do(t, svc, http.MethodDelete, "/courses/id-1", "")
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", recorder.Body.String())
	}
}

func TestListLessonsReturnsItems(t *testing.T) {
	t.Parallel()

	svc := &fakeService{t: t, listLessonsFn: func(context.Context, string) ([]course.Lesson, error) {
		return []course.Lesson{{ID: "lesson-1", CourseID: "id-1", Slug: "bai-1", Title: "Bài 1", Position: 0}}, nil
	}}

	recorder := do(t, svc, http.MethodGet, "/courses/id-1/lessons", "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	items, ok := decodeBody(t, recorder)["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("items = %v, want exactly one lesson", decodeBody(t, recorder)["items"])
	}
	first, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("items[0] = %v, want an object", items[0])
	}
	if first["course_id"] != "id-1" {
		t.Errorf("items[0].course_id = %v, want id-1", first["course_id"])
	}
}

func TestGetBySlugUsesItsOwnRoute(t *testing.T) {
	t.Parallel()

	var got string
	svc := &fakeService{t: t, getBySlugFn: func(_ context.Context, slug string) (course.Course, error) {
		got = slug
		return sampleCourse(), nil
	}}

	if recorder := do(t, svc, http.MethodGet, "/courses/by-slug/ngu-phap-co-ban", ""); recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if got != "ngu-phap-co-ban" {
		t.Errorf("slug = %q, want ngu-phap-co-ban", got)
	}
}

func (f *fakeService) CreateLesson(ctx context.Context, courseID string, params course.LessonCreateParams) (course.Lesson, error) {
	if f.createLessonFn == nil {
		f.t.Fatal("CreateLesson called unexpectedly")
	}
	return f.createLessonFn(ctx, courseID, params)
}

func (f *fakeService) GetLesson(ctx context.Context, viewer course.Viewer, lessonID string) (course.Lesson, error) {
	f.lastViewer = viewer
	if f.getLessonFn == nil {
		f.t.Fatal("GetLesson called unexpectedly")
	}
	return f.getLessonFn(ctx, lessonID)
}

func (f *fakeService) UpdateLesson(ctx context.Context, lessonID string, params course.LessonUpdateParams) (course.Lesson, error) {
	if f.updateLessonFn == nil {
		f.t.Fatal("UpdateLesson called unexpectedly")
	}
	return f.updateLessonFn(ctx, lessonID, params)
}

func (f *fakeService) DeleteLesson(ctx context.Context, lessonID string) error {
	if f.deleteLessonFn == nil {
		f.t.Fatal("DeleteLesson called unexpectedly")
	}
	return f.deleteLessonFn(ctx, lessonID)
}

func (f *fakeService) ReorderCourses(ctx context.Context, level course.Level, courseIDs []string) error {
	if f.reorderCoursesFn == nil {
		f.t.Fatal("ReorderCourses called unexpectedly")
	}
	return f.reorderCoursesFn(ctx, level, courseIDs)
}

func (f *fakeService) ReorderLessons(ctx context.Context, courseID string, lessonIDs []string) error {
	f.reorderedIDs = lessonIDs
	if f.reorderLessonsFn == nil {
		f.t.Fatal("ReorderLessons called unexpectedly")
	}
	return f.reorderLessonsFn(ctx, courseID, lessonIDs)
}

func TestCreateLessonReturns201WithLocation(t *testing.T) {
	t.Parallel()

	svc := &fakeService{t: t, createLessonFn: func(context.Context, string, course.LessonCreateParams) (course.Lesson, error) {
		return course.Lesson{ID: "lesson-1", CourseID: "course-1", Slug: "bai-1", Title: "Bài 1"}, nil
	}}

	recorder := do(t, svc, http.MethodPost, "/courses/course-1/lessons", `{"slug":"bai-1","title":"Bài 1"}`)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body: %s)", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Location"); got != "/v1/lessons/lesson-1" {
		t.Errorf("Location = %q, want /v1/lessons/lesson-1", got)
	}
}

func TestLessonItemRoutesAreFlat(t *testing.T) {
	t.Parallel()

	t.Run("get", func(t *testing.T) {
		t.Parallel()

		var got string
		svc := &fakeService{t: t, getLessonFn: func(_ context.Context, lessonID string) (course.Lesson, error) {
			got = lessonID
			return course.Lesson{ID: lessonID}, nil
		}}

		if recorder := do(t, svc, http.MethodGet, "/lessons/lesson-9", ""); recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", recorder.Code)
		}
		if got != "lesson-9" {
			t.Errorf("service received %q, want lesson-9", got)
		}
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()

		svc := &fakeService{t: t, deleteLessonFn: func(context.Context, string) error { return nil }}

		if recorder := do(t, svc, http.MethodDelete, "/lessons/lesson-9", ""); recorder.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want 204", recorder.Code)
		}
	})
}

func TestReorderLessonsPassesTheOrderThrough(t *testing.T) {
	t.Parallel()

	svc := &fakeService{t: t, reorderLessonsFn: func(context.Context, string, []string) error { return nil }}

	recorder := do(t, svc, http.MethodPut, "/courses/course-1/lessons/order",
		`{"lesson_ids":["c","a","b"]}`)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204 (body: %s)", recorder.Code, recorder.Body.String())
	}
	if len(svc.reorderedIDs) != 3 || svc.reorderedIDs[0] != "c" {
		t.Errorf("service received %v, want the order preserved", svc.reorderedIDs)
	}
}

func TestLessonErrorMapping(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err        error
		wantStatus int
		wantCode   string
	}{
		"bài không tồn tại":     {course.ErrLessonNotFound, http.StatusNotFound, "not_found"},
		"trùng slug trong khoá": {course.ErrLessonSlugTaken, http.StatusConflict, "conflict"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := &fakeService{t: t, getLessonFn: func(context.Context, string) (course.Lesson, error) {
				return course.Lesson{}, tc.err
			}}

			recorder := do(t, svc, http.MethodGet, "/lessons/lesson-1", "")
			if recorder.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tc.wantStatus)
			}
			if code := decodeBody(t, recorder)["code"]; code != tc.wantCode {
				t.Errorf("code = %v, want %q", code, tc.wantCode)
			}
		})
	}
}

/* ---------- handler dựng viewer từ token thật ---------- */

const handlerTestSecret = "secret-du-dai-cho-hs256-toi-thieu-32-byte"

// serverWithAuth gắn đúng middleware OptionalAuth thật, để test đi qua cả
// đường token → claims → viewer chứ không giả lập khúc giữa.
func serverWithAuth(t *testing.T, svc *fakeService) http.Handler {
	t.Helper()

	verifier, err := auth.NewVerifier(handlerTestSecret)
	if err != nil {
		t.Fatalf("NewVerifier() returned error: %v", err)
	}

	router := chi.NewRouter()
	noop := func(next http.Handler) http.Handler { return next }
	course.NewHandler(svc).Mount(router, noop, verifier.OptionalAuth())
	return router
}

func bearerFor(t *testing.T, role string) string {
	t.Helper()

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "user-1",
		"role": role,
		"exp":  time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(handlerTestSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return "Bearer " + token
}

func TestHandlerBuildsViewerFromToken(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		authorization string
		wantAdmin     bool
	}{
		{name: "không token", authorization: "", wantAdmin: false},
		{name: "token student", authorization: bearerFor(t, "student"), wantAdmin: false},
		{name: "token admin", authorization: bearerFor(t, "admin"), wantAdmin: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := &fakeService{t: t, listFn: func(context.Context, course.ListFilter) (course.Page, error) {
				return course.Page{Items: []course.Course{}}, nil
			}}
			req := httptest.NewRequest(http.MethodGet, "/courses", nil)
			if tc.authorization != "" {
				req.Header.Set("Authorization", tc.authorization)
			}
			recorder := httptest.NewRecorder()

			serverWithAuth(t, svc).ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", recorder.Code)
			}
			if svc.lastViewer.IsAdmin != tc.wantAdmin {
				t.Errorf("viewer.IsAdmin = %t, want %t", svc.lastViewer.IsAdmin, tc.wantAdmin)
			}
		})
	}
}
