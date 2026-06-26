// Package ports declares the password-reset repository and password hasher.
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PasswordResetRepository handles persistence for the reset flow.
type PasswordResetRepository interface {
	GetUserIDByEmail(ctx context.Context, email string) (uuid.UUID, error)
	SaveResetCode(ctx context.Context, userID uuid.UUID, code string, expiresAt time.Time) error
	GetStoredResetCode(ctx context.Context, userID uuid.UUID) (code string, expiresAt *time.Time, err error)
	UpdatePasswordAndClearCode(ctx context.Context, userID uuid.UUID, passwordHash string) error
	GetUserIDFromActiveCode(ctx context.Context, email, code string) (uuid.UUID, error)
}

// PasswordHasher abstracts Argon2 (or any other) hashing.
// Identity already has one in user/ports; we duplicate the interface here
// to avoid cross-module coupling. The same concrete hasher is wired in main.go.
type PasswordHasher interface {
	Hash(password string) (string, error)
}
