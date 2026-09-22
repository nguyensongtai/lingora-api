package vocabulary_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/nguyensongtai/lingora-api/internal/vocabulary"
)

// refusingService fail ngay nếu bị gọi: những test dưới đây chỉ quan tâm tới
// việc request có đi qua được lớp quyền hay không.
type refusingService struct{ t *testing.T }

func (s *refusingService) Create(context.Context, string, vocabulary.CreateParams) (vocabulary.Entry, error) {
	s.t.Fatal("Create đã chạy dù request lẽ ra bị chặn")
	return vocabulary.Entry{}, nil
}

func (s *refusingService) ListByLesson(context.Context, string) ([]vocabulary.Entry, error) {
	s.t.Fatal("ListByLesson đã chạy dù request lẽ ra bị chặn")
	return nil, nil
}

func (s *refusingService) Update(context.Context, string, vocabulary.UpdateParams) (vocabulary.Entry, error) {
	s.t.Fatal("Update đã chạy dù request lẽ ra bị chặn")
	return vocabulary.Entry{}, nil
}

func (s *refusingService) Delete(context.Context, string) error {
	s.t.Fatal("Delete đã chạy dù request lẽ ra bị chặn")
	return nil
}

func (s *refusingService) List(context.Context, string, *vocabulary.State) ([]vocabulary.Card, error) {
	s.t.Fatal("List đã chạy dù request lẽ ra bị chặn")
	return nil, nil
}

func (s *refusingService) Stats(context.Context, string) (vocabulary.Stats, error) {
	s.t.Fatal("Stats đã chạy dù request lẽ ra bị chặn")
	return vocabulary.Stats{}, nil
}

func (s *refusingService) Review(context.Context, string, string, vocabulary.Grade) (vocabulary.Review, error) {
	s.t.Fatal("Review đã chạy dù request lẽ ra bị chặn")
	return vocabulary.Review{}, nil
}

// blocking là middleware đóng vai lớp quyền đã từ chối, để test quan sát được
// route nào thực sự nằm sau nó.
func blocking(status int) func(http.Handler) http.Handler {
	return func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
		})
	}
}

// Nhánh /vocabulary từng để GET ở ngoài adminOnly, nên ai cũng đọc được toàn
// bộ từ và nghĩa mà không cần đăng nhập và không cần học xong bài — đi vòng
// hẳn qua quy tắc mở khoá của /me/vocabulary.
func TestContentRoutesAllRequireAdmin(t *testing.T) {
	t.Parallel()

	routes := []struct {
		method string
		target string
	}{
		{http.MethodGet, "/vocabulary?lesson_id=" + testLessonID},
		{http.MethodPost, "/vocabulary"},
		{http.MethodPatch, "/vocabulary/" + testEntryID},
		{http.MethodDelete, "/vocabulary/" + testEntryID},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.target, func(t *testing.T) {
			t.Parallel()

			recorder := serve(t, route.method, route.target)

			if recorder.Code != http.StatusForbidden {
				t.Errorf("status = %d, want 403: route này phải nằm sau adminOnly", recorder.Code)
			}
		})
	}
}

func TestPersonalRoutesRequireAuthentication(t *testing.T) {
	t.Parallel()

	routes := []struct {
		method string
		target string
	}{
		{http.MethodGet, "/me/vocabulary"},
		{http.MethodGet, "/me/vocabulary/stats"},
		{http.MethodPost, "/me/vocabulary/" + testEntryID + "/review"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.target, func(t *testing.T) {
			t.Parallel()

			recorder := serve(t, route.method, route.target)

			if recorder.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401: route này phải nằm sau authenticated", recorder.Code)
			}
		})
	}
}

const (
	testLessonID = "01929f01-0000-7000-8000-000000000003"
	testEntryID  = "01a0c48d-f504-7579-8435-53453914f916"
)

func serve(t *testing.T, method, target string) *httptest.ResponseRecorder {
	t.Helper()

	router := chi.NewRouter()
	vocabulary.NewHandler(&refusingService{t: t}).Mount(
		router,
		blocking(http.StatusForbidden),    // adminOnly
		blocking(http.StatusUnauthorized), // authenticated
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(method, target, nil))
	return recorder
}
