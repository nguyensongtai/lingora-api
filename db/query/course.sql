-- name: CreateCourse :one
INSERT INTO courses (slug, title, description, level, status, cover_image_url)
VALUES (
    sqlc.arg('slug'),
    sqlc.arg('title'),
    sqlc.arg('description'),
    sqlc.arg('level'),
    sqlc.arg('status'),
    sqlc.narg('cover_image_url')
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
UPDATE courses
SET
    slug            = COALESCE(sqlc.narg('slug'), slug),
    title           = COALESCE(sqlc.narg('title'), title),
    description     = COALESCE(sqlc.narg('description'), description),
    level           = COALESCE(sqlc.narg('level'), level),
    status          = COALESCE(sqlc.narg('status'), status),
    cover_image_url = CASE
        WHEN sqlc.arg('clear_cover_image')::boolean THEN NULL
        ELSE COALESCE(sqlc.narg('cover_image_url'), cover_image_url)
    END,
    updated_at      = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteCourse :execrows
UPDATE courses
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: CourseExists :one
SELECT EXISTS (
    SELECT 1 FROM courses WHERE id = sqlc.arg('id') AND deleted_at IS NULL
);
