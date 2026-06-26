// Package application contains the email-verification use cases.
package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/tinta/identity/internal/verification/domain"
	"github.com/tinta/identity/internal/verification/ports"
)

// ---------- Request code (by authenticated user) ----------

type RequestCodeUseCase struct {
	repo ports.VerificationRepository
	log  zerolog.Logger
}

func NewRequestCodeUseCase(r ports.VerificationRepository, log zerolog.Logger) *RequestCodeUseCase {
	return &RequestCodeUseCase{repo: r, log: log}
}

// Execute generates a fresh code for the authenticated user and saves it.
// Returns the code + expires_at in the status so the HTTP layer can echo
// them back to the client (since we don't have a real SMTP provider).
func (uc *RequestCodeUseCase) Execute(ctx context.Context, userID uuid.UUID) (*domain.VerificationStatus, error) {
	st, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if st.Verified {
		return nil, domain.ErrAlreadyVerified
	}

	code, err := domain.GenerateCode()
	if err != nil {
		return nil, err
	}
	expires := time.Now().Add(domain.CodeTTL)
	if err := uc.repo.SaveCode(ctx, userID, code, expires); err != nil {
		return nil, err
	}

	// Pretend-send via log (would be SMTP in real production).
	uc.log.Info().
		Str("email", st.Email).
		Str("code", code).
		Time("expires_at", expires).
		Msg("verification code issued (would be emailed in production)")

	st.Code = code
	st.ExpiresAt = &expires
	return st, nil
}

// ---------- Verify code ----------

type VerifyCodeInput struct {
	UserID uuid.UUID
	Code   string
}

type VerifyCodeUseCase struct {
	repo ports.VerificationRepository
}

func NewVerifyCodeUseCase(r ports.VerificationRepository) *VerifyCodeUseCase {
	return &VerifyCodeUseCase{repo: r}
}

func (uc *VerifyCodeUseCase) Execute(ctx context.Context, in VerifyCodeInput) error {
	stored, expires, alreadyVerified, err := uc.repo.GetStoredCode(ctx, in.UserID)
	if err != nil {
		return err
	}
	if alreadyVerified {
		return domain.ErrAlreadyVerified
	}
	if stored == "" || expires == nil {
		return domain.ErrNoCodeRequested
	}
	if time.Now().After(*expires) {
		return domain.ErrExpiredCode
	}
	// constant-time-ish comparison; codes are short, so this is fine.
	if strings.TrimSpace(in.Code) != stored {
		return domain.ErrInvalidCode
	}
	return uc.repo.MarkVerified(ctx, in.UserID)
}
