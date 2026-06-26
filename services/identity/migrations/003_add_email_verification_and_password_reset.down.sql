ALTER TABLE users
    DROP COLUMN IF EXISTS email_verified,
    DROP COLUMN IF EXISTS verification_code,
    DROP COLUMN IF EXISTS verification_expires_at,
    DROP COLUMN IF EXISTS password_reset_code,
    DROP COLUMN IF EXISTS password_reset_expires_at;
