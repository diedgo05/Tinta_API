// Package application contains the use cases of the Club module.
package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/tinta/community/internal/club/domain"
	"github.com/tinta/community/internal/club/ports"
)

// CreateClubInput is the input DTO for the CreateClub use case.
type CreateClubInput struct {
	CreatorID   uuid.UUID
	Name        string
	Description string
	BookID      *uuid.UUID
	IsPrivate   bool
}

// CreateClubUseCase creates a new reading club.
type CreateClubUseCase struct {
	repo ports.ClubRepository
}

// NewCreateClubUseCase wires the dependency.
func NewCreateClubUseCase(repo ports.ClubRepository) *CreateClubUseCase {
	return &CreateClubUseCase{repo: repo}
}

// Execute persists a new club owned by CreatorID.
func (uc *CreateClubUseCase) Execute(ctx context.Context, in CreateClubInput) (*domain.Club, error) {
	if err := domain.ValidateName(in.Name); err != nil {
		return nil, err
	}

	club := &domain.Club{
		CreatorID:   in.CreatorID,
		Name:        strings.TrimSpace(in.Name),
		Description: strings.TrimSpace(in.Description),
		BookID:      in.BookID,
		IsPrivate:   in.IsPrivate,
	}

	created, err := uc.repo.Create(ctx, club)
	if err != nil {
		return nil, fmt.Errorf("create club: %w", err)
	}
	return created, nil
}
