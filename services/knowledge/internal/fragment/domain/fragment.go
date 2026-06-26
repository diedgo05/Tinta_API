// Package domain contains the Fragment and KnowledgeDocument entities.
package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type KnowledgeDocument struct {
	ID          uuid.UUID
	TopicID     uuid.UUID
	Title       string
	Source      string
	License     string
	URLOriginal string
	Version     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Fragment is a chunked piece of a knowledge document. The mobile downloads
// the fragments to feed the on-device RAG (sqlite-vec).
type Fragment struct {
	ID         uuid.UUID
	DocumentID uuid.UUID
	TopicID    uuid.UUID
	TextChunk  string
	Position   int
	Tokens     int
	Embedding  json.RawMessage // [float, float, ...]
	HashChunk  string
	CreatedAt  time.Time
}

var (
	ErrFragmentNotFound = errors.New("fragment not found")
	ErrDocumentNotFound = errors.New("document not found")
	ErrInvalidPageSize  = errors.New("invalid page size")
)
