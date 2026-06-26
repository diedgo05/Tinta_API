// Package application contains Progress use cases.
package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/reading/internal/progress/domain"
	"github.com/tinta/reading/internal/progress/ports"
)

// StartReading or update progress (idempotent upsert).

type StartReadingInput struct {
	UserID      uuid.UUID
	BookID      uuid.UUID
	CurrentPage int
}

type StartReadingUseCase struct{ repo ports.ProgressRepository }

func NewStartReadingUseCase(r ports.ProgressRepository) *StartReadingUseCase {
	return &StartReadingUseCase{repo: r}
}

func (uc *StartReadingUseCase) Execute(ctx context.Context, in StartReadingInput) (*domain.Progress, error) {
	if in.CurrentPage < 0 {
		return nil, domain.ErrInvalidPage
	}
	return uc.repo.UpsertProgress(ctx, &domain.Progress{
		UserID:      in.UserID,
		BookID:      in.BookID,
		CurrentPage: in.CurrentPage,
		Status:      domain.StatusReading,
	})
}

type GetProgressUseCase struct{ repo ports.ProgressRepository }

func NewGetProgressUseCase(r ports.ProgressRepository) *GetProgressUseCase {
	return &GetProgressUseCase{repo: r}
}

func (uc *GetProgressUseCase) Execute(ctx context.Context, userID, bookID uuid.UUID) (*domain.Progress, error) {
	return uc.repo.GetByUserAndBook(ctx, userID, bookID)
}

type ListProgressUseCase struct{ repo ports.ProgressRepository }

func NewListProgressUseCase(r ports.ProgressRepository) *ListProgressUseCase {
	return &ListProgressUseCase{repo: r}
}

func (uc *ListProgressUseCase) Execute(ctx context.Context, userID uuid.UUID, status *domain.Status) ([]*domain.Progress, error) {
	if status != nil && !status.IsValid() {
		return nil, domain.ErrInvalidStatus
	}
	return uc.repo.ListByUser(ctx, userID, status)
}

type UpdateProgressInput struct {
	CurrentPage  *int
	TotalTimeMin *int
	Status       *string
}

type UpdateProgressUseCase struct{ repo ports.ProgressRepository }

func NewUpdateProgressUseCase(r ports.ProgressRepository) *UpdateProgressUseCase {
	return &UpdateProgressUseCase{repo: r}
}

func (uc *UpdateProgressUseCase) Execute(ctx context.Context, userID, bookID uuid.UUID, in UpdateProgressInput) (*domain.Progress, error) {
	if in.CurrentPage != nil && *in.CurrentPage < 0 {
		return nil, domain.ErrInvalidPage
	}
	var status *domain.Status
	if in.Status != nil {
		s := domain.Status(*in.Status)
		if !s.IsValid() {
			return nil, domain.ErrInvalidStatus
		}
		status = &s
	}
	return uc.repo.UpdateProgress(ctx, userID, bookID, ports.ProgressUpdates{
		CurrentPage: in.CurrentPage, TotalTimeMin: in.TotalTimeMin, Status: status,
	})
}

type DeleteProgressUseCase struct{ repo ports.ProgressRepository }

func NewDeleteProgressUseCase(r ports.ProgressRepository) *DeleteProgressUseCase {
	return &DeleteProgressUseCase{repo: r}
}

func (uc *DeleteProgressUseCase) Execute(ctx context.Context, userID, bookID uuid.UUID) error {
	return uc.repo.Delete(ctx, userID, bookID)
}
