-- ============================================================================
-- CATALOG · sqlc queries
-- ============================================================================

-- name: CreateGenre :one
INSERT INTO genres (name, slug, description) VALUES ($1,$2,$3) RETURNING *;

-- name: GetGenreByID :one
SELECT * FROM genres WHERE id=$1;

-- name: ListGenres :many
SELECT * FROM genres ORDER BY name ASC;

-- name: DeleteGenre :exec
DELETE FROM genres WHERE id=$1;

-- name: CreateBook :one
INSERT INTO books (genre_id, title, author, isbn, synopsis, cover_url, total_pages, license, language, published_year)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING *;

-- name: GetBookByID :one
SELECT * FROM books WHERE id=$1 AND deleted_at IS NULL;

-- name: SoftDeleteBook :exec
UPDATE books SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL;

-- name: CreateChapter :one
INSERT INTO chapters (book_id, number, title, start_page, end_page)
VALUES ($1,$2,$3,$4,$5) RETURNING *;

-- name: ListChaptersByBook :many
SELECT * FROM chapters WHERE book_id=$1 ORDER BY number ASC;
