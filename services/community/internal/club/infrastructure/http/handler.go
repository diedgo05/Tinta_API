package http

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tinta/community/internal/club/application"
	"github.com/tinta/community/internal/club/domain"
	"github.com/tinta/community/internal/club/ports"
	"github.com/tinta/shared/httpx"
	"github.com/tinta/shared/middleware"
)

// Handler bundles the use cases needed by the HTTP layer.
type Handler struct {
	createUC *application.CreateClubUseCase
	listUC   *application.ListClubsUseCase
	getUC    *application.GetClubUseCase
	updateUC *application.UpdateClubUseCase
	deleteUC *application.DeleteClubUseCase
}

// NewHandler constructs the club HTTP handler.
func NewHandler(
	create *application.CreateClubUseCase,
	list *application.ListClubsUseCase,
	get *application.GetClubUseCase,
	update *application.UpdateClubUseCase,
	delete *application.DeleteClubUseCase,
) *Handler {
	return &Handler{
		createUC: create,
		listUC:   list,
		getUC:    get,
		updateUC: update,
		deleteUC: delete,
	}
}

// Register adds the club routes to the router.
// All routes require authentication (the API gateway / mobile app sends JWT).
func (h *Handler) Register(router fiber.Router, authMiddleware fiber.Handler) {
	clubs := router.Group("/clubs", authMiddleware)
	clubs.Post("/", h.create)
	clubs.Get("/", h.list)
	clubs.Get("/:id", h.get)
	clubs.Patch("/:id", h.update)
	clubs.Delete("/:id", h.delete)
}

// create handles POST /clubs.
func (h *Handler) create(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return httpx.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
	}

	var body CreateClubRequest
	if err := c.BodyParser(&body); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}

	var bookID *uuid.UUID
	if body.BookID != nil && *body.BookID != "" {
		id, err := uuid.Parse(*body.BookID)
		if err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BOOK_ID", "invalid book id")
		}
		bookID = &id
	}

	club, err := h.createUC.Execute(c.Context(), application.CreateClubInput{
		CreatorID:   userID,
		Name:        body.Name,
		Description: body.Description,
		BookID:      bookID,
		IsPrivate:   body.IsPrivate,
	})
	if err != nil {
		return mapClubError(c, err)
	}
	return httpx.Created(c, toClubResponse(club))
}

// list handles GET /clubs?page=1&page_size=20&creator_id=...&book_id=...
func (h *Handler) list(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	filter := ports.ListFilter{
		Page:     page,
		PageSize: pageSize,
	}

	if v := c.Query("creator_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "INVALID_CREATOR_ID", "invalid creator id")
		}
		filter.CreatorID = &id
		// If the user is asking for THEIR OWN clubs, show private ones too.
		if userID, ok := middleware.UserIDFromContext(c); ok && userID == id {
			filter.IncludeAll = true
		}
	}
	if v := c.Query("book_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BOOK_ID", "invalid book id")
		}
		filter.BookID = &id
	}

	result, err := h.listUC.Execute(c.Context(), filter)
	if err != nil {
		return mapClubError(c, err)
	}

	items := make([]ClubResponse, 0, len(result.Items))
	for _, club := range result.Items {
		items = append(items, toClubResponse(club))
	}
	return httpx.OK(c, PaginatedClubsResponse{
		Items:    items,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// get handles GET /clubs/{id}.
func (h *Handler) get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid club id")
	}
	club, err := h.getUC.Execute(c.Context(), id)
	if err != nil {
		return mapClubError(c, err)
	}
	return httpx.OK(c, toClubResponse(club))
}

// update handles PATCH /clubs/{id}.
func (h *Handler) update(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return httpx.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid club id")
	}

	var body UpdateClubRequest
	if err := c.BodyParser(&body); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}

	var bookID *uuid.UUID
	if body.BookID != nil && *body.BookID != "" {
		parsed, err := uuid.Parse(*body.BookID)
		if err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BOOK_ID", "invalid book id")
		}
		bookID = &parsed
	}

	club, err := h.updateUC.Execute(c.Context(), application.UpdateClubInput{
		ClubID:      id,
		RequesterID: userID,
		Name:        body.Name,
		Description: body.Description,
		BookID:      bookID,
		IsPrivate:   body.IsPrivate,
	})
	if err != nil {
		return mapClubError(c, err)
	}
	return httpx.OK(c, toClubResponse(club))
}

// delete handles DELETE /clubs/{id}.
func (h *Handler) delete(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return httpx.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid club id")
	}
	if err := h.deleteUC.Execute(c.Context(), id, userID); err != nil {
		return mapClubError(c, err)
	}
	return httpx.NoContent(c)
}

func toClubResponse(c *domain.Club) ClubResponse {
	var bookID *string
	if c.BookID != nil {
		s := c.BookID.String()
		bookID = &s
	}
	return ClubResponse{
		ID:          c.ID.String(),
		CreatorID:   c.CreatorID.String(),
		BookID:      bookID,
		Name:        c.Name,
		Description: c.Description,
		IsPrivate:   c.IsPrivate,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func mapClubError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrClubNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "CLUB_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrNotAuthorized):
		return httpx.Error(c, fiber.StatusForbidden, "FORBIDDEN", err.Error())
	case errors.Is(err, domain.ErrInvalidClubName),
		errors.Is(err, domain.ErrInvalidPageSize):
		return httpx.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
