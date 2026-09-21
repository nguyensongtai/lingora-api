package progress_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/nguyensongtai/lingora-api/internal/auth"
	"github.com/nguyensongtai/lingora-api/internal/progress"
)

const testSecret = "secret-du-dai-cho-hs256-trong-test-1234"

type fakeService struct {
	t *testing.T

	completeFn   func(context.Context, string, string) error
	uncompleteFn func(context.Context, string, string) error
	snapshotFn   func(context.Context, string) (progress.Snapshot, error)
}

func (f *fakeService) Complete(ctx context.Context, userID, lessonID string) error {
	if f.completeFn == nil {
		f.t.Fatal("Complete called unexpectedly")
	}
	return f.completeFn(ctx, userID, lessonID)
}

func (f *fakeService) Uncomplete(ctx context.Context, userID, lessonID string) error {
	if f.uncompleteFn == nil {
		f.t.Fatal("Uncomplete called unexpectedly")
	}
	return f.uncompleteFn(ctx, userID, lessonID)
}

func (f *fakeService) Snapshot(ctx context.Context, userID string) (progress.Snapshot, error) {
	if f.snapshotFn == nil {
		f.t.Fatal("Snapshot called unexpectedly")
	}
	return f.snapshotFn(ctx, userID)
}

// mount dựng router thật kèm middleware auth thật, để test đi qua đúng đường
// mà production đi — kể cả chỗ lấy user id ra khỏi token.
func mount(t *testing.T, svc *fakeService) http.Handler {
	t.Helper()

	verifier, err := auth.NewVerifier(testSecret)
	if err != nil {
		t.Fatalf("NewVerifier() returned error: %v", err)
	}

	router := chi.NewRouter()
	progress.NewHandler(svc).Mount(router, verifier.RequireAuthenticated())
	return router
}

func bearer(t *testing.T, subject string) string {
	t.Helper()

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  subject,
		"role": "student",
		"exp":  time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return "Bearer " + token
}

func do(t *testing.T, handler http.Handler, method, path, authorization string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestProgressRequiresAToken(t *testing.T) {
	t.Parallel()

	handler := mount(t, &fakeService{t: t})

	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/me/progress"},
		{http.MethodPut, "/me/progress/lessons/" + lessonID},
		{http.MethodDelete, "/me/progress/lessons/" + lessonID},
	} {
		rec := do(t, handler, tc.method, tc.path, "")
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d, want 401", tc.method, tc.path, rec.Code)
		}
	}
}

func TestCompleteUsesTheTokenSubjectAsUser(t *testing.T) {
	t.Parallel()

	var gotUser, gotLesson string
	svc := &fakeService{
		t: t,
		completeFn: func(_ context.Context, userID, lesson string) error {
			gotUser, gotLesson = userID, lesson
			return nil
		},
	}

	rec := do(t, mount(t, svc), http.MethodPut, "/me/progress/lessons/"+lessonID, bearer(t, userID))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("PUT = %d, want 204 (body %s)", rec.Code, rec.Body.String())
	}
	// Không có đường nào truyền user id từ client: nó chỉ đến từ token.
	if gotUser != userID {
		t.Errorf("userID = %q, want %q", gotUser, userID)
	}
	if gotLesson != lessonID {
		t.Errorf("lessonID = %q, want %q", gotLesson, lessonID)
	}
}

func TestUncompleteReturns204(t *testing.T) {
	t.Parallel()

	svc := &fakeService{
		t:            t,
		uncompleteFn: func(context.Context, string, string) error { return nil },
	}

	rec := do(t, mount(t, svc), http.MethodDelete, "/me/progress/lessons/"+lessonID, bearer(t, userID))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE = %d, want 204", rec.Code)
	}
}

func TestErrorsMapToStatusCodes(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		err  error
		want int
	}{
		"id sai cú pháp":    {progress.ErrInvalidID, http.StatusBadRequest},
		"bài không tồn tại": {progress.ErrLessonNotFound, http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := &fakeService{
				t:          t,
				completeFn: func(context.Context, string, string) error { return tc.err },
			}

			rec := do(t, mount(t, svc), http.MethodPut, "/me/progress/lessons/"+lessonID, bearer(t, userID))

			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestSnapshotSerialisesNullLatestCourse(t *testing.T) {
	t.Parallel()

	svc := &fakeService{
		t: t,
		snapshotFn: func(context.Context, string) (progress.Snapshot, error) {
			return progress.Snapshot{}, nil
		},
	}

	rec := do(t, mount(t, svc), http.MethodGet, "/me/progress", bearer(t, userID))

	if rec.Code != http.StatusOK {
		t.Fatalf("GET = %d, want 200", rec.Code)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	// Mảng rỗng phải ra [] chứ không phải null: phía trước lặp thẳng trên nó.
	if got := string(body["courses"]); got != "[]" {
		t.Errorf("courses = %s, want []", got)
	}
	if got := string(body["completed_lesson_ids"]); got != "[]" {
		t.Errorf("completed_lesson_ids = %s, want []", got)
	}
	if got := string(body["latest_course_id"]); got != "null" {
		t.Errorf("latest_course_id = %s, want null", got)
	}
}
