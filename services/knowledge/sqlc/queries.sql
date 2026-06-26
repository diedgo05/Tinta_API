-- ============================================================================
-- KNOWLEDGE · sqlc queries
-- ============================================================================

-- name: ListTopics :many
SELECT * FROM topics ORDER BY name ASC;

-- name: GetTopicByID :one
SELECT * FROM topics WHERE id=$1;

-- name: ListUserTopics :many
SELECT * FROM user_topics WHERE user_id=$1 ORDER BY selected_at ASC;

-- name: CreateKnowledgeDocument :one
INSERT INTO knowledge_documents (topic_id, title, source, license, url_original, version)
VALUES ($1,$2,$3,$4,$5,$6) RETURNING *;

-- name: ListDocumentsByTopic :many
SELECT * FROM knowledge_documents WHERE topic_id=$1 ORDER BY created_at DESC;

-- name: CreateRAGFragment :one
INSERT INTO rag_fragments (document_id, topic_id, text_chunk, position, tokens, embedding, hash_chunk)
VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING *;

-- name: ListFragmentsByTopic :many
SELECT * FROM rag_fragments WHERE topic_id=$1 ORDER BY position ASC LIMIT $2 OFFSET $3;
