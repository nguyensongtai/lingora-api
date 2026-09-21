CREATE TABLE lessons (
    id         uuid PRIMARY KEY DEFAULT uuidv7(),
    course_id  uuid NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
    slug       text NOT NULL,
    title      text NOT NULL,
    position   integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT lessons_slug_format CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    CONSTRAINT lessons_title_not_blank CHECK (length(btrim(title)) > 0),
    CONSTRAINT lessons_position_non_negative CHECK (position >= 0)
);

CREATE UNIQUE INDEX lessons_course_id_slug_key ON lessons (course_id, slug) WHERE deleted_at IS NULL;

-- Bài học luôn được liệt kê theo thứ tự trong một khoá.
CREATE INDEX lessons_course_id_position_idx ON lessons (course_id, position, id) WHERE deleted_at IS NULL;
