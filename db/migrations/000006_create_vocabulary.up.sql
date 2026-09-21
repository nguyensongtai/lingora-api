-- Từ vựng thuộc về một bài học. Bậc CEFR suy ra từ khoá chứa bài nên không cần
-- cột riêng, và xoá bài thì từ của bài đi theo.
CREATE TABLE vocabulary_entries (
    id         uuid PRIMARY KEY DEFAULT uuidv7(),
    lesson_id  uuid NOT NULL REFERENCES lessons (id) ON DELETE CASCADE,
    word       text NOT NULL,
    ipa        text NOT NULL DEFAULT '',
    meaning    text NOT NULL,
    example    text NOT NULL DEFAULT '',
    example_vi text NOT NULL DEFAULT '',
    position   integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT vocabulary_entries_word_not_blank CHECK (length(btrim(word)) > 0),
    CONSTRAINT vocabulary_entries_meaning_not_blank CHECK (length(btrim(meaning)) > 0)
);

-- Một từ chỉ xuất hiện một lần trong cùng bài; so không phân biệt hoa thường.
CREATE UNIQUE INDEX vocabulary_entries_lesson_word_key
    ON vocabulary_entries (lesson_id, lower(word)) WHERE deleted_at IS NULL;

CREATE INDEX vocabulary_entries_lesson_position_idx
    ON vocabulary_entries (lesson_id, position) WHERE deleted_at IS NULL;

-- Lịch ôn theo SM-2 của từng người với từng từ. Không có hàng nghĩa là người đó
-- chưa ôn từ này lần nào — hàng đợi "đến hạn" tự suy ra, không phải ghi trước.
CREATE TABLE vocabulary_reviews (
    user_id          uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    entry_id         uuid NOT NULL REFERENCES vocabulary_entries (id) ON DELETE CASCADE,
    -- Ba cột trạng thái của SM-2. ease_factor là double precision cho gọn phía
    -- Go; sai số dấu phẩy động không ảnh hưởng tới lịch tính bằng ngày.
    ease_factor      double precision NOT NULL DEFAULT 2.5,
    interval_days    integer NOT NULL DEFAULT 0,
    repetitions      integer NOT NULL DEFAULT 0,
    due_on           date NOT NULL,
    last_reviewed_at timestamptz NOT NULL DEFAULT now(),
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, entry_id),
    CONSTRAINT vocabulary_reviews_ease_floor CHECK (ease_factor >= 1.3),
    CONSTRAINT vocabulary_reviews_interval_not_negative CHECK (interval_days >= 0),
    CONSTRAINT vocabulary_reviews_repetitions_not_negative CHECK (repetitions >= 0)
);

CREATE INDEX vocabulary_reviews_due_idx ON vocabulary_reviews (user_id, due_on);
