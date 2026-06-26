// Package http exposes the password-reset HTTP handlers.
package http

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tinta/identity/internal/passwordreset/application"
	"github.com/tinta/identity/internal/passwordreset/domain"
	"github.com/tinta/shared/httpx"
)

type RequestRequest struct {
	Email string `json:"email"`
}

// RequestResponse echoes the code for MVP (no SMTP yet).
type RequestResponse struct {
	Message   string     `json:"message"`
	Code      string     `json:"code,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type ConfirmRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

type Handler struct {
	requestUC *application.RequestResetUseCase
	confirmUC *application.ConfirmResetUseCase
}

func NewHandler(req *application.RequestResetUseCase, conf *application.ConfirmResetUseCase) *Handler {
	return &Handler{requestUC: req, confirmUC: conf}
}

// Register adds the public password-reset endpoints (no auth required).
func (h *Handler) Register(router fiber.Router) {
	g := router.Group("/auth/password-reset")
	g.Post("/request", h.request)
	g.Post("/confirm", h.confirm)
}

func (h *Handler) request(c *fiber.Ctx) error {
	var b RequestRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if b.Email == "" {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_EMAIL", "email is required")
	}
	st, err := h.requestUC.Execute(c.Context(), application.RequestResetInput{Email: b.Email})
	if err != nil {
		return mapErr(c, err)
	}
	// Always 200 with generic message (anti-enumeration). Code/expires only
	// included when the email matched a real user — for MVP without SMTP.
	resp := RequestResponse{
		Message: "if an account exists for that email, a reset code was issued",
	}
	if st.Code != "" {
		resp.Code = st.Code
		resp.ExpiresAt = st.ExpiresAt
	}
	return httpx.OK(c, resp)
}

func (h *Handler) confirm(c *fiber.Ctx) error {
	var b ConfirmRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	err := h.confirmUC.Execute(c.Context(), application.ConfirmResetInput{
		Email: b.Email, Code: b.Code, NewPassword: b.NewPassword,
	})
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, fiber.Map{"message": "password updated"})
}

func mapErr(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound), errors.Is(err, domain.ErrInvalidCode):
		// Treat both the same way to avoid enumeration on the confirm endpoint.
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_CODE", "invalid email or code")
	case errors.Is(err, domain.ErrExpiredCode):
		return httpx.Error(c, fiber.StatusBadRequest, "EXPIRED_CODE", err.Error())
	case errors.Is(err, domain.ErrNoCodeRequested):
		return httpx.Error(c, fiber.StatusBadRequest, "NO_CODE_REQUESTED", err.Error())
	case errors.Is(err, domain.ErrWeakPassword):
		return httpx.Error(c, fiber.StatusBadRequest, "WEAK_PASSWORD", err.Error())
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
