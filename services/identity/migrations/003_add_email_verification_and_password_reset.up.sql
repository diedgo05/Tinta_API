-- ============================================================================
-- Identity · 003: Email verification + password recovery
-- ADITIVA: solo añade columnas; no toca datos existentes.
-- Los admins ya creados quedan automáticamente con email_verified=TRUE.
-- ============================================================================

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS email_verified              BOOLEAN     NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS verification_code           VARCHAR(6),
    ADD COLUMN IF NOT EXISTS verification_expires_at     TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS password_reset_code         VARCHAR(6),
    ADD COLUMN IF NOT EXISTS password_reset_expires_at   TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_users_verification_code
    ON users(verification_code) WHERE verification_code IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_password_reset_code
    ON users(password_reset_code) WHERE password_reset_code IS NOT NULL;
