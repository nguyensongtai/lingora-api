package course_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nguyensongtai/lingora-api/internal/course"
	"github.com/nguyensongtai/lingora-api/internal/testdb"
)

// newRepo dựng repo trên một transaction sẽ được rollback, và trả kèm chính
// transaction đó để test gọi được testdb.Attempt.
func newRepo(t *testing.T) (*course.Repo, pgx.Tx) {
	t.Helper()

	tx := testdb.New(t)
	return course.NewRepo(tx), tx
}

// attemptCreate thử một thao tác sẽ vi phạm ràng buộc, trong savepoint riêng.
func attemptCreate(t *testing.T, tx pgx.Tx, fn func(*course.Repo) error) error {
	t.Helper()

	return testdb.Attempt(t, tx, func(nested pgx.Tx) error {
		return fn(course.NewRepo(nested))
	})
}

func seedCourse(t *testing.T, repo *course.Repo, slug string) course.Course {
	t.Helper()

	created, err := repo.Create(context.Background(), course.CreateParams{
		Slug:   slug,
		Title:  "Khoá " + slug,
		Level:  course.LevelA1,
		Status: course.StatusPublished,
	})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	return created
}

func TestRepoSlugIsFreedBySoftDelete(t *testing.T) {
	t.Parallel()

	repo, tx := newRepo(t)
	ctx := context.Background()
	first := seedCourse(t, repo, "khoa-thu-nghiem-slug")

	// Unique index có WHERE deleted_at IS NULL, nên slug phải được giải phóng
	// sau khi xoá mềm — đây là điều unit test với fake repo không kiểm được.
	err := attemptCreate(t, tx, func(r *course.Repo) error {
		_, createErr := r.Create(ctx, course.CreateParams{
			Slug: "khoa-thu-nghiem-slug", Title: "Trùng", Level: course.LevelA1, Status: course.StatusDraft,
		})
		return createErr
	})
	if !errors.Is(err, course.ErrSlugTaken) {
		t.Fatalf("tạo trùng slug: error = %v, want ErrSlugTaken", err)
	}

	if err := repo.SoftDelete(ctx, first.ID); err != nil {
		t.Fatalf("SoftDelete() returned error: %v", err)
	}
	if _, err := repo.Create(ctx, course.CreateParams{
		Slug: "khoa-thu-nghiem-slug", Title: "Dùng lại", Level: course.LevelA1, Status: course.StatusDraft,
	}); err != nil {
		t.Fatalf("tạo lại sau khi xoá mềm: %v", err)
	}

	if _, err := repo.GetByID(ctx, first.ID); !errors.Is(err, course.ErrNotFound) {
		t.Errorf("GetByID() của khoá đã xoá: error = %v, want ErrNotFound", err)
	}
	if err := repo.SoftDelete(ctx, first.ID); !errors.Is(err, course.ErrNotFound) {
		t.Errorf("xoá lần hai: error = %v, want ErrNotFound", err)
	}
}

func TestRepoUpdateKeepsUntouchedFields(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)
	ctx := context.Background()
	created, err := repo.Create(ctx, course.CreateParams{
		Slug: "khoa-cap-nhat-tung-phan", Title: "Tiêu đề gốc", Description: "Mô tả gốc",
		Level: course.LevelB1, Status: course.StatusDraft, CoverImageURL: strptr("https://cdn/a.png"),
	})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	title := "Tiêu đề mới"
	updated, err := repo.Update(ctx, created.ID, course.UpdateParams{Title: &title})
	if err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}

	// COALESCE trong câu UPDATE: field nil phải giữ nguyên giá trị cũ.
	if updated.Title != title {
		t.Errorf("Title = %q, want %q", updated.Title, title)
	}
	if updated.Description != "Mô tả gốc" || updated.Level != course.LevelB1 {
		t.Errorf("update một field lại đổi field khác: %+v", updated)
	}
	if updated.CoverImageURL == nil || *updated.CoverImageURL != "https://cdn/a.png" {
		t.Errorf("CoverImageURL = %v, want giữ nguyên", updated.CoverImageURL)
	}

	// ClearCoverImage là đường riêng để xoá ảnh, vì nil đã mang nghĩa "giữ nguyên".
	cleared, err := repo.Update(ctx, created.ID, course.UpdateParams{ClearCoverImage: true})
	if err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}
	if cleared.CoverImageURL != nil {
		t.Errorf("CoverImageURL = %v, want nil", cleared.CoverImageURL)
	}
}

func TestRepoListPagesByKeysetWhenTimestampsTie(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)
	ctx := context.Background()

	// Cả ba khoá tạo trong cùng một transaction nên created_at bằng nhau: đây
	// đúng là trường hợp cần tới id trong khoá sắp xếp.
	for _, slug := range []string{"keyset-mot", "keyset-hai", "keyset-ba"} {
		seedCourse(t, repo, slug)
	}

	page, err := repo.List(ctx, course.ListFilter{PageSize: 2})
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if len(page) < 2 {
		t.Fatalf("trang đầu có %d khoá, want ít nhất 2", len(page))
	}

	last := page[len(page)-1]
	next, err := repo.List(ctx, course.ListFilter{
		PageSize: 2,
		Cursor:   &course.Cursor{CreatedAt: last.CreatedAt, ID: last.ID},
	})
	if err != nil {
		t.Fatalf("List() trang sau returned error: %v", err)
	}

	// Không có bản ghi nào lặp lại giữa hai trang dù created_at trùng nhau.
	seen := map[string]bool{}
	for _, item := range page {
		seen[item.ID] = true
	}
	for _, item := range next {
		if seen[item.ID] {
			t.Errorf("khoá %s (%s) xuất hiện ở cả hai trang", item.Slug, item.ID)
		}
	}
}

func TestRepoReorderLessonsRenumbersWithoutGaps(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)
	ctx := context.Background()
	owner := seedCourse(t, repo, "khoa-sap-xep-bai")

	var ids []string
	for _, slug := range []string{"bai-mot", "bai-hai", "bai-ba"} {
		lesson, err := repo.CreateLesson(ctx, owner.ID, course.LessonCreateParams{Slug: slug, Title: slug})
		if err != nil {
			t.Fatalf("CreateLesson() returned error: %v", err)
		}
		ids = append(ids, lesson.ID)
	}

	reversed := []string{ids[2], ids[1], ids[0]}
	affected, err := repo.ReorderLessons(ctx, owner.ID, reversed)
	if err != nil {
		t.Fatalf("ReorderLessons() returned error: %v", err)
	}
	if affected != 3 {
		t.Errorf("đổi %d hàng, want 3", affected)
	}

	lessons, err := repo.ListLessons(ctx, owner.ID)
	if err != nil {
		t.Fatalf("ListLessons() returned error: %v", err)
	}
	// unnest WITH ORDINALITY phải cho position liền mạch 0,1,2 theo đúng thứ tự gửi lên.
	for index, lesson := range lessons {
		if lesson.ID != reversed[index] {
			t.Errorf("vị trí %d là %s, want %s", index, lesson.ID, reversed[index])
		}
		if lesson.Position != int32(index) {
			t.Errorf("bài %s có position %d, want %d", lesson.Slug, lesson.Position, index)
		}
	}
}

func TestRepoReorderIgnoresLessonsOfAnotherCourse(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)
	ctx := context.Background()
	mine := seedCourse(t, repo, "khoa-cua-toi")
	theirs := seedCourse(t, repo, "khoa-cua-nguoi-khac")

	lesson, err := repo.CreateLesson(ctx, theirs.ID, course.LessonCreateParams{Slug: "bai-la", Title: "Bài lạ"})
	if err != nil {
		t.Fatalf("CreateLesson() returned error: %v", err)
	}

	// Câu UPDATE có điều kiện course_id, nên id của khoá khác không đổi được gì.
	affected, err := repo.ReorderLessons(ctx, mine.ID, []string{lesson.ID})
	if err != nil {
		t.Fatalf("ReorderLessons() returned error: %v", err)
	}
	if affected != 0 {
		t.Errorf("đổi %d hàng, want 0", affected)
	}
}

func TestRepoLessonSlugIsUniquePerCourse(t *testing.T) {
	t.Parallel()

	repo, tx := newRepo(t)
	ctx := context.Background()
	first := seedCourse(t, repo, "khoa-slug-bai-mot")
	second := seedCourse(t, repo, "khoa-slug-bai-hai")

	params := course.LessonCreateParams{Slug: "bai-trung-ten", Title: "Bài"}
	if _, err := repo.CreateLesson(ctx, first.ID, params); err != nil {
		t.Fatalf("CreateLesson() returned error: %v", err)
	}
	duplicate := attemptCreate(t, tx, func(r *course.Repo) error {
		_, err := r.CreateLesson(ctx, first.ID, params)
		return err
	})
	if !errors.Is(duplicate, course.ErrLessonSlugTaken) {
		t.Errorf("trùng slug trong cùng khoá: error = %v, want ErrLessonSlugTaken", duplicate)
	}
	// Unique index là (course_id, slug), nên khoá khác dùng lại được.
	if _, err := repo.CreateLesson(ctx, second.ID, params); err != nil {
		t.Errorf("slug đó ở khoá khác: %v", err)
	}
}

func strptr(value string) *string { return &value }

func TestRepoCoursePositionsAreScopedToTheLevel(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)
	ctx := context.Background()

	create := func(slug string, level course.Level) course.Course {
		t.Helper()

		created, err := repo.Create(ctx, course.CreateParams{
			Slug: slug, Title: slug, Level: level, Status: course.StatusDraft,
		})
		if err != nil {
			t.Fatalf("Create() returned error: %v", err)
		}
		return created
	}

	// Mọi bậc đều có thể đã có khoá sẵn trong database (seed dev nạp đủ cả
	// sáu), nên chỉ so tương đối chứ không so với một con số tuyệt đối. Bản cũ
	// giả định C2 trống và đỏ ngay khi chạy trên database đã seed.
	//
	// Một khoá A1 chen giữa hai khoá C2 không được đẩy bộ đếm của C2: đó chính
	// là "bộ đếm tính riêng cho từng bậc".
	firstC2 := create("vi-tri-c2-mot", course.LevelC2)
	firstA1 := create("vi-tri-a1-mot", course.LevelA1)
	secondC2 := create("vi-tri-c2-hai", course.LevelC2)
	secondA1 := create("vi-tri-a1-hai", course.LevelA1)

	if secondC2.Position != firstC2.Position+1 {
		t.Errorf("khoá thứ hai của C2 có position %d, want %d", secondC2.Position, firstC2.Position+1)
	}
	if secondA1.Position != firstA1.Position+1 {
		t.Errorf("khoá thứ hai của A1 có position %d, want %d", secondA1.Position, firstA1.Position+1)
	}
}

func TestRepoChangingLevelMovesTheCourseToTheEnd(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)
	ctx := context.Background()

	moved, err := repo.Create(ctx, course.CreateParams{
		Slug: "khoa-doi-bac", Title: "Khoá đổi bậc", Level: course.LevelC1, Status: course.StatusDraft,
	})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	neighbour, err := repo.Create(ctx, course.CreateParams{
		Slug: "khoa-o-bac-dich", Title: "Khoá ở bậc đích", Level: course.LevelC2, Status: course.StatusDraft,
	})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	level := course.LevelC2
	updated, err := repo.Update(ctx, moved.ID, course.UpdateParams{Level: &level})
	if err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}

	// position cũ là thứ tự trong bậc cũ; mang sang bậc khác thì vô nghĩa và dễ
	// trùng, nên khoá phải về cuối bậc mới.
	if updated.Position != neighbour.Position+1 {
		t.Errorf("position sau khi đổi bậc = %d, want %d", updated.Position, neighbour.Position+1)
	}

	// Sửa field khác thì position đứng yên.
	title := "Tiêu đề khác"
	again, err := repo.Update(ctx, moved.ID, course.UpdateParams{Title: &title})
	if err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}
	if again.Position != updated.Position {
		t.Errorf("sửa tiêu đề lại đổi position: %d -> %d", updated.Position, again.Position)
	}
}

func TestRepoReorderCoursesIgnoresOtherLevels(t *testing.T) {
	t.Parallel()

	repo, _ := newRepo(t)
	ctx := context.Background()

	outsider, err := repo.Create(ctx, course.CreateParams{
		Slug: "khoa-bac-khac", Title: "Khoá bậc khác", Level: course.LevelC1, Status: course.StatusDraft,
	})
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	// Điều kiện level trong câu UPDATE: id của bậc khác không đổi được gì.
	affected, err := repo.ReorderCourses(ctx, course.LevelC2, []string{outsider.ID})
	if err != nil {
		t.Fatalf("ReorderCourses() returned error: %v", err)
	}
	if affected != 0 {
		t.Errorf("đổi %d hàng, want 0", affected)
	}
}
