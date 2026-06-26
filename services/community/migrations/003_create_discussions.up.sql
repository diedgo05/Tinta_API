-- ============================================================================
-- Community · 003: Discusiones (mensajes dentro de un club)
-- ============================================================================

CREATE TABLE IF NOT EXISTS discussions (
    id              UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
    club_id         UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL,
    chapter_number  INT,
    content         TEXT NOT NULL,
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_discussions_club_chapter
    ON discussions(club_id, chapter_number, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_discussions_user
    ON discussions(user_id) WHERE deleted_at IS NULL;
