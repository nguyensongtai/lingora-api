-- name: ListLessonsByCourse :many
SELECT * FROM lessons
WHERE course_id = sqlc.arg('course_id') AND deleted_at IS NULL
ORDER BY position ASC, id ASC;
