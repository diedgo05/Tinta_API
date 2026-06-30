// Package ports declares the RecommendedBookRepository interface.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/recommendations/internal/recommendedbook/domain"
)

// RecommendedBookRepository is the persistence port for RecommendedBook.
type RecommendedBookRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.RecommendedBook, error)
	// GetByIDs returns a map keyed by book_id for efficient batch lookup.
	// Books that are not found are simply omitted from the map (no error).
	GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.RecommendedBook, error)
}
