package application

import (
	"context"
	"errors"
	"fmt"

	authdomain "github.com/tinta/identity/internal/auth/domain"
	authports "github.com/tinta/identity/internal/auth/ports"
	userdomain "github.com/tinta/identity/internal/user/domain"
	userports "github.com/tinta/identity/internal/user/ports"
)

// RefreshOutput contains the newly issued tokens.
type RefreshOutput struct {
	AccessToken  string
	RefreshToken string
}

// RefreshUseCase rotates a refresh token, invalidating the old one.
type RefreshUseCase struct {
	userRepo  userports.UserRepository
	tokenRepo authports.RefreshTokenRepository
	signer    authports.TokenSigner
}

// NewRefreshUseCase wires the dependencies.
func NewRefreshUseCase(
	userRepo userports.UserRepository,
	tokenRepo authports.RefreshTokenRepository,
	signer authports.TokenSigner,
) *RefreshUseCase {
	return &RefreshUseCase{userRepo: userRepo, tokenRepo: tokenRepo, signer: signer}
}

// Execute validates the given refresh token, revokes it, and issues a new pair.
// userAgent and ipAddress are stored for audit on the new token.
func (uc *RefreshUseCase) Execute(ctx context.Context, refreshToken, userAgent, ipAddress string) (*RefreshOutput, error) {
	hash := hashToken(refreshToken)

	// 1. Lookup token by hash; expects non-revoked + not expired
	stored, err := uc.tokenRepo.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, authdomain.ErrInvalidToken) {
			return nil, authdomain.ErrInvalidToken
		}
		return nil, fmt.Errorf("lookup refresh token: %w", err)
	}
	if !stored.IsValid() {
		return nil, authdomain.ErrInvalidToken
	}

	// 2. Lookup the owner
	user, err := uc.userRepo.GetByID(ctx, stored.UserID)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return nil, authdomain.ErrInvalidToken
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	// 3. Revoke the old token (rotation)
	if err := uc.tokenRepo.RevokeByHash(ctx, hash); err != nil {
		return nil, fmt.Errorf("revoke old token: %w", err)
	}

	// 4. Issue a new pair
	access, _, err := uc.signer.SignAccess(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}
	newRefresh, newRefreshExp, err := uc.signer.SignRefresh(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	if _, err := uc.tokenRepo.Create(ctx, &authdomain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken(newRefresh),
		UserAgent: userAgent,
		IPAddress: ipAddress,
		ExpiresAt: newRefreshExp,
	}); err != nil {
		return nil, fmt.Errorf("persist new refresh token: %w", err)
	}

	return &RefreshOutput{AccessToken: access, RefreshToken: newRefresh}, nil
}
