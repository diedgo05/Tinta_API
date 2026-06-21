// Package jwtauth provides JWT token issuance and verification used across services.
// Only the Identity service signs tokens (with the private key);
// Community and Recommendations only verify them (with the public key).
package jwtauth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims is the JWT payload used by all Tinta services.
type Claims struct {
	UserID uuid.UUID `json:"sub"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	Type   string    `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// Verifier verifies JWT tokens using a public key.
// Used by every service that needs authentication.
type Verifier struct {
	publicKey *rsa.PublicKey
}

// NewVerifier loads an RSA public key from a PEM file.
func NewVerifier(publicKeyPath string) (*Verifier, error) {
	keyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return &Verifier{publicKey: pubKey}, nil
}

// Verify parses and validates a JWT, returning its claims if valid.
func (v *Verifier) Verify(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// Signer issues JWT tokens. Only the Identity service uses this.
type Signer struct {
	privateKey       *rsa.PrivateKey
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
}

// NewSigner loads an RSA private key from a PEM file.
func NewSigner(privateKeyPath string, accessTTL, refreshTTL time.Duration) (*Signer, error) {
	keyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	return &Signer{
		privateKey:      privKey,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}, nil
}

// SignAccessToken issues a short-lived access token.
func (s *Signer) SignAccessToken(userID uuid.UUID, email, role string) (string, time.Time, error) {
	return s.sign(userID, email, role, "access", s.accessTokenTTL)
}

// SignRefreshToken issues a long-lived refresh token.
func (s *Signer) SignRefreshToken(userID uuid.UUID, email, role string) (string, time.Time, error) {
	return s.sign(userID, email, role, "refresh", s.refreshTokenTTL)
}

func (s *Signer) sign(userID uuid.UUID, email, role, tokenType string, ttl time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(ttl)
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "tinta-identity",
			Subject:   userID.String(),
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}
