// Package ports declares the Genre repository interface.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/catalog/internal/genre/domain"
)

type GenreUpdates struct {
	Name        *string
	Description *string
}

type GenreRepository interface {
	Create(ctx context.Context, g *domain.Genre) (*domain.Genre, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Genre, error)
	List(ctx context.Context) ([]*domain.Genre, error)
	Update(ctx context.Context, id uuid.UUID, u GenreUpdates) (*domain.Genre, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
