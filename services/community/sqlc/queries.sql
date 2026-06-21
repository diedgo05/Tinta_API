-- ============================================================================
-- COMMUNITY · sqlc queries
-- ============================================================================

-- name: CreateClub :one
INSERT INTO clubs (creator_id, book_id, name, description, is_private)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetClubByID :one
SELECT * FROM clubs
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateClub :one
UPDATE clubs
SET
    name        = COALESCE(sqlc.narg(name)::varchar,        name),
    description = COALESCE(sqlc.narg(description)::text,    description),
    book_id     = COALESCE(sqlc.narg(book_id)::uuid,        book_id),
    is_private  = COALESCE(sqlc.narg(is_private)::boolean,  is_private),
    updated_at  = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteClub :exec
UPDATE clubs
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListPublicClubs :many
SELECT * FROM clubs
WHERE deleted_at IS NULL AND is_private = FALSE
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountPublicClubs :one
SELECT COUNT(*) FROM clubs
WHERE deleted_at IS NULL AND is_private = FALSE;
