// Package postgres implements ports.VerificationRepository.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tinta/identity/internal/verification/domain"
)

type VerificationRepository struct{ db *pgxpool.Pool }

func NewVerificationRepository(db *pgxpool.Pool) *VerificationRepository {
	return &VerificationRepository{db: db}
}

func (r *VerificationRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.VerificationStatus, error) {
	const q = `SELECT id, email, email_verified FROM users WHERE id=$1 AND deleted_at IS NULL`
	var st domain.VerificationStatus
	err := r.db.QueryRow(ctx, q, id).Scan(&st.UserID, &st.Email, &st.Verified)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &st, nil
}

func (r *VerificationRepository) GetUserByEmail(ctx context.Context, email string) (*domain.VerificationStatus, error) {
	const q = `SELECT id, email, email_verified FROM users WHERE email=$1 AND deleted_at IS NULL`
	var st domain.VerificationStatus
	err := r.db.QueryRow(ctx, q, email).Scan(&st.UserID, &st.Email, &st.Verified)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &st, nil
}

func (r *VerificationRepository) SaveCode(ctx context.Context, userID uuid.UUID, code string, expiresAt time.Time) error {
	const q = `UPDATE users SET
		verification_code=$2,
		verification_expires_at=$3,
		updated_at=NOW()
	WHERE id=$1 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, q, userID, code, expiresAt)
	if err != nil {
		return fmt.Errorf("save code: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *VerificationRepository) MarkVerified(ctx context.Context, userID uuid.UUID) error {
	const q = `UPDATE users SET
		email_verified=TRUE,
		verification_code=NULL,
		verification_expires_at=NULL,
		updated_at=NOW()
	WHERE id=$1 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("mark verified: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *VerificationRepository) GetStoredCode(ctx context.Context, userID uuid.UUID) (string, *time.Time, bool, error) {
	const q = `SELECT email_verified, verification_code, verification_expires_at FROM users WHERE id=$1 AND deleted_at IS NULL`
	var verified bool
	var code *string
	var expires *time.Time
	err := r.db.QueryRow(ctx, q, userID).Scan(&verified, &code, &expires)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil, false, domain.ErrUserNotFound
		}
		return "", nil, false, err
	}
	c := ""
	if code != nil {
		c = *code
	}
	return c, expires, verified, nil
}
