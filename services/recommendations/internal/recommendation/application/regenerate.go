package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/recommendations/internal/recommendation/ports"
)

// RegenerateRecommendationsUseCase asks the ML pipeline to produce fresh
// recommendations for the requesting user.
//
// The pipeline itself runs in Python and writes results back to this service's
// table. From the API's perspective, this is fire-and-forget.
type RegenerateRecommendationsUseCase struct {
	pipeline ports.MLPipeline
}

// NewRegenerateRecommendationsUseCase wires the dependency.
func NewRegenerateRecommendationsUseCase(pipeline ports.MLPipeline) *RegenerateRecommendationsUseCase {
	return &RegenerateRecommendationsUseCase{pipeline: pipeline}
}

// Execute publishes a regeneration job for the user.
func (uc *RegenerateRecommendationsUseCase) Execute(ctx context.Context, userID uuid.UUID) error {
	return uc.pipeline.RegenerateForUser(ctx, userID)
}
