CREATE TABLE IF NOT EXISTS annotations (
    id                 UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
    user_id            UUID NOT NULL,
    book_id            UUID,
    personal_doc_id    UUID,
    page               INT  NOT NULL DEFAULT 0,
    highlighted_text   TEXT NOT NULL,
    personal_note      TEXT,
    color              VARCHAR(20) NOT NULL DEFAULT 'yellow',
    deleted_at         TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (book_id IS NOT NULL OR personal_doc_id IS NOT NULL)
);
CREATE INDEX IF NOT EXISTS idx_annot_user_book ON annotations(user_id, book_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_annot_user_doc  ON annotations(user_id, personal_doc_id) WHERE deleted_at IS NULL;
