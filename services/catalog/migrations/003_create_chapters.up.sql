CREATE TABLE IF NOT EXISTS chapters (
    id          UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
    book_id     UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    number      INT  NOT NULL,
    title       VARCHAR(255) NOT NULL,
    start_page  INT,
    end_page    INT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (book_id, number)
);
CREATE INDEX IF NOT EXISTS idx_chapters_book ON chapters(book_id);
