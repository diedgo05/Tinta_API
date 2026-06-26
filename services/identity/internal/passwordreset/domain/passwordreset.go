// Package domain contains the password-reset entity and logic.
package domain

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ResetCodeTTL is how long a password-reset code stays valid.
const ResetCodeTTL = 15 * time.Minute

// MinPasswordLen mirrors the rule from the user module (kept locally to avoid coupling).
const MinPasswordLen = 8

type ResetStatus struct {
	UserID    uuid.UUID
	Email     string
	Code      string
	ExpiresAt *time.Time
}

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidCode     = errors.New("invalid reset code")
	ErrExpiredCode     = errors.New("reset code expired")
	ErrNoCodeRequested = errors.New("no reset code was requested")
	ErrWeakPassword    = errors.New("password is too weak (min 8 chars, must include letter + digit)")
)

// GenerateCode returns a 6-digit random code.
func GenerateCode() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	n := (uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])) % 1000000
	return fmt.Sprintf("%06d", n), nil
}

// ValidatePassword enforces the same rule as the user-register flow.
func ValidatePassword(p string) error {
	if len(p) < MinPasswordLen {
		return ErrWeakPassword
	}
	hasLetter, hasDigit := false, false
	for _, r := range p {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
			hasLetter = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return ErrWeakPassword
	}
	return nil
}
