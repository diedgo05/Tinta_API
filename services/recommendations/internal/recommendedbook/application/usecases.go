// Package application contains the RecommendedBook use cases.
package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/recommendations/internal/recommendedbook/domain"
	"github.com/tinta/recommendations/internal/recommendedbook/ports"
)

// GetRecommendedBooksUseCase batch-fetches metadata for a list of book IDs.
// Used by the recommendation handler to enrich the list response.
type GetRecommendedBooksUseCase struct {
	repo ports.RecommendedBookRepository
}

func NewGetRecommendedBooksUseCase(r ports.RecommendedBookRepository) *GetRecommendedBooksUseCase {
	return &GetRecommendedBooksUseCase{repo: r}
}

func (uc *GetRecommendedBooksUseCase) Execute(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.RecommendedBook, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]*domain.RecommendedBook{}, nil
	}
	return uc.repo.GetByIDs(ctx, ids)
}
