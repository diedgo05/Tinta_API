package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/recommendations/internal/recommendation/domain"
	"github.com/tinta/recommendations/internal/recommendation/ports"
)

// SubmitFeedbackInput carries the feedback request.
type SubmitFeedbackInput struct {
	RecommendationID uuid.UUID
	RequesterID      uuid.UUID
	Feedback         domain.Feedback
}

// SubmitFeedbackUseCase records the user's reaction to a recommendation.
// This feedback feeds future ML retraining jobs.
type SubmitFeedbackUseCase struct {
	repo ports.RecommendationRepository
}

// NewSubmitFeedbackUseCase wires the dependency.
func NewSubmitFeedbackUseCase(repo ports.RecommendationRepository) *SubmitFeedbackUseCase {
	return &SubmitFeedbackUseCase{repo: repo}
}

// Execute validates the feedback and persists it.
func (uc *SubmitFeedbackUseCase) Execute(ctx context.Context, in SubmitFeedbackInput) (*domain.Recommendation, error) {
	if !in.Feedback.IsValid() {
		return nil, domain.ErrInvalidFeedback
	}

	// Verify the recommendation belongs to the requester.
	existing, err := uc.repo.GetByID(ctx, in.RecommendationID)
	if err != nil {
		return nil, err
	}
	if existing.UserID != in.RequesterID {
		return nil, domain.ErrNotAuthorized
	}

	return uc.repo.SetFeedback(ctx, in.RecommendationID, in.Feedback)
}
