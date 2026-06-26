-- ============================================================================
-- NOTIFICATIONS · sqlc queries
-- ============================================================================

-- name: CreateNotification :one
INSERT INTO notifications (user_id, type, title, body, data)
VALUES ($1,$2,$3,$4,$5) RETURNING *;

-- name: GetNotificationByID :one
SELECT * FROM notifications WHERE id=$1;

-- name: ListNotificationsByUser :many
SELECT * FROM notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: MarkNotificationAsRead :one
UPDATE notifications SET read_at = COALESCE(read_at, NOW()) WHERE id=$1 RETURNING *;

-- name: MarkAllAsRead :execrows
UPDATE notifications SET read_at=NOW() WHERE user_id=$1 AND read_at IS NULL;

-- name: DeleteNotification :exec
DELETE FROM notifications WHERE id=$1;
