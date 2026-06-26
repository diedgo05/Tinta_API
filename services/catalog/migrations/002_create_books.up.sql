CREATE TABLE IF NOT EXISTS books (
    id              UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
    genre_id        UUID REFERENCES genres(id) ON DELETE SET NULL,
    title           VARCHAR(255) NOT NULL,
    author          VARCHAR(255) NOT NULL,
    isbn            VARCHAR(20),
    synopsis        TEXT,
    cover_url       VARCHAR(500),
    total_pages     INT NOT NULL DEFAULT 0,
    license         VARCHAR(50) NOT NULL DEFAULT 'unknown'
                    CHECK (license IN ('public_domain','creative_commons','copyrighted','user_owned','unknown')),
    language        VARCHAR(10) NOT NULL DEFAULT 'es',
    published_year  INT,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_books_genre      ON books(genre_id)   WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_books_title      ON books(title)      WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_books_author     ON books(author)     WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_books_isbn       ON books(isbn)       WHERE deleted_at IS NULL;
