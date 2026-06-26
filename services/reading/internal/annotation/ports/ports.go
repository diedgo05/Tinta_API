// Package ports declares the AnnotationRepository interface.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/reading/internal/annotation/domain"
)

type ListFilter struct {
	UserID        uuid.UUID
	BookID        *uuid.UUID
	PersonalDocID *uuid.UUID
}

type AnnotationUpdates struct {
	PersonalNote *string
	Color        *string
	Page         *int
}

type AnnotationRepository interface {
	Create(ctx context.Context, a *domain.Annotation) (*domain.Annotation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Annotation, error)
	List(ctx context.Context, f ListFilter) ([]*domain.Annotation, error)
	Update(ctx context.Context, id uuid.UUID, u AnnotationUpdates) (*domain.Annotation, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
