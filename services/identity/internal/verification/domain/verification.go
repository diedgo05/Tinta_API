// Package domain contains the EmailVerification entity and logic.
package domain

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CodeTTL is how long a verification code stays valid.
const CodeTTL = 15 * time.Minute

// VerificationStatus is a snapshot returned by use cases.
type VerificationStatus struct {
	UserID    uuid.UUID
	Email     string
	Verified  bool
	Code      string     // only filled when a new code was just generated (for response/log)
	ExpiresAt *time.Time // when Code is non-empty
}

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrAlreadyVerified = errors.New("email is already verified")
	ErrInvalidCode     = errors.New("invalid verification code")
	ErrExpiredCode     = errors.New("verification code expired")
	ErrNoCodeRequested = errors.New("no verification code was requested")
)

// GenerateCode returns a cryptographically random 6-digit code.
func GenerateCode() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Map 3 bytes (24 bits) to a number 0..999999, then format with leading zeros.
	n := (uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])) % 1000000
	return fmt.Sprintf("%06d", n), nil
}
