package course_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nguyensongtai/lingora-api/internal/course"
)

func existingCourse(_ context.Context, _ string) (bool, error) { return true, nil }

func TestCreateLessonNormalisesAndRequiresTheCourse(t *testing.T) {
	t.Parallel()

	t.Run("chuẩn hoá slug và title", func(t *testing.T) {
		t.Parallel()

		var got course.LessonCreateParams
		repo := &fakeRepo{
			t:        t,
			existsFn: existingCourse,
			createLessonFn: func(_ context.Context, _ string, params course.LessonCreateParams) (course.Lesson, error) {
				got = params
				return course.Lesson{ID: "lesson-1"}, nil
			},
		}

		if _, err := course.NewService(repo).CreateLesson(context.Background(), "course-1", course.LessonCreateParams{
			Slug:  "  Thi-Hien-Tai  ",
			Title: "  Thì hiện tại  ",
		}); err != nil {
			t.Fatalf("CreateLesson() returned error: %v", err)
		}

		if got.Slug != "thi-hien-tai" {
			t.Errorf("slug = %q, want %q", got.Slug, "thi-hien-tai")
		}
		if got.Title != "Thì hiện tại" {
			t.Errorf("title = %q, want it trimmed", got.Title)
		}
	})

	t.Run("khoá không tồn tại", func(t *testing.T) {
		t.Parallel()

		// createLessonFn để nil: repo không được chạm tới khi khoá không có thật.
		repo := &fakeRepo{t: t, existsFn: func(context.Context, string) (bool, error) {
			return false, nil
		}}

		_, err := course.NewService(repo).CreateLesson(context.Background(), "course-1", course.LessonCreateParams{
			Slug:  "bai-1",
			Title: "Bài 1",
		})
		if !errors.Is(err, course.ErrNotFound) {
			t.Fatalf("error = %v, want it to match ErrNotFound", err)
		}
	})

	t.Run("dữ liệu sai thì không hỏi tới khoá", func(t *testing.T) {
		t.Parallel()

		// existsFn cũng để nil: validate phải chặn trước mọi lời gọi repo.
		_, err := course.NewService(&fakeRepo{t: t}).CreateLesson(context.Background(), "course-1", course.LessonCreateParams{
			Slug: "SLUG SAI",
		})

		fields := fieldsOf(t, err)
		if _, ok := fields["slug"]; !ok {
			t.Errorf("fields = %v, want an entry for slug", fields)
		}
		if _, ok := fields["title"]; !ok {
			t.Errorf("fields = %v, want an entry for title", fields)
		}
	})
}

func TestUpdateLessonRejectsEmptyBody(t *testing.T) {
	t.Parallel()

	_, err := course.NewService(&fakeRepo{t: t}).UpdateLesson(context.Background(), "lesson-1", course.LessonUpdateParams{})
	if _, ok := fieldsOf(t, err)["body"]; !ok {
		t.Error("want a validation entry for body")
	}
}

func TestUpdateLessonPropagatesSlugTaken(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{t: t, updateLessonFn: func(context.Context, string, course.LessonUpdateParams) (course.Lesson, error) {
		return course.Lesson{}, course.ErrLessonSlugTaken
	}}

	_, err := course.NewService(repo).UpdateLesson(context.Background(), "lesson-1", course.LessonUpdateParams{
		Slug: ptr("bai-1"),
	})
	if !errors.Is(err, course.ErrLessonSlugTaken) {
		t.Fatalf("error = %v, want it to match ErrLessonSlugTaken", err)
	}
}

func TestDeleteLessonPropagatesNotFound(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{t: t, softDeleteLessonFn: func(context.Context, string) error {
		return course.ErrLessonNotFound
	}}

	if err := course.NewService(repo).DeleteLesson(context.Background(), "lesson-1"); !errors.Is(err, course.ErrLessonNotFound) {
		t.Fatalf("error = %v, want it to match ErrLessonNotFound", err)
	}
}

func TestReorderLessonsAcceptsExactlyTheCurrentSet(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{
		t:        t,
		existsFn: existingCourse,
		listLessonIDsFn: func(context.Context, string) ([]string, error) {
			return []string{"a", "b", "c"}, nil
		},
		reorderLessonsFn: func(context.Context, string, []string) (int64, error) { return 3, nil },
	}

	if err := course.NewService(repo).ReorderLessons(context.Background(), "course-1", []string{"c", "a", "b"}); err != nil {
		t.Fatalf("ReorderLessons() returned error: %v", err)
	}
	if got := repo.reorderLessonsArgs; len(got) != 3 || got[0] != "c" {
		t.Errorf("repo received %v, want the requested order preserved", got)
	}
}

func TestReorderLessonsRejectsAnIncompleteSet(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		current   []string
		requested []string
	}{
		"thiếu bài":      {[]string{"a", "b", "c"}, []string{"a", "b"}},
		"thừa bài":       {[]string{"a", "b"}, []string{"a", "b", "c"}},
		"id bị lặp":      {[]string{"a", "b"}, []string{"a", "a"}},
		"đổi mất một id": {[]string{"a", "b"}, []string{"a", "z"}},
		"danh sách rỗng": {[]string{"a"}, []string{}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// reorderLessonsFn để nil: repo không được chạm tới khi tập bài sai.
			repo := &fakeRepo{
				t:        t,
				existsFn: existingCourse,
				listLessonIDsFn: func(context.Context, string) ([]string, error) {
					return tc.current, nil
				},
			}

			err := course.NewService(repo).ReorderLessons(context.Background(), "course-1", tc.requested)
			if _, ok := fieldsOf(t, err)["lesson_ids"]; !ok {
				t.Errorf("want a validation entry for lesson_ids, got %v", err)
			}
		})
	}
}

func TestReorderLessonsRequiresTheCourse(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{t: t, existsFn: func(context.Context, string) (bool, error) { return false, nil }}

	err := course.NewService(repo).ReorderLessons(context.Background(), "course-1", []string{"a"})
	if !errors.Is(err, course.ErrNotFound) {
		t.Fatalf("error = %v, want it to match ErrNotFound", err)
	}
}

/* ---------- nội dung bài học ---------- */

func blockService(t *testing.T, repo *fakeRepo) *course.Service {
	t.Helper()

	if repo.getLessonFn == nil {
		repo.getLessonFn = func(context.Context, string) (course.Lesson, error) {
			return course.Lesson{ID: "lesson-1"}, nil
		}
	}
	return course.NewService(repo)
}

func replace(t *testing.T, blocks []course.Block) error {
	t.Helper()

	return blockService(t, &fakeRepo{t: t}).
		ReplaceBlocks(context.Background(), "lesson-1", blocks)
}

/**
 * Mỗi dạng khối chỉ dùng phần trường của nó. Database có ràng buộc bắt đúng
 * hình dạng này; service kiểm lại để lỗi ra 400 kèm tên field thay vì 500 kèm
 * một câu SQLSTATE.
 */
func TestReplaceBlocksRejectsTheWrongShape(t *testing.T) {
	t.Parallel()

	cases := map[string]course.Block{
		"note rỗng body":       {Kind: course.BlockNote},
		"note lẫn text_en":     {Kind: course.BlockNote, Body: "giải thích", TextEN: "Hello."},
		"example rỗng text_en": {Kind: course.BlockExample, TextVI: "Xin chào."},
		"example lẫn speaker":  {Kind: course.BlockExample, TextEN: "Hello.", Speaker: "An"},
		"dialogue thiếu người": {Kind: course.BlockDialogue, TextEN: "Hello."},
		"dialogue lẫn body":    {Kind: course.BlockDialogue, TextEN: "Hello.", Speaker: "An", Body: "x"},
		"dạng không có thật":   {Kind: course.BlockKind("bai-hat")},
	}

	for name, block := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := replace(t, []course.Block{block})

			if !errors.Is(err, course.ErrValidation) {
				t.Fatalf("error = %v, want it to match ErrValidation", err)
			}
		})
	}
}

func TestReplaceBlocksAcceptsEachKind(t *testing.T) {
	t.Parallel()

	blocks := []course.Block{
		{Kind: course.BlockNote, Body: "Thì hiện tại đơn dùng cho thói quen."},
		{Kind: course.BlockExample, TextEN: "I walk to school.", TextVI: "Tôi đi bộ đến trường."},
		{Kind: course.BlockDialogue, Speaker: "An", TextEN: "Where do you live?", TextVI: "Bạn sống ở đâu?"},
	}

	if err := replace(t, blocks); err != nil {
		t.Fatalf("ReplaceBlocks() returned error: %v", err)
	}
}

// Gửi mảng rỗng là cách duy nhất để xoá sạch nội dung bài, nên nó phải đi được
// tới repo chứ không bị chặn như dữ liệu thiếu.
func TestReplaceBlocksAcceptsAnEmptyListAsClearing(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{t: t}

	if err := blockService(t, repo).ReplaceBlocks(context.Background(), "lesson-1", nil); err != nil {
		t.Fatalf("ReplaceBlocks() returned error: %v", err)
	}
	if repo.replacedFor != "lesson-1" {
		t.Errorf("repo nhận lesson %q, want lesson-1", repo.replacedFor)
	}
	if len(repo.replacedBlocks) != 0 {
		t.Errorf("repo nhận %d khối, want 0", len(repo.replacedBlocks))
	}
}

func TestReplaceBlocksTrimsEveryField(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{t: t}
	blocks := []course.Block{
		{Kind: course.BlockDialogue, Speaker: "  An  ", TextEN: "  Hello.  ", TextVI: "  Xin chào.  "},
	}

	if err := blockService(t, repo).ReplaceBlocks(context.Background(), "lesson-1", blocks); err != nil {
		t.Fatalf("ReplaceBlocks() returned error: %v", err)
	}

	got := repo.replacedBlocks[0]
	if got.Speaker != "An" || got.TextEN != "Hello." || got.TextVI != "Xin chào." {
		t.Errorf("khối chưa được cắt khoảng trắng: %+v", got)
	}
}

// Ghi vào một bài không tồn tại phải là 404, không phải âm thầm thành công với
// 0 dòng.
func TestReplaceBlocksRequiresAnExistingLesson(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{t: t, getLessonFn: func(context.Context, string) (course.Lesson, error) {
		return course.Lesson{}, course.ErrLessonNotFound
	}}

	err := course.NewService(repo).ReplaceBlocks(context.Background(), "khong-co", nil)

	if !errors.Is(err, course.ErrLessonNotFound) {
		t.Fatalf("error = %v, want it to match ErrLessonNotFound", err)
	}
	if repo.replacedFor != "" {
		t.Error("repo bị gọi dù bài không tồn tại")
	}
}
