// Package application contains the Book use cases.
package application

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/tinta/catalog/internal/book/domain"
	"github.com/tinta/catalog/internal/book/ports"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// ---------- Create ----------

type CreateBookInput struct {
	GenreID       *uuid.UUID
	Title         string
	Author        string
	ISBN          string
	Synopsis      string
	CoverURL      string
	TotalPages    int
	License       string
	Language      string
	PublishedYear *int
}

type CreateBookUseCase struct{ repo ports.BookRepository }

func NewCreateBookUseCase(r ports.BookRepository) *CreateBookUseCase {
	return &CreateBookUseCase{repo: r}
}

func (uc *CreateBookUseCase) Execute(ctx context.Context, in CreateBookInput) (*domain.Book, error) {
	if err := domain.ValidateTitle(in.Title); err != nil {
		return nil, err
	}
	if err := domain.ValidateAuthor(in.Author); err != nil {
		return nil, err
	}
	license := domain.License(in.License)
	if in.License == "" {
		license = domain.LicenseUnknown
	}
	if !license.IsValid() {
		return nil, domain.ErrInvalidLicense
	}
	lang := in.Language
	if lang == "" {
		lang = "es"
	}
	return uc.repo.Create(ctx, &domain.Book{
		GenreID:       in.GenreID,
		Title:         strings.TrimSpace(in.Title),
		Author:        strings.TrimSpace(in.Author),
		ISBN:          strings.TrimSpace(in.ISBN),
		Synopsis:      in.Synopsis,
		CoverURL:      in.CoverURL,
		TotalPages:    in.TotalPages,
		License:       license,
		Language:      lang,
		PublishedYear: in.PublishedYear,
	})
}

// ---------- Get ----------

type GetBookUseCase struct{ repo ports.BookRepository }

func NewGetBookUseCase(r ports.BookRepository) *GetBookUseCase { return &GetBookUseCase{repo: r} }

func (uc *GetBookUseCase) Execute(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	return uc.repo.GetByID(ctx, id)
}

// ---------- List ----------

type ListBooksUseCase struct{ repo ports.BookRepository }

func NewListBooksUseCase(r ports.BookRepository) *ListBooksUseCase { return &ListBooksUseCase{repo: r} }

func (uc *ListBooksUseCase) Execute(ctx context.Context, f ports.ListFilter) (*ports.ListResult, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = defaultPageSize
	}
	if f.PageSize > maxPageSize {
		return nil, domain.ErrInvalidPageSize
	}
	return uc.repo.List(ctx, f)
}

// ---------- Update ----------

type UpdateBookInput struct {
	GenreID       *uuid.UUID
	Title         *string
	Author        *string
	ISBN          *string
	Synopsis      *string
	CoverURL      *string
	TotalPages    *int
	License       *string
	Language      *string
	PublishedYear *int
}

type UpdateBookUseCase struct{ repo ports.BookRepository }

func NewUpdateBookUseCase(r ports.BookRepository) *UpdateBookUseCase {
	return &UpdateBookUseCase{repo: r}
}

func (uc *UpdateBookUseCase) Execute(ctx context.Context, id uuid.UUID, in UpdateBookInput) (*domain.Book, error) {
	if in.Title != nil {
		if err := domain.ValidateTitle(*in.Title); err != nil {
			return nil, err
		}
	}
	if in.Author != nil {
		if err := domain.ValidateAuthor(*in.Author); err != nil {
			return nil, err
		}
	}
	var license *domain.License
	if in.License != nil {
		l := domain.License(*in.License)
		if !l.IsValid() {
			return nil, domain.ErrInvalidLicense
		}
		license = &l
	}
	return uc.repo.Update(ctx, id, ports.BookUpdates{
		GenreID:       in.GenreID,
		Title:         in.Title,
		Author:        in.Author,
		ISBN:          in.ISBN,
		Synopsis:      in.Synopsis,
		CoverURL:      in.CoverURL,
		TotalPages:    in.TotalPages,
		License:       license,
		Language:      in.Language,
		PublishedYear: in.PublishedYear,
	})
}

// ---------- Delete ----------

type DeleteBookUseCase struct{ repo ports.BookRepository }

func NewDeleteBookUseCase(r ports.BookRepository) *DeleteBookUseCase {
	return &DeleteBookUseCase{repo: r}
}

func (uc *DeleteBookUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.repo.SoftDelete(ctx, id)
}
