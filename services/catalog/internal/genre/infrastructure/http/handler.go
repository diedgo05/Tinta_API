// Package http exposes the Genre HTTP handlers.
package http

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tinta/catalog/internal/genre/application"
	"github.com/tinta/catalog/internal/genre/domain"
	"github.com/tinta/shared/httpx"
)

type CreateGenreRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type UpdateGenreRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type GenreResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Handler struct {
	createUC *application.CreateGenreUseCase
	getUC    *application.GetGenreUseCase
	listUC   *application.ListGenresUseCase
	updateUC *application.UpdateGenreUseCase
	deleteUC *application.DeleteGenreUseCase
}

func NewHandler(c *application.CreateGenreUseCase, g *application.GetGenreUseCase, l *application.ListGenresUseCase,
	u *application.UpdateGenreUseCase, d *application.DeleteGenreUseCase) *Handler {
	return &Handler{createUC: c, getUC: g, listUC: l, updateUC: u, deleteUC: d}
}

func (h *Handler) Register(router fiber.Router, authMW fiber.Handler) {
	g := router.Group("/genres")
	g.Get("/", h.list)
	g.Get("/:id", h.get)
	g.Post("/", authMW, h.create)
	g.Patch("/:id", authMW, h.update)
	g.Delete("/:id", authMW, h.delete)
}

func (h *Handler) create(c *fiber.Ctx) error {
	var b CreateGenreRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	g, err := h.createUC.Execute(c.Context(), application.CreateGenreInput{Name: b.Name, Description: b.Description})
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.Created(c, toResp(g))
}

func (h *Handler) get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid genre id")
	}
	g, err := h.getUC.Execute(c.Context(), id)
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, toResp(g))
}

func (h *Handler) list(c *fiber.Ctx) error {
	list, err := h.listUC.Execute(c.Context())
	if err != nil {
		return mapErr(c, err)
	}
	out := make([]GenreResponse, 0, len(list))
	for _, g := range list {
		out = append(out, toResp(g))
	}
	return httpx.OK(c, fiber.Map{"items": out, "total": len(out)})
}

func (h *Handler) update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid genre id")
	}
	var b UpdateGenreRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	g, err := h.updateUC.Execute(c.Context(), id, application.UpdateGenreInput{Name: b.Name, Description: b.Description})
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, toResp(g))
}

func (h *Handler) delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid genre id")
	}
	if err := h.deleteUC.Execute(c.Context(), id); err != nil {
		return mapErr(c, err)
	}
	return httpx.NoContent(c)
}

func toResp(g *domain.Genre) GenreResponse {
	return GenreResponse{
		ID: g.ID.String(), Name: g.Name, Slug: g.Slug, Description: g.Description,
		CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt,
	}
}

func mapErr(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrGenreNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "GENRE_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrGenreAlreadyExists):
		return httpx.Error(c, fiber.StatusConflict, "GENRE_EXISTS", err.Error())
	case errors.Is(err, domain.ErrInvalidGenreName):
		return httpx.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
