// Package ports declares the Discussion repository interface.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/community/internal/discussion/domain"
)

type ListFilter struct {
	ClubID        uuid.UUID
	ChapterNumber *int
	Page          int
	PageSize      int
}

type ListResult struct {
	Items    []*domain.Discussion
	Total    int64
	Page     int
	PageSize int
}

type DiscussionRepository interface {
	Create(ctx context.Context, d *domain.Discussion) (*domain.Discussion, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Discussion, error)
	List(ctx context.Context, f ListFilter) (*ListResult, error)
	UpdateContent(ctx context.Context, id uuid.UUID, content string) (*domain.Discussion, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
