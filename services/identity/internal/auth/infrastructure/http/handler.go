package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/tinta/identity/internal/auth/application"
	"github.com/tinta/identity/internal/auth/domain"
	"github.com/tinta/shared/httpx"
)

// Handler holds the auth use cases.
type Handler struct {
	loginUC   *application.LoginUseCase
	refreshUC *application.RefreshUseCase
	logoutUC  *application.LogoutUseCase
}

// NewHandler constructs the auth HTTP handler.
func NewHandler(
	login *application.LoginUseCase,
	refresh *application.RefreshUseCase,
	logout *application.LogoutUseCase,
) *Handler {
	return &Handler{loginUC: login, refreshUC: refresh, logoutUC: logout}
}

// Register adds the auth routes to the router.
func (h *Handler) Register(router fiber.Router) {
	router.Post("/auth/login", h.login)
	router.Post("/auth/refresh", h.refresh)
	router.Post("/auth/logout", h.logout)
}

// login handles POST /auth/login.
func (h *Handler) login(c *fiber.Ctx) error {
	var body LoginRequest
	if err := c.BodyParser(&body); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if body.Email == "" || body.Password == "" {
		return httpx.Error(c, fiber.StatusBadRequest, "MISSING_CREDENTIALS", "email and password are required")
	}

	out, err := h.loginUC.Execute(c.Context(), application.LoginInput{
		Email:     body.Email,
		Password:  body.Password,
		UserAgent: c.Get("User-Agent"),
		IPAddress: c.IP(),
	})
	if err != nil {
		return mapAuthError(c, err)
	}

	return httpx.OK(c, TokenPairResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		TokenType:    "Bearer",
	})
}

// refresh handles POST /auth/refresh.
func (h *Handler) refresh(c *fiber.Ctx) error {
	var body RefreshRequest
	if err := c.BodyParser(&body); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if body.RefreshToken == "" {
		return httpx.Error(c, fiber.StatusBadRequest, "MISSING_TOKEN", "refresh_token is required")
	}

	out, err := h.refreshUC.Execute(c.Context(), body.RefreshToken, c.Get("User-Agent"), c.IP())
	if err != nil {
		return mapAuthError(c, err)
	}

	return httpx.OK(c, TokenPairResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		TokenType:    "Bearer",
	})
}

// logout handles POST /auth/logout.
func (h *Handler) logout(c *fiber.Ctx) error {
	var body LogoutRequest
	if err := c.BodyParser(&body); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if body.RefreshToken == "" {
		return httpx.Error(c, fiber.StatusBadRequest, "MISSING_TOKEN", "refresh_token is required")
	}
	if err := h.logoutUC.Execute(c.Context(), body.RefreshToken); err != nil {
		return mapAuthError(c, err)
	}
	return httpx.NoContent(c)
}

// mapAuthError translates domain errors into HTTP responses.
func mapAuthError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return httpx.Error(c, fiber.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
	case errors.Is(err, domain.ErrInvalidToken),
		errors.Is(err, domain.ErrTokenExpired),
		errors.Is(err, domain.ErrTokenRevoked):
		return httpx.Error(c, fiber.StatusUnauthorized, "INVALID_TOKEN", err.Error())
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
