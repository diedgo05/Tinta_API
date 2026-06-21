// Package jwt adapts the shared jwtauth.Signer to the auth.TokenSigner port.
package jwt

import (
	"time"

	"github.com/google/uuid"
	"github.com/tinta/shared/jwtauth"
)

// SignerAdapter implements ports.TokenSigner by delegating to jwtauth.Signer.
type SignerAdapter struct {
	signer *jwtauth.Signer
}

// NewSignerAdapter wraps the shared signer.
func NewSignerAdapter(signer *jwtauth.Signer) *SignerAdapter {
	return &SignerAdapter{signer: signer}
}

// SignAccess implements ports.TokenSigner.
func (s *SignerAdapter) SignAccess(userID uuid.UUID, email, role string) (string, time.Time, error) {
	return s.signer.SignAccessToken(userID, email, role)
}

// SignRefresh implements ports.TokenSigner.
func (s *SignerAdapter) SignRefresh(userID uuid.UUID, email, role string) (string, time.Time, error) {
	return s.signer.SignRefreshToken(userID, email, role)
}
