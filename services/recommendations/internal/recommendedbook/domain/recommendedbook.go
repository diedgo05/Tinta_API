// Package domain contains the RecommendedBook aggregate.
//
// A RecommendedBook is the metadata of a book surfaced by the ML pipeline
// (Google Books API). It lives in the recommendations service alongside the
// `recommendations` table so the HTTP handler can return a single response
// with title/author/thumbnail without crossing service boundaries.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// RecommendedBook is the metadata of a book recommended by the ML pipeline.
type RecommendedBook struct {
	BookID         uuid.UUID
	GoogleVolumeID string
	Title          string
	Authors        []string
	Thumbnail      string
	InfoLink       string
	Description    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Errors.
var ErrRecommendedBookNotFound = errors.New("recommended book not found")
