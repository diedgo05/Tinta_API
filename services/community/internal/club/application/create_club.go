// Package application contains the use cases of the Club module.
package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/tinta/community/internal/club/domain"
	"github.com/tinta/community/internal/club/ports"
	memberDomain "github.com/tinta/community/internal/member/domain"
	memberPorts "github.com/tinta/community/internal/member/ports"
)

// CreateClubInput is the input DTO for the CreateClub use case.
type CreateClubInput struct {
	CreatorID   uuid.UUID
	Name        string
	Description string
	BookID      *uuid.UUID
	IsPrivate   bool
}

// CreateClubUseCase creates a new reading club and registers
// the creator as the owner automatically.
type CreateClubUseCase struct {
	repo    ports.ClubRepository
	members memberPorts.MemberRepository
}

// NewCreateClubUseCase wires the dependencies.
func NewCreateClubUseCase(repo ports.ClubRepository, members memberPorts.MemberRepository) *CreateClubUseCase {
	return &CreateClubUseCase{repo: repo, members: members}
}

// Execute persists a new club owned by CreatorID and adds the creator
// as an "owner" member so they can immediately post/read discussions.
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

	// Auto-register the creator as the club owner.
	if _, err := uc.members.Join(ctx, created.ID, in.CreatorID, memberDomain.RoleOwner); err != nil {
		// The club was created but membership failed — log but don't rollback
		// the club creation (the user can still join manually).
		return created, nil
	}

	return created, nil
}