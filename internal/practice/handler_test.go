package practice_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/nguyensongtai/lingora-api/internal/auth"
	"github.com/nguyensongtai/lingora-api/internal/course"
	"github.com/nguyensongtai/lingora-api/internal/practice"
)

const handlerTestSecret = "secret-du-dai-cho-hs256-toi-thieu-32-byte"

// recordingService ghi lại nhánh nào được gọi. Luyện chung và luyện trong bài
// khác nhau đúng ở chỗ có phạt hay không, nên đi nhầm nhánh là lỗi thật.
type recordingService struct {
	called    string
	lessonID  string
	viewer    course.Viewer
	lessonErr error
}

func (s *recordingService) Session(context.Context, string, int) ([]practice.Question, error) {
	s.called = "Session"
	return nil, nil
}

func (s *recordingService) Check(context.Context, string, string, practice.Kind, string) (practice.Result, error) {
	s.called = "Check"
	return practice.Result{Penalised: true}, nil
}

func (s *recordingService) LessonSession(_ context.Context, viewer course.Viewer, lessonID string) ([]practice.Question, error) {
	s.called, s.viewer, s.lessonID = "LessonSession", viewer, lessonID
	return nil, s.lessonErr
}

func (s *recordingService) CheckLesson(_ context.Context, viewer course.Viewer, lessonID, _ string, _ practice.Kind, _ string) (practice.Result, error) {
	s.called, s.viewer, s.lessonID = "CheckLesson", viewer, lessonID
	return practice.Result{}, s.lessonErr
}

// serve đi qua đúng middleware xác thực thật, để viewer dựng từ token thật.
func serve(t *testing.T, svc *recordingService, method, target, body, role string) *httptest.ResponseRecorder {
	t.Helper()

	verifier, err := auth.NewVerifier(handlerTestSecret)
	if err != nil {
		t.Fatalf("NewVerifier() returned error: %v", err)
	}
	router := chi.NewRouter()
	practice.NewHandler(svc).Mount(router, verifier.RequireAuthenticated())

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "user-1",
		"role": role,
		"exp":  time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(handlerTestSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestSessionWithLessonIDPractisesThatLesson(t *testing.T) {
	t.Parallel()

	svc := &recordingService{}
	recorder := serve(t, svc, http.MethodGet, "/me/practice/session?lesson_id=lesson-1&size=3", "", "admin")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", recorder.Code, recorder.Body.String())
	}
	if svc.called != "LessonSession" || svc.lessonID != "lesson-1" {
		t.Errorf("gọi %s(%q), want LessonSession(lesson-1)", svc.called, svc.lessonID)
	}
	if !svc.viewer.IsAdmin {
		t.Error("viewer không mang quyền admin của token")
	}

	var body struct {
		Questions []any `json:"questions"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body.Questions == nil {
		t.Errorf("questions = %s, want [] chứ không phải null", recorder.Body.String())
	}
}

func TestSessionWithoutLessonIDKeepsTheReviewSession(t *testing.T) {
	t.Parallel()

	svc := &recordingService{}
	serve(t, svc, http.MethodGet, "/me/practice/session?size=3", "", "student")

	if svc.called != "Session" {
		t.Errorf("gọi %s, want Session", svc.called)
	}
}

func TestAnswerWithLessonIDIsGradedWithoutPenalty(t *testing.T) {
	t.Parallel()

	svc := &recordingService{}
	recorder := serve(t, svc, http.MethodPost, "/me/practice/answers",
		`{"entry_id":"e-1","kind":"multiple_choice","answer":"x","lesson_id":"lesson-1"}`, "student")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", recorder.Code, recorder.Body.String())
	}
	if svc.called != "CheckLesson" || svc.lessonID != "lesson-1" {
		t.Errorf("gọi %s(%q), want CheckLesson(lesson-1)", svc.called, svc.lessonID)
	}
	if svc.viewer.IsAdmin {
		t.Error("token student lại thành admin")
	}

	// Không có lesson_id thì vẫn là chấm có phạt như cũ.
	plain := &recordingService{}
	serve(t, plain, http.MethodPost, "/me/practice/answers",
		`{"entry_id":"e-1","kind":"multiple_choice","answer":"x"}`, "student")
	if plain.called != "Check" {
		t.Errorf("không lesson_id: gọi %s, want Check", plain.called)
	}
}

func TestLessonPracticeErrorsMapLikeTheLessonItself(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		err  error
		want int
	}{
		// Bài của khoá nháp: giống hệt bài không tồn tại, như GET /lessons/{id}.
		"bài nháp":        {err: course.ErrNotFound, want: http.StatusNotFound},
		"bài không có":    {err: course.ErrLessonNotFound, want: http.StatusNotFound},
		"id không hợp lệ": {err: course.ErrInvalidID, want: http.StatusBadRequest},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := &recordingService{lessonErr: tc.err}
			recorder := serve(t, svc, http.MethodGet, "/me/practice/session?lesson_id=x", "", "student")
			if recorder.Code != tc.want {
				t.Errorf("status = %d, want %d", recorder.Code, tc.want)
			}
		})
	}
}
