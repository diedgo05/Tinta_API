-- Tabla principal de usuarios del sistema Tinta.
-- Vive en el esquema 'identity' (configurado en search_path por el servicio).

CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    name            VARCHAR(120) NOT NULL,
    role            VARCHAR(20)  NOT NULL DEFAULT 'user'
                    CHECK (role IN ('user', 'moderator', 'admin', 'system')),
    email_verified  BOOLEAN      NOT NULL DEFAULT FALSE,
    avatar_url      VARCHAR(500),
    language        VARCHAR(10)  NOT NULL DEFAULT 'es',
    deleted_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email      ON users(email)      WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_role       ON users(role)       WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at) WHERE deleted_at IS NULL;
