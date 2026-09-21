-- name: CreateCourse :one
-- Khoá mới đứng cuối bậc của nó, giống cách bài mới đứng cuối khoá.
INSERT INTO courses (slug, title, description, level, status, cover_image_url, position)
VALUES (
    sqlc.arg('slug'),
    sqlc.arg('title'),
    sqlc.arg('description'),
    sqlc.arg('level'),
    sqlc.arg('status'),
    sqlc.narg('cover_image_url'),
    (
        SELECT coalesce(max(position) + 1, 0)
        FROM courses AS sibling
        WHERE sibling.level = sqlc.arg('level') AND sibling.deleted_at IS NULL
    )
)
RETURNING *;

-- name: GetCourseByID :one
SELECT * FROM courses
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: GetCourseBySlug :one
SELECT * FROM courses
WHERE slug = sqlc.arg('slug') AND deleted_at IS NULL;

-- name: ListCourses :many
-- Keyset pagination: con trỏ là cặp (created_at, id) của bản ghi cuối trang trước.
SELECT * FROM courses
WHERE deleted_at IS NULL
  AND (sqlc.narg('status')::course_status IS NULL OR status = sqlc.narg('status')::course_status)
  AND (sqlc.narg('level')::course_level IS NULL OR level = sqlc.narg('level')::course_level)
  AND (
      sqlc.narg('cursor_created_at')::timestamptz IS NULL
      OR (created_at, id) < (sqlc.narg('cursor_created_at')::timestamptz, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('page_size');

-- name: UpdateCourse :one
-- Mọi tham số đều optional: NULL nghĩa là giữ nguyên giá trị cũ.
UPDATE courses AS c
SET
    slug            = COALESCE(sqlc.narg('slug'), c.slug),
    title           = COALESCE(sqlc.narg('title'), c.title),
    description     = COALESCE(sqlc.narg('description'), c.description),
    level           = COALESCE(sqlc.narg('level'), c.level),
    status          = COALESCE(sqlc.narg('status'), c.status),
    cover_image_url = CASE
        WHEN sqlc.arg('clear_cover_image')::boolean THEN NULL
        ELSE COALESCE(sqlc.narg('cover_image_url'), c.cover_image_url)
    END,
    -- Đổi bậc thì khoá về cuối bậc mới: position cũ là thứ tự trong bậc cũ,
    -- mang sang bậc khác thì vô nghĩa và dễ trùng với khoá đang có ở đó.
    position        = CASE
        WHEN sqlc.narg('level') IS NOT NULL AND sqlc.narg('level') <> c.level THEN (
            SELECT coalesce(max(position) + 1, 0)
            FROM courses AS sibling
            WHERE sibling.level = sqlc.narg('level') AND sibling.deleted_at IS NULL
        )
        ELSE c.position
    END,
    updated_at      = now()
WHERE c.id = sqlc.arg('id') AND c.deleted_at IS NULL
RETURNING c.*;

-- name: SoftDeleteCourse :execrows
UPDATE courses
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: CourseExists :one
SELECT EXISTS (
    SELECT 1 FROM courses WHERE id = sqlc.arg('id') AND deleted_at IS NULL
);

-- name: ListCourseIDsByLevel :many
SELECT id FROM courses
WHERE level = sqlc.arg('level') AND deleted_at IS NULL
ORDER BY position, created_at, id;

-- name: ReorderCourses :execrows
-- Đánh lại position theo đúng thứ tự mảng gửi lên, trong một câu. Điều kiện
-- level giữ cho id của bậc khác không đổi được gì.
UPDATE courses AS c
SET position = new_order.position, updated_at = now()
FROM (
    SELECT id, (ordinality - 1)::integer AS position
    FROM unnest(sqlc.arg('course_ids')::uuid[]) WITH ORDINALITY AS t(id, ordinality)
) AS new_order
WHERE c.id = new_order.id
  AND c.level = sqlc.arg('level')
  AND c.deleted_at IS NULL;
