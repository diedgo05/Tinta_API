package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tinta/identity/internal/user/application"
	"github.com/tinta/identity/internal/user/domain"
	"github.com/tinta/shared/httpx"
	"github.com/tinta/shared/middleware"
)

// Handler bundles the use cases needed by the HTTP layer.
// Wiring is performed in cmd/api/main.go.
type Handler struct {
	createUC *application.CreateUserUseCase
	getUC    *application.GetUserUseCase
	updateUC *application.UpdateUserUseCase
	deleteUC *application.DeleteUserUseCase
}

// NewHandler constructs the user HTTP handler.
func NewHandler(
	create *application.CreateUserUseCase,
	get *application.GetUserUseCase,
	update *application.UpdateUserUseCase,
	delete *application.DeleteUserUseCase,
) *Handler {
	return &Handler{
		createUC: create,
		getUC:    get,
		updateUC: update,
		deleteUC: delete,
	}
}

// Register adds the user routes to the router.
// /users/me is protected; the rest depends on the route.
func (h *Handler) Register(router fiber.Router, authMiddleware fiber.Handler) {
	router.Post("/users", h.create)
	router.Get("/users/me", authMiddleware, h.getMe)
	router.Get("/users/:id", h.getPublic)
	router.Patch("/users/me", authMiddleware, h.updateMe)
	router.Delete("/users/me", authMiddleware, h.deleteMe)
}

// create handles POST /users (registration).
func (h *Handler) create(c *fiber.Ctx) error {
	var body RegisterRequest
	if err := c.BodyParser(&body); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}

	user, err := h.createUC.Execute(c.Context(), application.CreateUserInput{
		Email:    body.Email,
		Password: body.Password,
		Name:     body.Name,
		Language: body.Language,
	})
	if err != nil {
		return mapUserError(c, err)
	}
	return httpx.Created(c, toUserResponse(user))
}

// getMe handles GET /users/me.
func (h *Handler) getMe(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return httpx.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
	}
	user, err := h.getUC.Execute(c.Context(), userID)
	if err != nil {
		return mapUserError(c, err)
	}
	return httpx.OK(c, toUserResponse(user))
}

// getPublic handles GET /users/:id (public profile).
func (h *Handler) getPublic(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid user id")
	}
	user, err := h.getUC.Execute(c.Context(), id)
	if err != nil {
		return mapUserError(c, err)
	}
	return httpx.OK(c, PublicUserResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
	})
}

// updateMe handles PATCH /users/me.
func (h *Handler) updateMe(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return httpx.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
	}

	var body UpdateUserRequest
	if err := c.BodyParser(&body); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}

	user, err := h.updateUC.Execute(c.Context(), userID, application.UpdateUserInput{
		Name:      body.Name,
		AvatarURL: body.AvatarURL,
		Language:  body.Language,
	})
	if err != nil {
		return mapUserError(c, err)
	}
	return httpx.OK(c, toUserResponse(user))
}

// deleteMe handles DELETE /users/me (ARCO cancellation).
func (h *Handler) deleteMe(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return httpx.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
	}
	if err := h.deleteUC.Execute(c.Context(), userID); err != nil {
		return mapUserError(c, err)
	}
	return httpx.NoContent(c)
}

// toUserResponse converts a domain.User into the HTTP DTO.
func toUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:            u.ID.String(),
		Email:         u.Email,
		Name:          u.Name,
		Role:          string(u.Role),
		EmailVerified: u.EmailVerified,
		AvatarURL:     u.AvatarURL,
		Language:      u.Language,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

// mapUserError translates domain errors into HTTP error responses.
func mapUserError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "USER_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return httpx.Error(c, fiber.StatusConflict, "EMAIL_EXISTS", err.Error())
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrPasswordTooWeak),
		errors.Is(err, domain.ErrNameTooShort),
		errors.Is(err, domain.ErrInvalidLanguage):
		return httpx.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
