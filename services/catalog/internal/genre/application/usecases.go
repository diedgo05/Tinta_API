// Package application contains the Genre use cases.
package application

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/tinta/catalog/internal/genre/domain"
	"github.com/tinta/catalog/internal/genre/ports"
)

type CreateGenreInput struct {
	Name        string
	Description string
}

type CreateGenreUseCase struct{ repo ports.GenreRepository }

func NewCreateGenreUseCase(r ports.GenreRepository) *CreateGenreUseCase {
	return &CreateGenreUseCase{repo: r}
}

func (uc *CreateGenreUseCase) Execute(ctx context.Context, in CreateGenreInput) (*domain.Genre, error) {
	if err := domain.ValidateName(in.Name); err != nil {
		return nil, err
	}
	return uc.repo.Create(ctx, &domain.Genre{
		Name:        strings.TrimSpace(in.Name),
		Slug:        domain.Slugify(in.Name),
		Description: in.Description,
	})
}

type GetGenreUseCase struct{ repo ports.GenreRepository }

func NewGetGenreUseCase(r ports.GenreRepository) *GetGenreUseCase { return &GetGenreUseCase{repo: r} }

func (uc *GetGenreUseCase) Execute(ctx context.Context, id uuid.UUID) (*domain.Genre, error) {
	return uc.repo.GetByID(ctx, id)
}

type ListGenresUseCase struct{ repo ports.GenreRepository }

func NewListGenresUseCase(r ports.GenreRepository) *ListGenresUseCase {
	return &ListGenresUseCase{repo: r}
}

func (uc *ListGenresUseCase) Execute(ctx context.Context) ([]*domain.Genre, error) {
	return uc.repo.List(ctx)
}

type UpdateGenreInput struct {
	Name        *string
	Description *string
}

type UpdateGenreUseCase struct{ repo ports.GenreRepository }

func NewUpdateGenreUseCase(r ports.GenreRepository) *UpdateGenreUseCase {
	return &UpdateGenreUseCase{repo: r}
}

func (uc *UpdateGenreUseCase) Execute(ctx context.Context, id uuid.UUID, in UpdateGenreInput) (*domain.Genre, error) {
	if in.Name != nil {
		if err := domain.ValidateName(*in.Name); err != nil {
			return nil, err
		}
	}
	return uc.repo.Update(ctx, id, ports.GenreUpdates{Name: in.Name, Description: in.Description})
}

type DeleteGenreUseCase struct{ repo ports.GenreRepository }

func NewDeleteGenreUseCase(r ports.GenreRepository) *DeleteGenreUseCase {
	return &DeleteGenreUseCase{repo: r}
}

func (uc *DeleteGenreUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}
