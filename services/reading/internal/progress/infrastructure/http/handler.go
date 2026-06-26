// Package http exposes the Progress HTTP handlers.
package http

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tinta/reading/internal/progress/application"
	"github.com/tinta/reading/internal/progress/domain"
	"github.com/tinta/shared/httpx"
	"github.com/tinta/shared/middleware"
)

type StartReadingRequest struct {
	BookID      string `json:"book_id"`
	CurrentPage int    `json:"current_page,omitempty"`
}

type UpdateProgressRequest struct {
	CurrentPage  *int    `json:"current_page,omitempty"`
	TotalTimeMin *int    `json:"total_time_min,omitempty"`
	Status       *string `json:"status,omitempty"`
}

type ProgressResponse struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	BookID       string     `json:"book_id"`
	CurrentPage  int        `json:"current_page"`
	TotalTimeMin int        `json:"total_time_min"`
	Status       string     `json:"status"`
	StartedAt    time.Time  `json:"started_at"`
	LastReadAt   time.Time  `json:"last_read_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}

type Handler struct {
	startUC  *application.StartReadingUseCase
	getUC    *application.GetProgressUseCase
	listUC   *application.ListProgressUseCase
	updateUC *application.UpdateProgressUseCase
	deleteUC *application.DeleteProgressUseCase
}

func NewHandler(s *application.StartReadingUseCase, g *application.GetProgressUseCase, l *application.ListProgressUseCase,
	u *application.UpdateProgressUseCase, d *application.DeleteProgressUseCase) *Handler {
	return &Handler{startUC: s, getUC: g, listUC: l, updateUC: u, deleteUC: d}
}

func (h *Handler) Register(router fiber.Router, authMW fiber.Handler) {
	p := router.Group("/reading", authMW)
	p.Post("/", h.start)            // start/upsert
	p.Get("/", h.list)              // list my progress
	p.Get("/:book_id", h.get)       // get progress for a book
	p.Patch("/:book_id", h.update)  // update progress
	p.Delete("/:book_id", h.delete) // delete
}

func (h *Handler) start(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	var b StartReadingRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid body")
	}
	bookID, err := uuid.Parse(b.BookID)
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BOOK_ID", "invalid book id")
	}
	p, err := h.startUC.Execute(c.Context(), application.StartReadingInput{UserID: userID, BookID: bookID, CurrentPage: b.CurrentPage})
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, toResp(p))
}

func (h *Handler) get(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	bookID, err := uuid.Parse(c.Params("book_id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BOOK_ID", "invalid book id")
	}
	p, err := h.getUC.Execute(c.Context(), userID, bookID)
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, toResp(p))
}

func (h *Handler) list(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	var status *domain.Status
	if v := c.Query("status"); v != "" {
		s := domain.Status(v)
		status = &s
	}
	items, err := h.listUC.Execute(c.Context(), userID, status)
	if err != nil {
		return mapErr(c, err)
	}
	out := make([]ProgressResponse, 0, len(items))
	for _, p := range items {
		out = append(out, toResp(p))
	}
	return httpx.OK(c, fiber.Map{"items": out, "total": len(out)})
}

func (h *Handler) update(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	bookID, err := uuid.Parse(c.Params("book_id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BOOK_ID", "invalid book id")
	}
	var b UpdateProgressRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid body")
	}
	p, err := h.updateUC.Execute(c.Context(), userID, bookID, application.UpdateProgressInput{
		CurrentPage: b.CurrentPage, TotalTimeMin: b.TotalTimeMin, Status: b.Status,
	})
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, toResp(p))
}

func (h *Handler) delete(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	bookID, err := uuid.Parse(c.Params("book_id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BOOK_ID", "invalid book id")
	}
	if err := h.deleteUC.Execute(c.Context(), userID, bookID); err != nil {
		return mapErr(c, err)
	}
	return httpx.NoContent(c)
}

func toResp(p *domain.Progress) ProgressResponse {
	return ProgressResponse{
		ID: p.ID.String(), UserID: p.UserID.String(), BookID: p.BookID.String(),
		CurrentPage: p.CurrentPage, TotalTimeMin: p.TotalTimeMin, Status: string(p.Status),
		StartedAt: p.StartedAt, LastReadAt: p.LastReadAt, FinishedAt: p.FinishedAt,
	}
}

func mapErr(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrProgressNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "PROGRESS_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrInvalidStatus), errors.Is(err, domain.ErrInvalidPage):
		return httpx.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
