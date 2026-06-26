// Package postgres implements ports.PasswordResetRepository.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tinta/identity/internal/passwordreset/domain"
)

type PasswordResetRepository struct{ db *pgxpool.Pool }

func NewPasswordResetRepository(db *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) GetUserIDByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT id FROM users WHERE email=$1 AND deleted_at IS NULL`, email).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, domain.ErrUserNotFound
		}
		return uuid.Nil, err
	}
	return id, nil
}

func (r *PasswordResetRepository) SaveResetCode(ctx context.Context, userID uuid.UUID, code string, expiresAt time.Time) error {
	const q = `UPDATE users SET
		password_reset_code=$2,
		password_reset_expires_at=$3,
		updated_at=NOW()
	WHERE id=$1 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, q, userID, code, expiresAt)
	if err != nil {
		return fmt.Errorf("save reset code: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *PasswordResetRepository) GetStoredResetCode(ctx context.Context, userID uuid.UUID) (string, *time.Time, error) {
	const q = `SELECT password_reset_code, password_reset_expires_at FROM users WHERE id=$1 AND deleted_at IS NULL`
	var code *string
	var exp *time.Time
	err := r.db.QueryRow(ctx, q, userID).Scan(&code, &exp)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil, domain.ErrUserNotFound
		}
		return "", nil, err
	}
	c := ""
	if code != nil {
		c = *code
	}
	return c, exp, nil
}

// GetUserIDFromActiveCode atomically validates code, email and expiration.
// Returns the user id only if everything matches.
func (r *PasswordResetRepository) GetUserIDFromActiveCode(ctx context.Context, email, code string) (uuid.UUID, error) {
	const q = `SELECT id, password_reset_code, password_reset_expires_at
		FROM users WHERE email=$1 AND deleted_at IS NULL`
	var id uuid.UUID
	var stored *string
	var exp *time.Time
	err := r.db.QueryRow(ctx, q, email).Scan(&id, &stored, &exp)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, domain.ErrUserNotFound
		}
		return uuid.Nil, err
	}
	if stored == nil || exp == nil {
		return uuid.Nil, domain.ErrNoCodeRequested
	}
	if time.Now().After(*exp) {
		return uuid.Nil, domain.ErrExpiredCode
	}
	if *stored != code {
		return uuid.Nil, domain.ErrInvalidCode
	}
	return id, nil
}

func (r *PasswordResetRepository) UpdatePasswordAndClearCode(ctx context.Context, userID uuid.UUID, hash string) error {
	const q = `UPDATE users SET
		password_hash=$2,
		password_reset_code=NULL,
		password_reset_expires_at=NULL,
		updated_at=NOW()
	WHERE id=$1 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, q, userID, hash)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
