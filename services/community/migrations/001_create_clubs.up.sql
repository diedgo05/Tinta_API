CREATE TABLE IF NOT EXISTS clubs (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    creator_id   UUID NOT NULL,           -- references identity.users (cross-schema, no FK)
    book_id      UUID,                    -- optional, will reference catalog.books in the future
    name         VARCHAR(120) NOT NULL,
    description  TEXT,
    is_private   BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_clubs_creator_id ON clubs(creator_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_clubs_book_id    ON clubs(book_id)    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_clubs_created_at ON clubs(created_at DESC) WHERE deleted_at IS NULL;

-- Note: there are NO foreign keys to other services' schemas.
-- Cross-service references are validated at the application layer
-- (or via events), per the SOA principle of no shared database.
