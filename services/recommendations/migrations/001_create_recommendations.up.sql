-- Recommendations produced by ML models. Each row associates a user with a
-- book and includes a score, source cluster, and the user's feedback.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS recommendations (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID         NOT NULL,
    book_id         UUID         NOT NULL,
    score           DOUBLE PRECISION NOT NULL,
    cluster_id      INTEGER,
    source          VARCHAR(50)  NOT NULL DEFAULT 'collaborative'
                    CHECK (source IN ('collaborative', 'content', 'hybrid', 'trending')),
    feedback        VARCHAR(20)
                    CHECK (feedback IS NULL OR feedback IN ('like', 'dislike')),
    feedback_at     TIMESTAMPTZ,
    dismissed_at    TIMESTAMPTZ,
    generated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, book_id)
);

CREATE INDEX IF NOT EXISTS idx_reco_user_id_active
    ON recommendations(user_id)
    WHERE dismissed_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_reco_user_score
    ON recommendations(user_id, score DESC)
    WHERE dismissed_at IS NULL;
