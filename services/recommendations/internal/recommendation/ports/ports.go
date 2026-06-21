// Package ports declares the interfaces consumed by the application layer.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/recommendations/internal/recommendation/domain"
)

// ListFilter carries pagination for the user's recommendations.
type ListFilter struct {
	UserID   uuid.UUID
	Page     int
	PageSize int
}

// ListResult bundles a page of recommendations and the total.
type ListResult struct {
	Items    []*domain.Recommendation
	Total    int64
	Page     int
	PageSize int
}

// RecommendationRepository persists recommendations.
type RecommendationRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Recommendation, error)
	List(ctx context.Context, filter ListFilter) (*ListResult, error)
	SetFeedback(ctx context.Context, id uuid.UUID, feedback domain.Feedback) (*domain.Recommendation, error)
	Dismiss(ctx context.Context, id uuid.UUID) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

// MLPipeline is the abstraction over whatever produces recommendations.
// In V1 we expose a stub that publishes a job; the actual ML pipeline runs
// asynchronously in Python and writes results back to this service's table.
type MLPipeline interface {
	RegenerateForUser(ctx context.Context, userID uuid.UUID) error
}
