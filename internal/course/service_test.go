package course_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/nguyensongtai/lingora-api/internal/course"
)

// fakeRepo cho phép mỗi test chỉ khai đúng method nó quan tâm; method không khai
// sẽ fail test nếu bị gọi ngoài ý muốn.
type fakeRepo struct {
	t *testing.T

	createFn      func(context.Context, course.CreateParams) (course.Course, error)
	getByIDFn     func(context.Context, string) (course.Course, error)
	getBySlugFn   func(context.Context, string) (course.Course, error)
	listFn        func(context.Context, course.ListFilter) ([]course.Course, error)
	updateFn      func(context.Context, string, course.UpdateParams) (course.Course, error)
	softDeleteFn  func(context.Context, string) error
	existsFn      func(context.Context, string) (bool, error)
	listLessonsFn func(context.Context, string) ([]course.Lesson, error)

	createLessonFn     func(context.Context, string, course.LessonCreateParams) (course.Lesson, error)
	getLessonFn        func(context.Context, string) (course.Lesson, error)
	updateLessonFn     func(context.Context, string, course.LessonUpdateParams) (course.Lesson, error)
	softDeleteLessonFn func(context.Context, string) error
	listLessonIDsFn    func(context.Context, string) ([]string, error)
	reorderLessonsFn   func(context.Context, string, []string) (int64, error)

	listLessonsCalled  bool
	reorderLessonsArgs []string
}

func (f *fakeRepo) Create(ctx context.Context, params course.CreateParams) (course.Course, error) {
	if f.createFn == nil {
		f.t.Fatal("Create called unexpectedly")
	}
	return f.createFn(ctx, params)
}

func (f *fakeRepo) GetByID(ctx context.Context, id string) (course.Course, error) {
	if f.getByIDFn == nil {
		f.t.Fatal("GetByID called unexpectedly")
	}
	return f.getByIDFn(ctx, id)
}

func (f *fakeRepo) GetBySlug(ctx context.Context, slug string) (course.Course, error) {
	if f.getBySlugFn == nil {
		f.t.Fatal("GetBySlug called unexpectedly")
	}
	return f.getBySlugFn(ctx, slug)
}

func (f *fakeRepo) List(ctx context.Context, filter course.ListFilter) ([]course.Course, error) {
	if f.listFn == nil {
		f.t.Fatal("List called unexpectedly")
	}
	return f.listFn(ctx, filter)
}

func (f *fakeRepo) Update(ctx context.Context, id string, params course.UpdateParams) (course.Course, error) {
	if f.updateFn == nil {
		f.t.Fatal("Update called unexpectedly")
	}
	return f.updateFn(ctx, id, params)
}

func (f *fakeRepo) SoftDelete(ctx context.Context, id string) error {
	if f.softDeleteFn == nil {
		f.t.Fatal("SoftDelete called unexpectedly")
	}
	return f.softDeleteFn(ctx, id)
}

func (f *fakeRepo) Exists(ctx context.Context, id string) (bool, error) {
	if f.existsFn == nil {
		f.t.Fatal("Exists called unexpectedly")
	}
	return f.existsFn(ctx, id)
}

func (f *fakeRepo) ListLessons(ctx context.Context, courseID string) ([]course.Lesson, error) {
	f.listLessonsCalled = true
	if f.listLessonsFn == nil {
		f.t.Fatal("ListLessons called unexpectedly")
	}
	return f.listLessonsFn(ctx, courseID)
}

//go:fix inline
func ptr[T any](value T) *T { return new(value) }

// fieldsOf lấy chi tiết lỗi theo field, fail test nếu err không phải lỗi validate.
func fieldsOf(t *testing.T, err error) map[string]string {
	t.Helper()

	if !errors.Is(err, course.ErrValidation) {
		t.Fatalf("error = %v, want it to match ErrValidation", err)
	}
	var validationErr *course.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %v, want a *course.ValidationError", err)
	}
	return validationErr.Fields
}

func TestCreateNormalisesInputAndDefaultsStatus(t *testing.T) {
	t.Parallel()

	var got course.CreateParams
	repo := &fakeRepo{t: t, createFn: func(_ context.Context, params course.CreateParams) (course.Course, error) {
		got = params
		return course.Course{ID: "id-1", Slug: params.Slug}, nil
	}}

	created, err := course.NewService(repo).Create(context.Background(), course.CreateParams{
		Slug:        "  Ngu-Phap-Co-Ban  ",
		Title:       "  Ngữ pháp cơ bản  ",
		Description: "  Khoá nhập môn  ",
		Level:       course.LevelA1,
	})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	if got.Slug != "ngu-phap-co-ban" {
		t.Errorf("slug = %q, want %q", got.Slug, "ngu-phap-co-ban")
	}
	if got.Title != "Ngữ pháp cơ bản" {
		t.Errorf("title = %q, want it trimmed", got.Title)
	}
	if got.Description != "Khoá nhập môn" {
		t.Errorf("description = %q, want it trimmed", got.Description)
	}
	if got.Status != course.StatusDraft {
		t.Errorf("status = %q, want the draft default", got.Status)
	}
	if created.ID != "id-1" {
		t.Errorf("returned course id = %q, want %q", created.ID, "id-1")
	}
}

func TestCreateRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		params     course.CreateParams
		wantFields []string
	}{
		"slug rỗng": {
			params:     course.CreateParams{Title: "Tiêu đề", Level: course.LevelA1},
			wantFields: []string{"slug"},
		},
		"slug sai định dạng": {
			params:     course.CreateParams{Slug: "Ngữ Pháp!", Title: "Tiêu đề", Level: course.LevelA1},
			wantFields: []string{"slug"},
		},
		"title rỗng": {
			params:     course.CreateParams{Slug: "grammar", Level: course.LevelA1},
			wantFields: []string{"title"},
		},
		"level không hợp lệ": {
			params:     course.CreateParams{Slug: "grammar", Title: "Tiêu đề", Level: course.Level("Z9")},
			wantFields: []string{"level"},
		},
		"status không hợp lệ": {
			params:     course.CreateParams{Slug: "grammar", Title: "Tiêu đề", Level: course.LevelA1, Status: course.Status("live")},
			wantFields: []string{"status"},
		},
		"cover image không phải url": {
			params: course.CreateParams{
				Slug: "grammar", Title: "Tiêu đề", Level: course.LevelA1,
				CoverImageURL: new("khong-phai-url"),
			},
			wantFields: []string{"cover_image_url"},
		},
		"nhiều lỗi cùng lúc": {
			params:     course.CreateParams{Slug: "BAD SLUG", Level: course.Level("")},
			wantFields: []string{"slug", "title", "level"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// createFn để nil: repo bị gọi nghĩa là validate đã lọt.
			service := course.NewService(&fakeRepo{t: t})

			_, err := service.Create(context.Background(), tc.params)
			fields := fieldsOf(t, err)

			for _, field := range tc.wantFields {
				if _, ok := fields[field]; !ok {
					t.Errorf("fields = %v, want an entry for %q", fields, field)
				}
			}
			if len(fields) != len(tc.wantFields) {
				t.Errorf("fields = %v, want exactly %d entries", fields, len(tc.wantFields))
			}
		})
	}
}

func TestCreatePropagatesSlugTaken(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{t: t, createFn: func(context.Context, course.CreateParams) (course.Course, error) {
		return course.Course{}, course.ErrSlugTaken
	}}

	_, err := course.NewService(repo).Create(context.Background(), course.CreateParams{
		Slug: "grammar", Title: "Tiêu đề", Level: course.LevelB1,
	})
	if !errors.Is(err, course.ErrSlugTaken) {
		t.Fatalf("error = %v, want it to match ErrSlugTaken", err)
	}
}

func TestGetPropagatesNotFound(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{t: t, getByIDFn: func(context.Context, string) (course.Course, error) {
		return course.Course{}, course.ErrNotFound
	}}

	if _, err := course.NewService(repo).Get(context.Background(), "id-1"); !errors.Is(err, course.ErrNotFound) {
		t.Fatalf("error = %v, want it to match ErrNotFound", err)
	}
}

func TestListClampsPageSizeAndAsksForOneExtraRow(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		requested int32
		wantAsked int32
	}{
		"không truyền thì dùng mặc định": {requested: 0, wantAsked: course.DefaultPageSize + 1},
		"số âm cũng dùng mặc định":       {requested: -5, wantAsked: course.DefaultPageSize + 1},
		"trong khoảng thì giữ nguyên":    {requested: 7, wantAsked: 8},
		"vượt trần thì bị chặn":          {requested: 500, wantAsked: course.MaxPageSize + 1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var asked int32
			repo := &fakeRepo{t: t, listFn: func(_ context.Context, filter course.ListFilter) ([]course.Course, error) {
				asked = filter.PageSize
				return nil, nil
			}}

			if _, err := course.NewService(repo).List(context.Background(), course.ListFilter{PageSize: tc.requested}); err != nil {
				t.Fatalf("List() returned error: %v", err)
			}
			if asked != tc.wantAsked {
				t.Errorf("repo asked for %d rows, want %d", asked, tc.wantAsked)
			}
		})
	}
}

func TestListBuildsNextCursorOnlyWhenMoreRowsExist(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.September, 21, 10, 0, 0, 0, time.UTC)
	rows := []course.Course{
		{ID: "id-1", CreatedAt: createdAt},
		{ID: "id-2", CreatedAt: createdAt.Add(-time.Hour)},
		{ID: "id-3", CreatedAt: createdAt.Add(-2 * time.Hour)},
	}

	t.Run("còn trang sau", func(t *testing.T) {
		t.Parallel()

		repo := &fakeRepo{t: t, listFn: func(context.Context, course.ListFilter) ([]course.Course, error) {
			return rows, nil
		}}

		page, err := course.NewService(repo).List(context.Background(), course.ListFilter{PageSize: 2})
		if err != nil {
			t.Fatalf("List() returned error: %v", err)
		}
		if len(page.Items) != 2 {
			t.Fatalf("len(items) = %d, want 2 (hàng dư chỉ để dò trang sau)", len(page.Items))
		}
		if page.NextCursor == nil {
			t.Fatal("NextCursor = nil, want the cursor of the last returned row")
		}
		if page.NextCursor.ID != "id-2" || !page.NextCursor.CreatedAt.Equal(rows[1].CreatedAt) {
			t.Errorf("NextCursor = %+v, want id-2 at %v", page.NextCursor, rows[1].CreatedAt)
		}
	})

	t.Run("hết dữ liệu", func(t *testing.T) {
		t.Parallel()

		repo := &fakeRepo{t: t, listFn: func(context.Context, course.ListFilter) ([]course.Course, error) {
			return rows[:2], nil
		}}

		page, err := course.NewService(repo).List(context.Background(), course.ListFilter{PageSize: 5})
		if err != nil {
			t.Fatalf("List() returned error: %v", err)
		}
		if len(page.Items) != 2 {
			t.Fatalf("len(items) = %d, want 2", len(page.Items))
		}
		if page.NextCursor != nil {
			t.Errorf("NextCursor = %+v, want nil", page.NextCursor)
		}
	})
}

func TestListRejectsInvalidFilters(t *testing.T) {
	t.Parallel()

	_, err := course.NewService(&fakeRepo{t: t}).List(context.Background(), course.ListFilter{
		Status: ptr(course.Status("live")),
		Level:  ptr(course.Level("Z9")),
	})

	fields := fieldsOf(t, err)
	if _, ok := fields["status"]; !ok {
		t.Errorf("fields = %v, want an entry for status", fields)
	}
	if _, ok := fields["level"]; !ok {
		t.Errorf("fields = %v, want an entry for level", fields)
	}
}

func TestUpdateRejectsEmptyAndContradictoryInput(t *testing.T) {
	t.Parallel()

	t.Run("không có field nào", func(t *testing.T) {
		t.Parallel()

		_, err := course.NewService(&fakeRepo{t: t}).Update(context.Background(), "id-1", course.UpdateParams{})
		if _, ok := fieldsOf(t, err)["body"]; !ok {
			t.Errorf("want a validation entry for body")
		}
	})

	t.Run("vừa đặt vừa xoá ảnh bìa", func(t *testing.T) {
		t.Parallel()

		_, err := course.NewService(&fakeRepo{t: t}).Update(context.Background(), "id-1", course.UpdateParams{
			CoverImageURL:   new("https://cdn.example.com/a.png"),
			ClearCoverImage: true,
		})
		if _, ok := fieldsOf(t, err)["cover_image_url"]; !ok {
			t.Errorf("want a validation entry for cover_image_url")
		}
	})
}

func TestUpdateNormalisesSlug(t *testing.T) {
	t.Parallel()

	var got course.UpdateParams
	repo := &fakeRepo{t: t, updateFn: func(_ context.Context, _ string, params course.UpdateParams) (course.Course, error) {
		got = params
		return course.Course{ID: "id-1"}, nil
	}}

	if _, err := course.NewService(repo).Update(context.Background(), "id-1", course.UpdateParams{
		Slug: new("  Ngu-Phap  "),
	}); err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}

	if got.Slug == nil || *got.Slug != "ngu-phap" {
		t.Errorf("slug = %v, want %q", got.Slug, "ngu-phap")
	}
}

func TestDeletePropagatesNotFound(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{t: t, softDeleteFn: func(context.Context, string) error {
		return course.ErrNotFound
	}}

	if err := course.NewService(repo).Delete(context.Background(), "id-1"); !errors.Is(err, course.ErrNotFound) {
		t.Fatalf("error = %v, want it to match ErrNotFound", err)
	}
}

func TestListLessonsRequiresAnExistingCourse(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{t: t, existsFn: func(context.Context, string) (bool, error) {
		return false, nil
	}}

	_, err := course.NewService(repo).ListLessons(context.Background(), "id-1")
	if !errors.Is(err, course.ErrNotFound) {
		t.Fatalf("error = %v, want it to match ErrNotFound", err)
	}
	if repo.listLessonsCalled {
		t.Error("ListLessons was called on the repo although the course does not exist")
	}
}

func TestListLessonsReturnsEmptySliceForCourseWithoutLessons(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{
		t:        t,
		existsFn: func(context.Context, string) (bool, error) { return true, nil },
		listLessonsFn: func(context.Context, string) ([]course.Lesson, error) {
			return []course.Lesson{}, nil
		},
	}

	lessons, err := course.NewService(repo).ListLessons(context.Background(), "id-1")
	if err != nil {
		t.Fatalf("ListLessons() returned error: %v", err)
	}
	if len(lessons) != 0 {
		t.Fatalf("len(lessons) = %d, want 0", len(lessons))
	}
}

func TestGetBySlugNormalisesInput(t *testing.T) {
	t.Parallel()

	var got string
	repo := &fakeRepo{t: t, getBySlugFn: func(_ context.Context, slug string) (course.Course, error) {
		got = slug
		return course.Course{ID: "id-1"}, nil
	}}

	if _, err := course.NewService(repo).GetBySlug(context.Background(), "  Ngu-Phap  "); err != nil {
		t.Fatalf("GetBySlug() returned error: %v", err)
	}
	if got != "ngu-phap" {
		t.Errorf("repo received slug %q, want %q", got, "ngu-phap")
	}
}

func TestCreateEnforcesLengthLimits(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		params    course.CreateParams
		wantField string
	}{
		"slug quá dài": {
			params: course.CreateParams{
				Slug: strings.Repeat("a", 121), Title: "Tiêu đề", Level: course.LevelA1,
			},
			wantField: "slug",
		},
		"title quá dài": {
			params: course.CreateParams{
				Slug: "grammar", Title: strings.Repeat("á", 201), Level: course.LevelA1,
			},
			wantField: "title",
		},
		"description quá dài": {
			params: course.CreateParams{
				Slug: "grammar", Title: "Tiêu đề", Level: course.LevelA1,
				Description: strings.Repeat("á", 5001),
			},
			wantField: "description",
		},
		"cover image quá dài": {
			params: course.CreateParams{
				Slug: "grammar", Title: "Tiêu đề", Level: course.LevelA1,
				CoverImageURL: ptr("https://cdn.example.com/" + strings.Repeat("a", 2048)),
			},
			wantField: "cover_image_url",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := course.NewService(&fakeRepo{t: t}).Create(context.Background(), tc.params)
			if _, ok := fieldsOf(t, err)[tc.wantField]; !ok {
				t.Errorf("want a validation entry for %q", tc.wantField)
			}
		})
	}
}

func TestCreateAcceptsMaximumLengths(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{t: t, createFn: func(context.Context, course.CreateParams) (course.Course, error) {
		return course.Course{ID: "id-1"}, nil
	}}

	// Đúng ngưỡng vẫn phải qua: giới hạn là "tối đa", không phải "nhỏ hơn".
	_, err := course.NewService(repo).Create(context.Background(), course.CreateParams{
		Slug:        strings.Repeat("a", 120),
		Title:       strings.Repeat("á", 200),
		Description: strings.Repeat("á", 5000),
		Level:       course.LevelC2,
	})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
}

func TestValidationErrorMessageListsFieldsInOrder(t *testing.T) {
	t.Parallel()

	err := &course.ValidationError{Fields: map[string]string{
		"title": "không được rỗng",
		"slug":  "không được rỗng",
	}}

	const want = "course: validation failed (slug: không được rỗng; title: không được rỗng)"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func (f *fakeRepo) CreateLesson(ctx context.Context, courseID string, params course.LessonCreateParams) (course.Lesson, error) {
	if f.createLessonFn == nil {
		f.t.Fatal("CreateLesson called unexpectedly")
	}
	return f.createLessonFn(ctx, courseID, params)
}

func (f *fakeRepo) GetLesson(ctx context.Context, lessonID string) (course.Lesson, error) {
	if f.getLessonFn == nil {
		f.t.Fatal("GetLesson called unexpectedly")
	}
	return f.getLessonFn(ctx, lessonID)
}

func (f *fakeRepo) UpdateLesson(ctx context.Context, lessonID string, params course.LessonUpdateParams) (course.Lesson, error) {
	if f.updateLessonFn == nil {
		f.t.Fatal("UpdateLesson called unexpectedly")
	}
	return f.updateLessonFn(ctx, lessonID, params)
}

func (f *fakeRepo) SoftDeleteLesson(ctx context.Context, lessonID string) error {
	if f.softDeleteLessonFn == nil {
		f.t.Fatal("SoftDeleteLesson called unexpectedly")
	}
	return f.softDeleteLessonFn(ctx, lessonID)
}

func (f *fakeRepo) ListLessonIDs(ctx context.Context, courseID string) ([]string, error) {
	if f.listLessonIDsFn == nil {
		f.t.Fatal("ListLessonIDs called unexpectedly")
	}
	return f.listLessonIDsFn(ctx, courseID)
}

func (f *fakeRepo) ReorderLessons(ctx context.Context, courseID string, lessonIDs []string) (int64, error) {
	f.reorderLessonsArgs = lessonIDs
	if f.reorderLessonsFn == nil {
		f.t.Fatal("ReorderLessons called unexpectedly")
	}
	return f.reorderLessonsFn(ctx, courseID, lessonIDs)
}
