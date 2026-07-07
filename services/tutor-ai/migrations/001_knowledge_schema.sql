-- ═══════════════════════════════════════════════════════════════════════════
-- Migración 001: esquema `knowledge` para RAG
--
-- Crea las tablas necesarias para:
--   - Registrar documentos PDF (subidos por usuarios o curados)
--   - Almacenar chunks (fragmentos de texto)
--   - Almacenar embeddings vectoriales para búsqueda semántica
--
-- Requiere la extensión pgvector. Railway Postgres la trae disponible pero
-- hay que habilitarla explícitamente.
-- ═══════════════════════════════════════════════════════════════════════════

CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE SCHEMA IF NOT EXISTS knowledge;

-- ───────────────────────────────────────────────────────────────────────────
-- knowledge.documents
--
-- Un documento representa un PDF completo, ya sea:
--   - user_id NOT NULL: PDF subido por un usuario
--   - user_id NULL:     contenido curado por el equipo (futura base académica)
--
-- El campo `status` refleja el estado del pipeline de ingesta.
-- ───────────────────────────────────────────────────────────────────────────
CREATE TABLE knowledge.documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID,
    filename        VARCHAR(255) NOT NULL,
    total_pages     INT,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'processing', 'ready', 'failed')),
    error_message   TEXT,
    uploaded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ
);

CREATE INDEX idx_documents_user_id ON knowledge.documents(user_id)
    WHERE user_id IS NOT NULL;

CREATE INDEX idx_documents_status ON knowledge.documents(status);

-- ───────────────────────────────────────────────────────────────────────────
-- knowledge.chunks
--
-- Cada documento se parte en muchos chunks de ~500 palabras con overlap.
-- Un chunk conoce su documento padre y su posición dentro de él.
-- El page_number permite al LLM decir "según la página 15..."
-- ───────────────────────────────────────────────────────────────────────────
CREATE TABLE knowledge.chunks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id     UUID NOT NULL REFERENCES knowledge.documents(id) ON DELETE CASCADE,
    chunk_text      TEXT NOT NULL,
    chunk_position  INT NOT NULL,
    page_number     INT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chunks_document ON knowledge.chunks(document_id, chunk_position);

-- ───────────────────────────────────────────────────────────────────────────
-- knowledge.chunk_embeddings
--
-- Vectores de 384 dimensiones (tamaño de MiniLM-L12-v2 multilingüe).
-- Índice HNSW para búsqueda por similitud coseno rápida.
-- ───────────────────────────────────────────────────────────────────────────
CREATE TABLE knowledge.chunk_embeddings (
    chunk_id        UUID PRIMARY KEY REFERENCES knowledge.chunks(id) ON DELETE CASCADE,
    embedding       vector(384) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Índice HNSW: mejor precisión que IVFFlat para volúmenes pequeños/medianos.
-- Parámetros por defecto: m=16, ef_construction=64 (buen balance para PoC).
CREATE INDEX idx_chunk_embeddings_hnsw
    ON knowledge.chunk_embeddings
    USING hnsw (embedding vector_cosine_ops);
