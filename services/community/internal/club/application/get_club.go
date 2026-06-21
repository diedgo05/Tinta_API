package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/community/internal/club/domain"
	"github.com/tinta/community/internal/club/ports"
)

// GetClubUseCase returns a club by ID.
type GetClubUseCase struct {
	repo ports.ClubRepository
}

// NewGetClubUseCase wires the dependency.
func NewGetClubUseCase(repo ports.ClubRepository) *GetClubUseCase {
	return &GetClubUseCase{repo: repo}
}

// Execute returns the club or domain.ErrClubNotFound.
func (uc *GetClubUseCase) Execute(ctx context.Context, id uuid.UUID) (*domain.Club, error) {
	return uc.repo.GetByID(ctx, id)
}
