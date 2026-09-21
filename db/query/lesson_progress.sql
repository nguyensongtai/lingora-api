-- name: CompleteLesson :exec
-- Idempotent: bấm "Đã xong" lần thứ hai không được dời completed_at, vì đó là
-- mốc người học hoàn thành bài chứ không phải lần bấm gần nhất.
INSERT INTO lesson_progress (user_id, lesson_id)
VALUES (sqlc.arg('user_id'), sqlc.arg('lesson_id'))
ON CONFLICT (user_id, lesson_id) DO NOTHING;

-- name: UncompleteLesson :execrows
DELETE FROM lesson_progress
WHERE user_id = sqlc.arg('user_id')
  AND lesson_id = sqlc.arg('lesson_id');

-- name: ListCompletedLessonIDs :many
-- Bài của khoá đã xoá mềm không còn nằm trong lộ trình nên cũng không kể tới.
SELECT p.lesson_id
FROM lesson_progress AS p
JOIN lessons AS l ON l.id = p.lesson_id
JOIN courses AS c ON c.id = l.course_id
WHERE p.user_id = sqlc.arg('user_id')
  AND l.deleted_at IS NULL
  AND c.deleted_at IS NULL
ORDER BY p.completed_at DESC;

-- name: ListCourseProgress :many
-- Mọi khoá chưa xoá đều có mặt, kể cả khoá chưa học bài nào: phía trước cần
-- biết mẫu số để vẽ tỉ lệ, không chỉ biết tử số.
SELECT
    c.id AS course_id,
    count(l.id)::bigint AS lesson_count,
    count(p.lesson_id)::bigint AS completed_count
FROM courses AS c
LEFT JOIN lessons AS l
    ON l.course_id = c.id AND l.deleted_at IS NULL
LEFT JOIN lesson_progress AS p
    ON p.lesson_id = l.id AND p.user_id = sqlc.arg('user_id')
WHERE c.deleted_at IS NULL
GROUP BY c.id;

-- name: LatestCompletedCourse :one
-- Khoá được đụng tới gần đây nhất. Tách khỏi ListCourseProgress vì max() trên
-- tập rỗng khiến sqlc không suy được kiểu; ở đây "chưa học gì" là không có hàng.
SELECT l.course_id
FROM lesson_progress AS p
JOIN lessons AS l ON l.id = p.lesson_id
JOIN courses AS c ON c.id = l.course_id
WHERE p.user_id = sqlc.arg('user_id')
  AND l.deleted_at IS NULL
  AND c.deleted_at IS NULL
ORDER BY p.completed_at DESC
LIMIT 1;

-- name: ListDailyCompletions :many
-- Gom số bài hoàn thành theo ngày ở múi giờ Việt Nam. XP hôm nay, chuỗi ngày
-- liên tiếp và bảy chấm trong tuần đều suy ra từ đúng bảng này, nên không có
-- con số nào phải giữ đồng bộ bằng tay.
SELECT
    (p.completed_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date AS day,
    count(*)::bigint AS completions
FROM lesson_progress AS p
JOIN lessons AS l ON l.id = p.lesson_id
JOIN courses AS c ON c.id = l.course_id
WHERE p.user_id = sqlc.arg('user_id')
  AND l.deleted_at IS NULL
  AND c.deleted_at IS NULL
GROUP BY 1
ORDER BY 1 DESC
LIMIT 400;
