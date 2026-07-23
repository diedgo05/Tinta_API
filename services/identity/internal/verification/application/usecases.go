package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/tinta/identity/internal/verification/domain"
	"github.com/tinta/identity/internal/verification/ports"
	"github.com/tinta/shared/mailer"
)

type RequestCodeUseCase struct {
	repo   ports.VerificationRepository
	mailer mailer.Mailer
	log    zerolog.Logger
}

func NewRequestCodeUseCase(r ports.VerificationRepository, m mailer.Mailer, log zerolog.Logger) *RequestCodeUseCase {
	return &RequestCodeUseCase{repo: r, mailer: m, log: log}
}

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

	subject, body := mailer.VerificationCodeEmail(code)
	if err := uc.mailer.Send(st.Email, subject, body); err != nil {
		uc.log.Error().Err(err).Str("email", st.Email).Msg("failed to send verification email")
	} else {
		uc.log.Info().Str("email", st.Email).Msg("verification email sent")
	}

	st.Code = code
	st.ExpiresAt = &expires
	return st, nil
}

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
	if strings.TrimSpace(in.Code) != stored {
		return domain.ErrInvalidCode
	}
	return uc.repo.MarkVerified(ctx, in.UserID)
}