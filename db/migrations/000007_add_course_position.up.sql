-- Lộ trình cần thứ tự do người soạn đặt, không phải thứ tự ngày tạo.
ALTER TABLE courses ADD COLUMN position integer NOT NULL DEFAULT 0;

-- Điền cho dữ liệu cũ: giữ nguyên thứ tự đang thấy — trong mỗi bậc, khoá tạo
-- trước đứng trước.
WITH ordered AS (
    SELECT id,
           (row_number() OVER (PARTITION BY level ORDER BY created_at, id) - 1)::integer AS position
    FROM courses
    WHERE deleted_at IS NULL
)
UPDATE courses
SET position = ordered.position
FROM ordered
WHERE courses.id = ordered.id;

CREATE INDEX courses_level_position_idx ON courses (level, position) WHERE deleted_at IS NULL;
