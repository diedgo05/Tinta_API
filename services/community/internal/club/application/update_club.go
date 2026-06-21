package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/community/internal/club/domain"
	"github.com/tinta/community/internal/club/ports"
)

// UpdateClubInput holds the optional update fields plus the requesting user.
type UpdateClubInput struct {
	ClubID      uuid.UUID
	RequesterID uuid.UUID
	Name        *string
	Description *string
	BookID      *uuid.UUID
	IsPrivate   *bool
}

// UpdateClubUseCase updates a club. Only the creator can update it.
type UpdateClubUseCase struct {
	repo ports.ClubRepository
}

// NewUpdateClubUseCase wires the dependency.
func NewUpdateClubUseCase(repo ports.ClubRepository) *UpdateClubUseCase {
	return &UpdateClubUseCase{repo: repo}
}

// Execute validates ownership and applies the partial update.
func (uc *UpdateClubUseCase) Execute(ctx context.Context, in UpdateClubInput) (*domain.Club, error) {
	// 1. Fetch the existing club to check ownership.
	existing, err := uc.repo.GetByID(ctx, in.ClubID)
	if err != nil {
		return nil, err
	}
	if !existing.CanBeManagedBy(in.RequesterID) {
		return nil, domain.ErrNotAuthorized
	}

	// 2. Validate optional fields.
	if in.Name != nil {
		if err := domain.ValidateName(*in.Name); err != nil {
			return nil, err
		}
	}

	// 3. Apply the update.
	return uc.repo.Update(ctx, in.ClubID, ports.ClubUpdates{
		Name:        in.Name,
		Description: in.Description,
		BookID:      in.BookID,
		IsPrivate:   in.IsPrivate,
	})
}
