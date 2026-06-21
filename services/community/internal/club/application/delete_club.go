package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/community/internal/club/domain"
	"github.com/tinta/community/internal/club/ports"
)

// DeleteClubUseCase soft-deletes a club. Only the creator can do this.
type DeleteClubUseCase struct {
	repo ports.ClubRepository
}

// NewDeleteClubUseCase wires the dependency.
func NewDeleteClubUseCase(repo ports.ClubRepository) *DeleteClubUseCase {
	return &DeleteClubUseCase{repo: repo}
}

// Execute checks ownership and soft-deletes the club.
func (uc *DeleteClubUseCase) Execute(ctx context.Context, clubID, requesterID uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, clubID)
	if err != nil {
		return err
	}
	if !existing.CanBeManagedBy(requesterID) {
		return domain.ErrNotAuthorized
	}
	return uc.repo.SoftDelete(ctx, clubID)
}
