-- Điểm tốt nhất của mỗi người ở bước Luyện tập của mỗi bài. Người dùng chốt:
-- chỉ để người học thấy mình đã nắm bài tới đâu — KHÔNG cộng XP. XP vẫn chỉ
-- đến từ việc hoàn thành bài.
--
-- Chỉ giữ điểm tốt nhất, không giữ lịch sử từng lượt: chưa có màn hình nào
-- cần tới lịch sử đó.
CREATE TABLE lesson_practice_scores (
    user_id      uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    lesson_id    uuid NOT NULL REFERENCES lessons (id) ON DELETE CASCADE,
    best_correct integer NOT NULL,
    -- total là số câu của lượt đạt điểm đó. Bài đổi số từ thì total đổi theo,
    -- và điểm cũ — tính trên một bộ từ khác — bị thay chứ không so tiếp.
    total        integer NOT NULL,
    attempts     integer NOT NULL DEFAULT 1,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, lesson_id),
    CONSTRAINT lesson_practice_scores_total_positive CHECK (total > 0),
    CONSTRAINT lesson_practice_scores_correct_range CHECK (best_correct BETWEEN 0 AND total),
    CONSTRAINT lesson_practice_scores_attempts_positive CHECK (attempts > 0)
);
