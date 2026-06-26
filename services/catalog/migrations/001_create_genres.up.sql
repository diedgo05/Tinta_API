CREATE TABLE IF NOT EXISTS genres (
    id          UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
    name        VARCHAR(120) NOT NULL UNIQUE,
    slug        VARCHAR(120) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_genres_slug ON genres(slug);
