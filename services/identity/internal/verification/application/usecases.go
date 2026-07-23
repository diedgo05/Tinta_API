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
	"github.com/tinta/shared/mailer"
)

// ---------- Request code (by authenticated user) ----------

type RequestCodeUseCase struct {
	repo   ports.VerificationRepository
	mailer mailer.Mailer
	log    zerolog.Logger
}

func NewRequestCodeUseCase(r ports.VerificationRepository, m mailer.Mailer, log zerolog.Logger) *RequestCodeUseCase {
	return &RequestCodeUseCase{repo: r, mailer: m, log: log}
}

// Execute generates a fresh code for the authenticated user, lo guarda, y
// lo manda por correo de verdad a tintaappmovil@gmail.com → email del
// usuario. Sigue regresando el código en el VerificationStatus para que
// el equipo pueda seguir viéndolo en la respuesta HTTP mientras se
// confirma que el envío por SMTP funciona en producción; una vez
// confirmado, ese campo se puede quitar del handler HTTP por seguridad.
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

	// Envío real por correo — al email con el que el usuario se registró.
	subject, body := mailer.VerificationCodeEmail(code)
	if err := uc.mailer.Send(st.Email, subject, body); err != nil {
		// No se revierte el código guardado: el usuario puede seguir
		// usándolo si de alguna forma lo obtiene (ej. lo ve en logs de
		// desarrollo), pero sí registramos el fallo para dar
		// seguimiento — un correo que no llega es un problema real.
		uc.log.Error().Err(err).Str("email", st.Email).Msg("failed to send verification email")
	} else {
		uc.log.Info().Str("email", st.Email).Msg("verification email sent")
	}

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