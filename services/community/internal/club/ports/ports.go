// Package ports declares the interfaces consumed by the application layer
// of the Community service.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/community/internal/club/domain"
)

// ListFilter carries pagination + optional filters for ListClubs.
type ListFilter struct {
	Page         int
	PageSize     int
	IncludeAll   bool       // if false (default), only public clubs are returned
	CreatorID    *uuid.UUID // if set, only clubs from this creator
	BookID       *uuid.UUID // if set, only clubs about this book
}

// ListResult bundles a page of clubs and the total count.
type ListResult struct {
	Items      []*domain.Club
	Total      int64
	Page       int
	PageSize   int
}

// ClubUpdates holds optional fields for a partial update.
// nil = "do not change".
type ClubUpdates struct {
	Name        *string
	Description *string
	BookID      *uuid.UUID
	IsPrivate   *bool
}

// ClubRepository is the persistence port for the Club aggregate.
type ClubRepository interface {
	Create(ctx context.Context, c *domain.Club) (*domain.Club, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Club, error)
	List(ctx context.Context, filter ListFilter) (*ListResult, error)
	Update(ctx context.Context, id uuid.UUID, updates ClubUpdates) (*domain.Club, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
