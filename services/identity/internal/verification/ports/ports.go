// Package ports declares interfaces for the verification module.
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tinta/identity/internal/verification/domain"
)

// VerificationRepository handles reads/writes for email verification.
// It does NOT touch the `users` row directly outside of these methods.
type VerificationRepository interface {
	// GetUserByID returns minimal user data needed to verify.
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.VerificationStatus, error)

	// GetUserByEmail used by the resend-by-email flow.
	GetUserByEmail(ctx context.Context, email string) (*domain.VerificationStatus, error)

	// SaveCode persists a freshly generated code and its expiration on the user row.
	SaveCode(ctx context.Context, userID uuid.UUID, code string, expiresAt time.Time) error

	// MarkVerified flips email_verified=TRUE and clears the code columns.
	MarkVerified(ctx context.Context, userID uuid.UUID) error

	// GetStoredCode returns the active code + expiration for a user (for the verify step).
	GetStoredCode(ctx context.Context, userID uuid.UUID) (code string, expiresAt *time.Time, verified bool, err error)
}
