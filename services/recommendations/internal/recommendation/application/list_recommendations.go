// Package application contains the use cases of the Recommendation module.
package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/recommendations/internal/recommendation/ports"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// ListRecommendationsUseCase returns the user's active recommendations.
type ListRecommendationsUseCase struct {
	repo ports.RecommendationRepository
}

// NewListRecommendationsUseCase wires the dependency.
func NewListRecommendationsUseCase(repo ports.RecommendationRepository) *ListRecommendationsUseCase {
	return &ListRecommendationsUseCase{repo: repo}
}

// Execute returns a paginated list, sorted by score desc.
func (uc *ListRecommendationsUseCase) Execute(ctx context.Context, userID uuid.UUID, page, pageSize int) (*ports.ListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return uc.repo.List(ctx, ports.ListFilter{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	})
}
