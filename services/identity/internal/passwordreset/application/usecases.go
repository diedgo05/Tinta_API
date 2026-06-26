// Package application contains the password-reset use cases.
package application

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/tinta/identity/internal/passwordreset/domain"
	"github.com/tinta/identity/internal/passwordreset/ports"
)

// ---------- Request reset code (PUBLIC - by email) ----------

type RequestResetInput struct {
	Email string
}

type RequestResetUseCase struct {
	repo ports.PasswordResetRepository
	log  zerolog.Logger
}

func NewRequestResetUseCase(r ports.PasswordResetRepository, log zerolog.Logger) *RequestResetUseCase {
	return &RequestResetUseCase{repo: r, log: log}
}

// Execute is intentionally LENIENT about missing users: we don't reveal
// whether an email exists in the system (anti-enumeration).
// Returns the code+expires only when the user actually exists; for unknown
// emails we return a synthetic success with empty code so the HTTP layer
// always responds 200 with a generic message.
func (uc *RequestResetUseCase) Execute(ctx context.Context, in RequestResetInput) (*domain.ResetStatus, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	userID, err := uc.repo.GetUserIDByEmail(ctx, email)
	if err != nil {
		// Don't leak; return empty status (HTTP responds generic message).
		uc.log.Info().Str("email", email).Msg("password reset requested for unknown email (silently ignored)")
		return &domain.ResetStatus{Email: email}, nil
	}

	code, err := domain.GenerateCode()
	if err != nil {
		return nil, err
	}
	exp := time.Now().Add(domain.ResetCodeTTL)
	if err := uc.repo.SaveResetCode(ctx, userID, code, exp); err != nil {
		return nil, err
	}

	uc.log.Info().
		Str("email", email).
		Str("code", code).
		Time("expires_at", exp).
		Msg("password reset code issued (would be emailed in production)")

	return &domain.ResetStatus{
		UserID: userID, Email: email, Code: code, ExpiresAt: &exp,
	}, nil
}

// ---------- Confirm reset (PUBLIC - by email + code + new password) ----------

type ConfirmResetInput struct {
	Email       string
	Code        string
	NewPassword string
}

type ConfirmResetUseCase struct {
	repo   ports.PasswordResetRepository
	hasher ports.PasswordHasher
}

func NewConfirmResetUseCase(r ports.PasswordResetRepository, h ports.PasswordHasher) *ConfirmResetUseCase {
	return &ConfirmResetUseCase{repo: r, hasher: h}
}

func (uc *ConfirmResetUseCase) Execute(ctx context.Context, in ConfirmResetInput) error {
	if err := domain.ValidatePassword(in.NewPassword); err != nil {
		return err
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))
	code := strings.TrimSpace(in.Code)

	userID, err := uc.repo.GetUserIDFromActiveCode(ctx, email, code)
	if err != nil {
		return err
	}

	hash, err := uc.hasher.Hash(in.NewPassword)
	if err != nil {
		return err
	}
	return uc.repo.UpdatePasswordAndClearCode(ctx, userID, hash)
}
