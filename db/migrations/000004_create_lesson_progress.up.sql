-- Một hàng nghĩa là "người này đã đánh dấu xong bài này". Không có cột trạng
-- thái: bỏ đánh dấu là xoá hàng, nên bảng chỉ có đúng một cách hiểu.
CREATE TABLE lesson_progress (
    user_id      uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    lesson_id    uuid NOT NULL REFERENCES lessons (id) ON DELETE CASCADE,
    completed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, lesson_id)
);

-- Tổng hợp tiến độ luôn quét theo một người và sắp theo thời điểm gần nhất.
CREATE INDEX lesson_progress_user_completed_at_idx
    ON lesson_progress (user_id, completed_at DESC);
