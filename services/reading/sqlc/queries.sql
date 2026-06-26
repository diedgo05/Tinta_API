-- ============================================================================
-- READING · sqlc queries
-- ============================================================================

-- name: UpsertProgress :one
INSERT INTO reading_progress (user_id, book_id, current_page, status)
VALUES ($1,$2,$3,$4)
ON CONFLICT (user_id, book_id) DO UPDATE
SET current_page = GREATEST(reading_progress.current_page, EXCLUDED.current_page),
    last_read_at = NOW(), updated_at = NOW()
RETURNING *;

-- name: GetProgress :one
SELECT * FROM reading_progress WHERE user_id=$1 AND book_id=$2;

-- name: ListProgressByUser :many
SELECT * FROM reading_progress WHERE user_id=$1 ORDER BY last_read_at DESC;

-- name: CreateAnnotation :one
INSERT INTO annotations (user_id, book_id, personal_doc_id, page, highlighted_text, personal_note, color)
VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING *;

-- name: ListAnnotationsByBook :many
SELECT * FROM annotations
WHERE user_id=$1 AND book_id=$2 AND deleted_at IS NULL
ORDER BY page ASC, created_at ASC;

-- name: SoftDeleteAnnotation :exec
UPDATE annotations SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL;
