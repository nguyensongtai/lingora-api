CREATE TYPE course_level AS ENUM ('A1', 'A2', 'B1', 'B2', 'C1', 'C2');
CREATE TYPE course_status AS ENUM ('draft', 'published');

CREATE TABLE courses (
    id              uuid PRIMARY KEY DEFAULT uuidv7(),
    slug            text NOT NULL,
    title           text NOT NULL,
    description     text NOT NULL DEFAULT '',
    level           course_level NOT NULL,
    status          course_status NOT NULL DEFAULT 'draft',
    cover_image_url text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz,
    CONSTRAINT courses_slug_format CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    CONSTRAINT courses_title_not_blank CHECK (length(btrim(title)) > 0)
);

-- Slug chỉ unique trong phạm vi hàng chưa xoá mềm, nên slug được tái sử dụng sau khi xoá.
CREATE UNIQUE INDEX courses_slug_key ON courses (slug) WHERE deleted_at IS NULL;

-- Index phục vụ keyset pagination ORDER BY created_at DESC, id DESC.
CREATE INDEX courses_created_at_id_idx ON courses (created_at DESC, id DESC) WHERE deleted_at IS NULL;
CREATE INDEX courses_status_created_at_id_idx ON courses (status, created_at DESC, id DESC) WHERE deleted_at IS NULL;
