-- name: CreateVocabularyEntry :one
-- Từ mới luôn đứng cuối bài, giống cách bài mới đứng cuối khoá.
INSERT INTO vocabulary_entries (lesson_id, word, ipa, meaning, example, example_vi, position)
VALUES (
    sqlc.arg('lesson_id'),
    sqlc.arg('word'),
    sqlc.arg('ipa'),
    sqlc.arg('meaning'),
    sqlc.arg('example'),
    sqlc.arg('example_vi'),
    (
        SELECT coalesce(max(position) + 1, 0)
        FROM vocabulary_entries
        WHERE lesson_id = sqlc.arg('lesson_id') AND deleted_at IS NULL
    )
)
RETURNING *;

-- name: GetVocabularyEntry :one
SELECT * FROM vocabulary_entries
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: ListVocabularyByLesson :many
SELECT * FROM vocabulary_entries
WHERE lesson_id = sqlc.arg('lesson_id') AND deleted_at IS NULL
ORDER BY position, id;

-- name: UpdateVocabularyEntry :one
UPDATE vocabulary_entries
SET word       = coalesce(sqlc.narg('word'), word),
    ipa        = coalesce(sqlc.narg('ipa'), ipa),
    meaning    = coalesce(sqlc.narg('meaning'), meaning),
    example    = coalesce(sqlc.narg('example'), example),
    example_vi = coalesce(sqlc.narg('example_vi'), example_vi),
    updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteVocabularyEntry :execrows
UPDATE vocabulary_entries
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL;

-- name: ListVocabularyForUser :many
-- Toàn bộ từ thuộc những bài người này đã đánh dấu hoàn thành, kèm lịch ôn nếu
-- đã có. Chưa học bài thì chưa thấy từ của bài — hàng đợi bám theo tiến độ mà
-- không cần một lần ghi nào lúc học xong.
SELECT
    e.id, e.lesson_id, e.word, e.ipa, e.meaning, e.example, e.example_vi,
    e.position, e.created_at, e.updated_at,
    c.level AS level,
    r.ease_factor, r.interval_days, r.repetitions, r.due_on, r.last_reviewed_at
FROM vocabulary_entries AS e
JOIN lessons AS l ON l.id = e.lesson_id
JOIN courses AS c ON c.id = l.course_id
JOIN lesson_progress AS p ON p.lesson_id = l.id AND p.user_id = sqlc.arg('user_id')
LEFT JOIN vocabulary_reviews AS r
    ON r.entry_id = e.id AND r.user_id = sqlc.arg('user_id')
WHERE e.deleted_at IS NULL
  AND l.deleted_at IS NULL
  AND c.deleted_at IS NULL
ORDER BY
    -- Chưa ôn lần nào đứng trước, rồi tới hạn cũ nhất.
    r.due_on NULLS FIRST,
    e.created_at,
    e.id;

-- name: GetVocabularyReview :one
SELECT * FROM vocabulary_reviews
WHERE user_id = sqlc.arg('user_id') AND entry_id = sqlc.arg('entry_id');

-- name: UpsertVocabularyReview :one
INSERT INTO vocabulary_reviews (
    user_id, entry_id, ease_factor, interval_days, repetitions, due_on, last_reviewed_at
)
VALUES (
    sqlc.arg('user_id'),
    sqlc.arg('entry_id'),
    sqlc.arg('ease_factor'),
    sqlc.arg('interval_days'),
    sqlc.arg('repetitions'),
    sqlc.arg('due_on'),
    now()
)
ON CONFLICT (user_id, entry_id) DO UPDATE
SET ease_factor      = excluded.ease_factor,
    interval_days    = excluded.interval_days,
    repetitions      = excluded.repetitions,
    due_on           = excluded.due_on,
    last_reviewed_at = now(),
    updated_at       = now()
RETURNING *;

-- name: CountVocabularyLearnedThisWeek :one
-- "Từ mới tuần này" đếm theo lần ôn đầu tiên, tức là lúc hàng được tạo.
SELECT count(*)::bigint FROM vocabulary_reviews
WHERE user_id = sqlc.arg('user_id')
  AND created_at >= now() - interval '7 days';
