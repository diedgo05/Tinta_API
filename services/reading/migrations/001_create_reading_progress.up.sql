CREATE TABLE IF NOT EXISTS reading_progress (
    id              UUID PRIMARY KEY DEFAULT public.uuid_generate_v4(),
    user_id         UUID NOT NULL,
    book_id         UUID NOT NULL,
    current_page    INT  NOT NULL DEFAULT 0,
    total_time_min  INT  NOT NULL DEFAULT 0,
    status          VARCHAR(20) NOT NULL DEFAULT 'reading'
                    CHECK (status IN ('reading','paused','finished','abandoned')),
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_read_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, book_id)
);
CREATE INDEX IF NOT EXISTS idx_reading_user ON reading_progress(user_id);
CREATE INDEX IF NOT EXISTS idx_reading_status ON reading_progress(user_id, status);
