// Package ports declares the ProgressRepository interface.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/reading/internal/progress/domain"
)

type ProgressUpdates struct {
	CurrentPage  *int
	TotalTimeMin *int
	Status       *domain.Status
}

type ProgressRepository interface {
	UpsertProgress(ctx context.Context, p *domain.Progress) (*domain.Progress, error)
	GetByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (*domain.Progress, error)
	ListByUser(ctx context.Context, userID uuid.UUID, status *domain.Status) ([]*domain.Progress, error)
	UpdateProgress(ctx context.Context, userID, bookID uuid.UUID, u ProgressUpdates) (*domain.Progress, error)
	Delete(ctx context.Context, userID, bookID uuid.UUID) error
}
