package application

import (
	"context"

	"github.com/tinta/community/internal/club/domain"
	"github.com/tinta/community/internal/club/ports"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// ListClubsUseCase returns a paginated list of clubs.
type ListClubsUseCase struct {
	repo ports.ClubRepository
}

// NewListClubsUseCase wires the dependency.
func NewListClubsUseCase(repo ports.ClubRepository) *ListClubsUseCase {
	return &ListClubsUseCase{repo: repo}
}

// Execute applies sane defaults and queries the repository.
func (uc *ListClubsUseCase) Execute(ctx context.Context, filter ports.ListFilter) (*ports.ListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = defaultPageSize
	}
	if filter.PageSize > maxPageSize {
		return nil, domain.ErrInvalidPageSize
	}
	return uc.repo.List(ctx, filter)
}
