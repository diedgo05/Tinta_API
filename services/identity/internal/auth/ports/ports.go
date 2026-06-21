// Package ports declares the auth interfaces consumed by the application layer.
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tinta/identity/internal/auth/domain"
)

// RefreshTokenRepository persists refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, t *domain.RefreshToken) (*domain.RefreshToken, error)
	GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	RevokeByHash(ctx context.Context, hash string) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

// TokenSigner issues access and refresh JWTs.
type TokenSigner interface {
	SignAccess(userID uuid.UUID, email, role string) (token string, expiresAt time.Time, err error)
	SignRefresh(userID uuid.UUID, email, role string) (token string, expiresAt time.Time, err error)
}
