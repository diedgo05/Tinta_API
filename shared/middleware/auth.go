// Package middleware provides reusable Fiber middlewares for Tinta services.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tinta/shared/jwtauth"
)

// Context keys used by RequireAuth to expose the authenticated user to handlers.
const (
	CtxUserID = "userID"
	CtxEmail  = "email"
	CtxRole   = "role"
)

// RequireAuth returns a middleware that validates a JWT in the Authorization header
// and stores the user claims in fiber.Ctx locals.
//
// Usage:
//
//	app.Use(middleware.RequireAuth(verifier))
//	// then in handler:
//	userID := c.Locals(middleware.CtxUserID).(uuid.UUID)
func RequireAuth(verifier *jwtauth.Verifier) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid authorization header format")
		}

		claims, err := verifier.Verify(parts[1])
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		// Only access tokens are valid for protected routes.
		if claims.Type != "access" {
			return fiber.NewError(fiber.StatusUnauthorized, "token type not allowed")
		}

		c.Locals(CtxUserID, claims.UserID)
		c.Locals(CtxEmail, claims.Email)
		c.Locals(CtxRole, claims.Role)
		return c.Next()
	}
}

// RequireRole returns a middleware that enforces a specific role.
// Must be used AFTER RequireAuth.
func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals(CtxRole).(string)
		if !ok || role == "" {
			return fiber.NewError(fiber.StatusForbidden, "missing role")
		}
		for _, allowed := range allowedRoles {
			if role == allowed {
				return c.Next()
			}
		}
		return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
	}
}

// UserIDFromContext is a helper to safely extract the user ID from fiber.Ctx.
func UserIDFromContext(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(CtxUserID).(uuid.UUID)
	return id, ok
}
