-- Nội dung của bài học. Trước migration này, "bài học" chỉ có tiêu đề: không có
-- chỗ nào chứa thứ người học thật sự phải học.

ALTER TABLE lessons ADD COLUMN summary text NOT NULL DEFAULT '';

-- Ba dạng khối đủ cho cả bài ngữ pháp lẫn bài tình huống:
--   note     — đoạn giải thích bằng tiếng Việt
--   example  — một câu mẫu, tiếng Anh kèm bản dịch
--   dialogue — một lượt thoại, giống example nhưng có tên người nói
CREATE TYPE lesson_block_kind AS ENUM ('note', 'example', 'dialogue');

CREATE TABLE lesson_blocks (
    id         uuid PRIMARY KEY DEFAULT uuidv7(),
    lesson_id  uuid NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    kind       lesson_block_kind NOT NULL,

    -- Mỗi dạng chỉ dùng một phần các cột dưới đây; ràng buộc shape bên dưới
    -- bắt buộc đúng phần đó phải có và phần còn lại phải rỗng.
    body       text NOT NULL DEFAULT '',
    text_en    text NOT NULL DEFAULT '',
    text_vi    text NOT NULL DEFAULT '',
    speaker    text NOT NULL DEFAULT '',

    position   integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT lesson_blocks_position_non_negative CHECK (position >= 0),

    -- Không có ràng buộc này thì một khối "example" vẫn lưu được với text_en
    -- rỗng, và người học nhận một dòng trắng giữa bài.
    CONSTRAINT lesson_blocks_shape CHECK (
        CASE kind
            WHEN 'note' THEN
                length(btrim(body)) > 0 AND text_en = '' AND text_vi = '' AND speaker = ''
            WHEN 'example' THEN
                body = '' AND length(btrim(text_en)) > 0 AND speaker = ''
            WHEN 'dialogue' THEN
                body = '' AND length(btrim(text_en)) > 0 AND length(btrim(speaker)) > 0
        END
    )
);

-- Khối không xoá mềm như course/lesson/vocabulary: nó không có slug để dùng
-- lại, không ai tham chiếu tới, và không gắn tiến độ nào. Xoá là xoá hẳn.
CREATE INDEX lesson_blocks_lesson_position_idx ON lesson_blocks (lesson_id, position, id);
