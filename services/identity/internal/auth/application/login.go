// Package application contains the auth use cases (login, refresh, logout).
package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	authdomain "github.com/tinta/identity/internal/auth/domain"
	authports "github.com/tinta/identity/internal/auth/ports"
	userdomain "github.com/tinta/identity/internal/user/domain"
	userports "github.com/tinta/identity/internal/user/ports"
)

// LoginInput is the input DTO for the Login use case.
type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

// LoginOutput contains the issued tokens.
type LoginOutput struct {
	AccessToken  string
	RefreshToken string
}

// LoginUseCase authenticates a user with email + password and issues tokens.
type LoginUseCase struct {
	userRepo  userports.UserRepository
	hasher    userports.PasswordHasher
	tokenRepo authports.RefreshTokenRepository
	signer    authports.TokenSigner
}

// NewLoginUseCase wires the dependencies.
func NewLoginUseCase(
	userRepo userports.UserRepository,
	hasher userports.PasswordHasher,
	tokenRepo authports.RefreshTokenRepository,
	signer authports.TokenSigner,
) *LoginUseCase {
	return &LoginUseCase{userRepo: userRepo, hasher: hasher, tokenRepo: tokenRepo, signer: signer}
}

// Execute performs the login.
func (uc *LoginUseCase) Execute(ctx context.Context, in LoginInput) (*LoginOutput, error) {
	email := userdomain.NormalizeEmail(in.Email)

	// 1. Lookup user
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return nil, authdomain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	// 2. Verify password
	if err := uc.hasher.Verify(in.Password, user.PasswordHash); err != nil {
		return nil, authdomain.ErrInvalidCredentials
	}

	// 3. Issue tokens
	access, _, err := uc.signer.SignAccess(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}
	refresh, refreshExp, err := uc.signer.SignRefresh(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	// 4. Persist refresh token (hashed, never the raw token)
	hash := hashToken(refresh)
	if _, err := uc.tokenRepo.Create(ctx, &authdomain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash,
		UserAgent: in.UserAgent,
		IPAddress: in.IPAddress,
		ExpiresAt: refreshExp,
	}); err != nil {
		return nil, fmt.Errorf("persist refresh token: %w", err)
	}

	return &LoginOutput{AccessToken: access, RefreshToken: refresh}, nil
}

// hashToken returns a deterministic hash of a refresh token.
// We store the hash so a database leak does not expose usable tokens.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
