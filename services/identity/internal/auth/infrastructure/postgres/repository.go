// Package postgres implements ports.RefreshTokenRepository using pgx.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tinta/identity/internal/auth/domain"
)

// RefreshTokenRepository persists refresh tokens.
type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

// NewRefreshTokenRepository wires the pgx pool.
func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

type tokenRow struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string
	UserAgent  *string
	IPAddress  *string
	RevokedAt  *time.Time
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

func (r tokenRow) toDomain() *domain.RefreshToken {
	ua := ""
	if r.UserAgent != nil {
		ua = *r.UserAgent
	}
	ip := ""
	if r.IPAddress != nil {
		ip = *r.IPAddress
	}
	return &domain.RefreshToken{
		ID:        r.ID,
		UserID:    r.UserID,
		TokenHash: r.TokenHash,
		UserAgent: ua,
		IPAddress: ip,
		RevokedAt: r.RevokedAt,
		ExpiresAt: r.ExpiresAt,
		CreatedAt: r.CreatedAt,
	}
}

// Create inserts a new refresh token.
func (r *RefreshTokenRepository) Create(ctx context.Context, t *domain.RefreshToken) (*domain.RefreshToken, error) {
	const q = `
		INSERT INTO refresh_tokens (user_id, token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, token_hash, user_agent, ip_address, revoked_at, expires_at, created_at`

	var row tokenRow
	err := r.db.QueryRow(ctx, q,
		t.UserID, t.TokenHash, nullable(t.UserAgent), nullable(t.IPAddress), t.ExpiresAt,
	).Scan(
		&row.ID, &row.UserID, &row.TokenHash, &row.UserAgent, &row.IPAddress,
		&row.RevokedAt, &row.ExpiresAt, &row.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert refresh token: %w", err)
	}
	return row.toDomain(), nil
}

// GetByHash returns the (non-revoked, non-expired) token matching the hash.
func (r *RefreshTokenRepository) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	const q = `
		SELECT id, user_id, token_hash, user_agent, ip_address, revoked_at, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()`

	var row tokenRow
	err := r.db.QueryRow(ctx, q, hash).Scan(
		&row.ID, &row.UserID, &row.TokenHash, &row.UserAgent, &row.IPAddress,
		&row.RevokedAt, &row.ExpiresAt, &row.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidToken
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	return row.toDomain(), nil
}

// RevokeByHash marks a token as revoked. Idempotent.
func (r *RefreshTokenRepository) RevokeByHash(ctx context.Context, hash string) error {
	const q = `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL`
	if _, err := r.db.Exec(ctx, q, hash); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

// RevokeAllForUser revokes every active token of a user.
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	const q = `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`
	if _, err := r.db.Exec(ctx, q, userID); err != nil {
		return fmt.Errorf("revoke all refresh tokens: %w", err)
	}
	return nil
}

// nullable converts an empty string to a NULL when inserting via pgx.
func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
