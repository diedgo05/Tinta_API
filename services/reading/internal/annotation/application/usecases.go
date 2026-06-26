// Package application contains Annotation use cases.
package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/reading/internal/annotation/domain"
	"github.com/tinta/reading/internal/annotation/ports"
)

// ---------- Create ----------

type CreateAnnotationInput struct {
	UserID          uuid.UUID
	BookID          *uuid.UUID
	PersonalDocID   *uuid.UUID
	Page            int
	HighlightedText string
	PersonalNote    string
	Color           string
}

type CreateAnnotationUseCase struct{ repo ports.AnnotationRepository }

func NewCreateAnnotationUseCase(r ports.AnnotationRepository) *CreateAnnotationUseCase {
	return &CreateAnnotationUseCase{repo: r}
}

func (uc *CreateAnnotationUseCase) Execute(ctx context.Context, in CreateAnnotationInput) (*domain.Annotation, error) {
	if err := domain.ValidateHighlight(in.HighlightedText); err != nil {
		return nil, err
	}
	if in.BookID == nil && in.PersonalDocID == nil {
		return nil, domain.ErrNoTarget
	}
	color := in.Color
	if color == "" {
		color = "yellow"
	}
	return uc.repo.Create(ctx, &domain.Annotation{
		UserID: in.UserID, BookID: in.BookID, PersonalDocID: in.PersonalDocID,
		Page: in.Page, HighlightedText: in.HighlightedText, PersonalNote: in.PersonalNote, Color: color,
	})
}

// ---------- List ----------

type ListAnnotationsUseCase struct{ repo ports.AnnotationRepository }

func NewListAnnotationsUseCase(r ports.AnnotationRepository) *ListAnnotationsUseCase {
	return &ListAnnotationsUseCase{repo: r}
}

func (uc *ListAnnotationsUseCase) Execute(ctx context.Context, f ports.ListFilter) ([]*domain.Annotation, error) {
	return uc.repo.List(ctx, f)
}

// ---------- Update ----------

type UpdateAnnotationInput struct {
	AnnotationID uuid.UUID
	RequesterID  uuid.UUID
	PersonalNote *string
	Color        *string
	Page         *int
}

type UpdateAnnotationUseCase struct{ repo ports.AnnotationRepository }

func NewUpdateAnnotationUseCase(r ports.AnnotationRepository) *UpdateAnnotationUseCase {
	return &UpdateAnnotationUseCase{repo: r}
}

func (uc *UpdateAnnotationUseCase) Execute(ctx context.Context, in UpdateAnnotationInput) (*domain.Annotation, error) {
	existing, err := uc.repo.GetByID(ctx, in.AnnotationID)
	if err != nil {
		return nil, err
	}
	if !existing.CanBeManagedBy(in.RequesterID) {
		return nil, domain.ErrNotAuthorized
	}
	return uc.repo.Update(ctx, in.AnnotationID, ports.AnnotationUpdates{
		PersonalNote: in.PersonalNote, Color: in.Color, Page: in.Page,
	})
}

// ---------- Delete ----------

type DeleteAnnotationUseCase struct{ repo ports.AnnotationRepository }

func NewDeleteAnnotationUseCase(r ports.AnnotationRepository) *DeleteAnnotationUseCase {
	return &DeleteAnnotationUseCase{repo: r}
}

func (uc *DeleteAnnotationUseCase) Execute(ctx context.Context, id, requesterID uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !existing.CanBeManagedBy(requesterID) {
		return domain.ErrNotAuthorized
	}
	return uc.repo.SoftDelete(ctx, id)
}
