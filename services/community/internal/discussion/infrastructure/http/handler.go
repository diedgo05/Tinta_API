// Package http exposes the Discussion HTTP handlers.
package http

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tinta/community/internal/discussion/application"
	"github.com/tinta/community/internal/discussion/domain"
	"github.com/tinta/community/internal/discussion/ports"
	"github.com/tinta/shared/httpx"
	"github.com/tinta/shared/middleware"
)

type PostDiscussionRequest struct {
	ChapterNumber *int   `json:"chapter_number,omitempty"`
	Content       string `json:"content"`
}

type UpdateDiscussionRequest struct {
	Content string `json:"content"`
}

type DiscussionResponse struct {
	ID            string    `json:"id"`
	ClubID        string    `json:"club_id"`
	UserID        string    `json:"user_id"`
	ChapterNumber *int      `json:"chapter_number,omitempty"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PaginatedResponse struct {
	Items    []DiscussionResponse `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

type Handler struct {
	postUC   *application.PostDiscussionUseCase
	listUC   *application.ListDiscussionsUseCase
	updateUC *application.UpdateDiscussionUseCase
	deleteUC *application.DeleteDiscussionUseCase
}

func NewHandler(p *application.PostDiscussionUseCase, l *application.ListDiscussionsUseCase,
	u *application.UpdateDiscussionUseCase, d *application.DeleteDiscussionUseCase) *Handler {
	return &Handler{postUC: p, listUC: l, updateUC: u, deleteUC: d}
}

func (h *Handler) Register(router fiber.Router, authMW fiber.Handler) {
	g := router.Group("", authMW)
	g.Post("/clubs/:club_id/discussions", h.post)
	g.Get("/clubs/:club_id/discussions", h.list)
	g.Patch("/discussions/:id", h.update)
	g.Delete("/discussions/:id", h.delete)
}

func (h *Handler) post(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	clubID, err := uuid.Parse(c.Params("club_id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_CLUB_ID", "invalid club id")
	}
	var b PostDiscussionRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid body")
	}
	d, err := h.postUC.Execute(c.Context(), application.PostDiscussionInput{
		ClubID: clubID, UserID: userID, ChapterNumber: b.ChapterNumber, Content: b.Content,
	})
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.Created(c, toResp(d))
}

func (h *Handler) list(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	clubID, err := uuid.Parse(c.Params("club_id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_CLUB_ID", "invalid club id")
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "50"))

	f := ports.ListFilter{ClubID: clubID, Page: page, PageSize: pageSize}
	if v := c.Query("chapter_number"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.ChapterNumber = &n
		}
	}
	res, err := h.listUC.Execute(c.Context(), userID, f)
	if err != nil {
		return mapErr(c, err)
	}
	items := make([]DiscussionResponse, 0, len(res.Items))
	for _, d := range res.Items {
		items = append(items, toResp(d))
	}
	return httpx.OK(c, PaginatedResponse{Items: items, Total: res.Total, Page: res.Page, PageSize: res.PageSize})
}

func (h *Handler) update(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid discussion id")
	}
	var b UpdateDiscussionRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid body")
	}
	d, err := h.updateUC.Execute(c.Context(), application.UpdateDiscussionInput{
		DiscussionID: id, RequesterID: userID, Content: b.Content,
	})
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, toResp(d))
}

func (h *Handler) delete(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid discussion id")
	}
	if err := h.deleteUC.Execute(c.Context(), id, userID); err != nil {
		return mapErr(c, err)
	}
	return httpx.NoContent(c)
}

func toResp(d *domain.Discussion) DiscussionResponse {
	return DiscussionResponse{
		ID: d.ID.String(), ClubID: d.ClubID.String(), UserID: d.UserID.String(),
		ChapterNumber: d.ChapterNumber, Content: d.Content,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

func mapErr(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrDiscussionNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "DISCUSSION_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrNotAuthorized):
		return httpx.Error(c, fiber.StatusForbidden, "FORBIDDEN", err.Error())
	case errors.Is(err, domain.ErrEmptyContent), errors.Is(err, domain.ErrContentTooLong):
		return httpx.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
