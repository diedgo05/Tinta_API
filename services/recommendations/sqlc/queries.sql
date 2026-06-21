-- ============================================================================
-- RECOMMENDATIONS · sqlc queries
-- ============================================================================

-- name: GetRecommendationByID :one
SELECT * FROM recommendations WHERE id = $1;

-- name: ListUserRecommendations :many
SELECT * FROM recommendations
WHERE user_id = $1 AND dismissed_at IS NULL
ORDER BY score DESC, generated_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUserRecommendations :one
SELECT COUNT(*) FROM recommendations
WHERE user_id = $1 AND dismissed_at IS NULL;

-- name: SetRecommendationFeedback :one
UPDATE recommendations
SET feedback = $2, feedback_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DismissRecommendation :exec
UPDATE recommendations
SET dismissed_at = NOW(), updated_at = NOW()
WHERE id = $1 AND dismissed_at IS NULL;

-- name: DeleteUserRecommendations :exec
DELETE FROM recommendations WHERE user_id = $1;

-- Used by the ML pipeline to bulk-insert new recommendations.
-- name: InsertRecommendation :one
INSERT INTO recommendations (user_id, book_id, score, cluster_id, source, generated_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (user_id, book_id) DO UPDATE
SET score = EXCLUDED.score,
    cluster_id = EXCLUDED.cluster_id,
    source = EXCLUDED.source,
    generated_at = NOW(),
    dismissed_at = NULL,
    updated_at = NOW()
RETURNING *;
