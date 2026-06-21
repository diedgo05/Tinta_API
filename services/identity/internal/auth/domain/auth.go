// Package domain contains the auth business rules.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// RefreshToken is the persistent token used to renew sessions.
type RefreshToken struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string
	UserAgent  string
	IPAddress  string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
}

// IsValid reports whether the token can still be used.
func (rt *RefreshToken) IsValid() bool {
	return rt.RevokedAt == nil && rt.ExpiresAt.After(time.Now())
}

// Sentinel errors.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid refresh token")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenRevoked       = errors.New("token revoked")
)
