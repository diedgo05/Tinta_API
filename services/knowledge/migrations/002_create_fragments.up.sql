-- Documentos curados de la base de conocimientos
CREATE TABLE IF NOT EXISTS knowledge_documents (
    id            UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
    topic_id      UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    title         VARCHAR(500) NOT NULL,
    source        VARCHAR(50)  NOT NULL,
    license       VARCHAR(50)  NOT NULL DEFAULT 'cc-by',
    url_original  VARCHAR(1000),
    version       VARCHAR(20)  NOT NULL DEFAULT 'v1',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_knowdocs_topic ON knowledge_documents(topic_id);

-- Fragmentos descargables al celular (texto + embedding vector como JSON)
CREATE TABLE IF NOT EXISTS rag_fragments (
    id           UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
    document_id  UUID NOT NULL REFERENCES knowledge_documents(id) ON DELETE CASCADE,
    topic_id     UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    text_chunk   TEXT NOT NULL,
    position     INT  NOT NULL DEFAULT 0,
    tokens       INT  NOT NULL DEFAULT 0,
    -- embedding vector stored as JSON array of floats; ChromaDB holds the real index in V2
    embedding    JSONB,
    hash_chunk   VARCHAR(64) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_fragments_topic    ON rag_fragments(topic_id);
CREATE INDEX IF NOT EXISTS idx_fragments_document ON rag_fragments(document_id);
CREATE INDEX IF NOT EXISTS idx_fragments_hash     ON rag_fragments(hash_chunk);

-- Selección de temas por usuario (2-5 al onboarding)
CREATE TABLE IF NOT EXISTS user_topics (
    id              UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
    user_id         UUID NOT NULL,
    topic_id        UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    downloaded      BOOLEAN NOT NULL DEFAULT FALSE,
    selected_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    downloaded_at   TIMESTAMPTZ,
    version         VARCHAR(20) NOT NULL DEFAULT 'v1',
    UNIQUE (user_id, topic_id)
);
CREATE INDEX IF NOT EXISTS idx_user_topics_user ON user_topics(user_id);
