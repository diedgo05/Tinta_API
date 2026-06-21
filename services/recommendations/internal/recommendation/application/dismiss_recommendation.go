package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/recommendations/internal/recommendation/domain"
	"github.com/tinta/recommendations/internal/recommendation/ports"
)

// DismissRecommendationUseCase hides a recommendation from the user's list.
type DismissRecommendationUseCase struct {
	repo ports.RecommendationRepository
}

// NewDismissRecommendationUseCase wires the dependency.
func NewDismissRecommendationUseCase(repo ports.RecommendationRepository) *DismissRecommendationUseCase {
	return &DismissRecommendationUseCase{repo: repo}
}

// Execute marks the recommendation as dismissed after verifying ownership.
func (uc *DismissRecommendationUseCase) Execute(ctx context.Context, id, requesterID uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing.UserID != requesterID {
		return domain.ErrNotAuthorized
	}
	return uc.repo.Dismiss(ctx, id)
}
