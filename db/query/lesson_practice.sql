-- name: RecordLessonPracticeScore :one
-- Ghi một lượt luyện: giữ điểm cao hơn nếu cùng số câu, thay hẳn nếu số câu
-- đã khác — điểm cũ tính trên một bộ từ khác, so với nó là vô nghĩa.
INSERT INTO lesson_practice_scores (user_id, lesson_id, best_correct, total)
VALUES (sqlc.arg('user_id'), sqlc.arg('lesson_id'), sqlc.arg('correct'), sqlc.arg('total'))
ON CONFLICT (user_id, lesson_id) DO UPDATE
SET best_correct = CASE
        WHEN lesson_practice_scores.total = excluded.total
            THEN GREATEST(lesson_practice_scores.best_correct, excluded.best_correct)
        ELSE excluded.best_correct
    END,
    total      = excluded.total,
    attempts   = lesson_practice_scores.attempts + 1,
    updated_at = now()
RETURNING lesson_id, best_correct, total, attempts;

-- name: ListLessonPracticeScores :many
-- Điểm của một người ở mọi bài còn sống. Bài đã xoá mềm thì điểm của nó ẩn đi.
SELECT s.lesson_id, s.best_correct, s.total, s.attempts
FROM lesson_practice_scores AS s
JOIN lessons AS l ON l.id = s.lesson_id
WHERE s.user_id = sqlc.arg('user_id') AND l.deleted_at IS NULL
ORDER BY s.updated_at DESC, s.lesson_id;
