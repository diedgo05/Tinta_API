// Package http exposes verification HTTP handlers.
package http

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/tinta/identity/internal/verification/application"
	"github.com/tinta/identity/internal/verification/domain"
	"github.com/tinta/shared/httpx"
	"github.com/tinta/shared/middleware"
)

type VerifyRequest struct {
	Code string `json:"code"`
}

// RequestCodeResponse echoes the code back since this MVP has no SMTP.
// In production, only `expires_at` and a generic message would be returned.
type RequestCodeResponse struct {
	Message   string    `json:"message"`
	Code      string    `json:"code,omitempty"` // echo for MVP
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

type Handler struct {
	requestUC *application.RequestCodeUseCase
	verifyUC  *application.VerifyCodeUseCase
}

func NewHandler(req *application.RequestCodeUseCase, ver *application.VerifyCodeUseCase) *Handler {
	return &Handler{requestUC: req, verifyUC: ver}
}

// Register adds verification endpoints. All require auth (user logged in).
func (h *Handler) Register(router fiber.Router, authMW fiber.Handler) {
	g := router.Group("/auth/verification", authMW)
	g.Post("/request", h.request) // generate + save + log code
	g.Post("/verify", h.verify)   // confirm code → flip email_verified=TRUE
}

func (h *Handler) request(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return httpx.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
	}
	st, err := h.requestUC.Execute(c.Context(), userID)
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, RequestCodeResponse{
		Message:   "verification code issued (sent to your email in production)",
		Code:      st.Code,
		ExpiresAt: *st.ExpiresAt,
	})
}

func (h *Handler) verify(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return httpx.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
	}
	var b VerifyRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := h.verifyUC.Execute(c.Context(), application.VerifyCodeInput{UserID: userID, Code: b.Code}); err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, fiber.Map{"message": "email verified", "verified": true})
}

func mapErr(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "USER_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrAlreadyVerified):
		return httpx.Error(c, fiber.StatusConflict, "ALREADY_VERIFIED", err.Error())
	case errors.Is(err, domain.ErrInvalidCode):
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_CODE", err.Error())
	case errors.Is(err, domain.ErrExpiredCode):
		return httpx.Error(c, fiber.StatusBadRequest, "EXPIRED_CODE", err.Error())
	case errors.Is(err, domain.ErrNoCodeRequested):
		return httpx.Error(c, fiber.StatusBadRequest, "NO_CODE_REQUESTED", err.Error())
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
