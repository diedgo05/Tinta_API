// Package http exposes the Book HTTP handlers.
package http

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tinta/catalog/internal/book/application"
	"github.com/tinta/catalog/internal/book/domain"
	"github.com/tinta/catalog/internal/book/ports"
	"github.com/tinta/shared/httpx"
)

type CreateBookRequest struct {
	GenreID       *string `json:"genre_id,omitempty"`
	Title         string  `json:"title"`
	Author        string  `json:"author"`
	ISBN          string  `json:"isbn,omitempty"`
	Synopsis      string  `json:"synopsis,omitempty"`
	CoverURL      string  `json:"cover_url,omitempty"`
	TotalPages    int     `json:"total_pages,omitempty"`
	License       string  `json:"license,omitempty"`
	Language      string  `json:"language,omitempty"`
	PublishedYear *int    `json:"published_year,omitempty"`
}

type UpdateBookRequest struct {
	GenreID       *string `json:"genre_id,omitempty"`
	Title         *string `json:"title,omitempty"`
	Author        *string `json:"author,omitempty"`
	ISBN          *string `json:"isbn,omitempty"`
	Synopsis      *string `json:"synopsis,omitempty"`
	CoverURL      *string `json:"cover_url,omitempty"`
	TotalPages    *int    `json:"total_pages,omitempty"`
	License       *string `json:"license,omitempty"`
	Language      *string `json:"language,omitempty"`
	PublishedYear *int    `json:"published_year,omitempty"`
}

type BookResponse struct {
	ID            string    `json:"id"`
	GenreID       *string   `json:"genre_id,omitempty"`
	Title         string    `json:"title"`
	Author        string    `json:"author"`
	ISBN          string    `json:"isbn,omitempty"`
	Synopsis      string    `json:"synopsis,omitempty"`
	CoverURL      string    `json:"cover_url,omitempty"`
	TotalPages    int       `json:"total_pages"`
	License       string    `json:"license"`
	Language      string    `json:"language"`
	PublishedYear *int      `json:"published_year,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PaginatedBooksResponse struct {
	Items    []BookResponse `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

type Handler struct {
	createUC *application.CreateBookUseCase
	getUC    *application.GetBookUseCase
	listUC   *application.ListBooksUseCase
	updateUC *application.UpdateBookUseCase
	deleteUC *application.DeleteBookUseCase
}

func NewHandler(c *application.CreateBookUseCase, g *application.GetBookUseCase, l *application.ListBooksUseCase,
	u *application.UpdateBookUseCase, d *application.DeleteBookUseCase) *Handler {
	return &Handler{createUC: c, getUC: g, listUC: l, updateUC: u, deleteUC: d}
}

func (h *Handler) Register(router fiber.Router, authMW fiber.Handler) {
	books := router.Group("/books")
	books.Get("/", h.list)            // listar es público
	books.Get("/:id", h.get)          // detalle es público
	books.Post("/", authMW, h.create) // crear requiere auth (admin)
	books.Patch("/:id", authMW, h.update)
	books.Delete("/:id", authMW, h.delete)
}

func (h *Handler) create(c *fiber.Ctx) error {
	var b CreateBookRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	var genreID *uuid.UUID
	if b.GenreID != nil && *b.GenreID != "" {
		id, err := uuid.Parse(*b.GenreID)
		if err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "INVALID_GENRE_ID", "invalid genre id")
		}
		genreID = &id
	}
	book, err := h.createUC.Execute(c.Context(), application.CreateBookInput{
		GenreID: genreID, Title: b.Title, Author: b.Author, ISBN: b.ISBN, Synopsis: b.Synopsis,
		CoverURL: b.CoverURL, TotalPages: b.TotalPages, License: b.License, Language: b.Language,
		PublishedYear: b.PublishedYear,
	})
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.Created(c, toResp(book))
}

func (h *Handler) get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid book id")
	}
	book, err := h.getUC.Execute(c.Context(), id)
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, toResp(book))
}

func (h *Handler) list(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	f := ports.ListFilter{Page: page, PageSize: pageSize, Search: c.Query("search", "")}
	if v := c.Query("genre_id"); v != "" {
		gid, err := uuid.Parse(v)
		if err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "INVALID_GENRE_ID", "invalid genre id")
		}
		f.GenreID = &gid
	}
	res, err := h.listUC.Execute(c.Context(), f)
	if err != nil {
		return mapErr(c, err)
	}
	items := make([]BookResponse, 0, len(res.Items))
	for _, b := range res.Items {
		items = append(items, toResp(b))
	}
	return httpx.OK(c, PaginatedBooksResponse{Items: items, Total: res.Total, Page: res.Page, PageSize: res.PageSize})
}

func (h *Handler) update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid book id")
	}
	var b UpdateBookRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	var genreID *uuid.UUID
	if b.GenreID != nil && *b.GenreID != "" {
		gid, err := uuid.Parse(*b.GenreID)
		if err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "INVALID_GENRE_ID", "invalid genre id")
		}
		genreID = &gid
	}
	book, err := h.updateUC.Execute(c.Context(), id, application.UpdateBookInput{
		GenreID: genreID, Title: b.Title, Author: b.Author, ISBN: b.ISBN, Synopsis: b.Synopsis,
		CoverURL: b.CoverURL, TotalPages: b.TotalPages, License: b.License, Language: b.Language,
		PublishedYear: b.PublishedYear,
	})
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, toResp(book))
}

func (h *Handler) delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid book id")
	}
	if err := h.deleteUC.Execute(c.Context(), id); err != nil {
		return mapErr(c, err)
	}
	return httpx.NoContent(c)
}

func toResp(b *domain.Book) BookResponse {
	var gid *string
	if b.GenreID != nil {
		s := b.GenreID.String()
		gid = &s
	}
	return BookResponse{
		ID: b.ID.String(), GenreID: gid, Title: b.Title, Author: b.Author, ISBN: b.ISBN,
		Synopsis: b.Synopsis, CoverURL: b.CoverURL, TotalPages: b.TotalPages, License: string(b.License),
		Language: b.Language, PublishedYear: b.PublishedYear, CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt,
	}
}

func mapErr(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrBookNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "BOOK_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrInvalidTitle), errors.Is(err, domain.ErrInvalidAuthor),
		errors.Is(err, domain.ErrInvalidLicense), errors.Is(err, domain.ErrInvalidPageSize):
		return httpx.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
