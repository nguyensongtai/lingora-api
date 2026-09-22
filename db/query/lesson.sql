-- name: ListLessonsByCourse :many
SELECT * FROM lessons
WHERE course_id = sqlc.arg('course_id') AND deleted_at IS NULL
ORDER BY position ASC, id ASC;

-- name: ListLessonIDsByCourse :many
SELECT id FROM lessons
WHERE course_id = sqlc.arg('course_id') AND deleted_at IS NULL
ORDER BY position ASC, id ASC;

-- name: CreateLesson :one
-- Bài mới luôn đứng cuối khoá; thứ tự chỉ đổi được qua ReorderLessons.
INSERT INTO lessons (course_id, slug, title, position)
VALUES (
    sqlc.arg('course_id'),
    sqlc.arg('slug'),
    sqlc.arg('title'),
    COALESCE(
        (
            SELECT MAX(position) + 1 FROM lessons
            WHERE course_id = sqlc.arg('course_id') AND deleted_at IS NULL
        ),
        0
    )
)
RETURNING *;

-- name: GetLessonByID :one
SELECT * FROM lessons
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: UpdateLesson :one
-- position cố ý không nằm ở đây: đổi thứ tự là việc của ReorderLessons.
UPDATE lessons
SET
    slug       = COALESCE(sqlc.narg('slug'), slug),
    title      = COALESCE(sqlc.narg('title'), title),
    summary    = COALESCE(sqlc.narg('summary'), summary),
    updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteLesson :execrows
UPDATE lessons
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: ReorderLessons :execrows
-- Vị trí mới lấy từ chính thứ tự của mảng id truyền vào, nên toàn bộ khoá được
-- đánh lại số trong một câu lệnh, không có khoảng thời gian nào thứ tự bị lệch.
UPDATE lessons AS l
SET position = new_order.position, updated_at = now()
FROM (
    SELECT id, (ordinality - 1)::integer AS position
    FROM unnest(sqlc.arg('lesson_ids')::uuid[]) WITH ORDINALITY AS t(id, ordinality)
) AS new_order
WHERE l.id = new_order.id
  AND l.course_id = sqlc.arg('course_id')
  AND l.deleted_at IS NULL;

-- name: ListLessonBlocks :many
SELECT * FROM lesson_blocks
WHERE lesson_id = sqlc.arg('lesson_id')
ORDER BY position ASC, id ASC;

-- name: ReplaceLessonBlocks :execrows
-- Thay toàn bộ nội dung của một bài trong MỘT câu lệnh.
--
-- DELETE nằm trong CTE nên nó chạy tới cùng dù phần INSERT không sinh dòng nào
-- — gửi danh sách rỗng là xoá sạch nội dung bài, đúng như mong đợi. Gộp hai
-- thao tác vào một câu để không có khoảnh khắc nào bài bị trống giữa chừng.
--
-- Tham số là jsonb chứ không phải năm mảng song song: unnest nhiều mảng thì
-- sqlc không phân tích được, còn năm mảng rời thì lệch độ dài là ghi sai mà
-- không ai biết.
--
-- position lấy từ ORDINALITY nên luôn liền mạch từ 0, bất kể client gửi gì.
WITH cleared AS (
    DELETE FROM lesson_blocks WHERE lesson_id = sqlc.arg('lesson_id')
)
INSERT INTO lesson_blocks (lesson_id, kind, body, text_en, text_vi, speaker, position)
SELECT
    sqlc.arg('lesson_id'),
    (t.block->>'kind')::lesson_block_kind,
    coalesce(t.block->>'body', ''),
    coalesce(t.block->>'text_en', ''),
    coalesce(t.block->>'text_vi', ''),
    coalesce(t.block->>'speaker', ''),
    (t.ordinality - 1)::integer
FROM jsonb_array_elements(sqlc.arg('blocks')::jsonb) WITH ORDINALITY AS t(block, ordinality);
