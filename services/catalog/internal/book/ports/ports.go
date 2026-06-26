// Package ports declares the interfaces consumed by the application layer.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/catalog/internal/book/domain"
)

// ListFilter for paginated queries.
type ListFilter struct {
	Page     int
	PageSize int
	GenreID  *uuid.UUID
	Search   string // matches title or author (ILIKE %...%)
}

// ListResult bundles a page of books and the total.
type ListResult struct {
	Items    []*domain.Book
	Total    int64
	Page     int
	PageSize int
}

// BookUpdates carries optional fields. nil = "do not change".
type BookUpdates struct {
	GenreID       *uuid.UUID
	Title         *string
	Author        *string
	ISBN          *string
	Synopsis      *string
	CoverURL      *string
	TotalPages    *int
	License       *domain.License
	Language      *string
	PublishedYear *int
}

// BookRepository is the persistence port for the Book aggregate.
type BookRepository interface {
	Create(ctx context.Context, b *domain.Book) (*domain.Book, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Book, error)
	List(ctx context.Context, f ListFilter) (*ListResult, error)
	Update(ctx context.Context, id uuid.UUID, u BookUpdates) (*domain.Book, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
