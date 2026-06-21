package application

import (
	"context"

	authports "github.com/tinta/identity/internal/auth/ports"
)

// LogoutUseCase revokes a refresh token, ending the session.
type LogoutUseCase struct {
	tokenRepo authports.RefreshTokenRepository
}

// NewLogoutUseCase wires the dependency.
func NewLogoutUseCase(tokenRepo authports.RefreshTokenRepository) *LogoutUseCase {
	return &LogoutUseCase{tokenRepo: tokenRepo}
}

// Execute revokes the given refresh token. Idempotent: revoking twice is fine.
func (uc *LogoutUseCase) Execute(ctx context.Context, refreshToken string) error {
	hash := hashToken(refreshToken)
	return uc.tokenRepo.RevokeByHash(ctx, hash)
}
