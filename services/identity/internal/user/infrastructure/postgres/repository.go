// Package postgres implements ports.UserRepository using pgx.
//
// NOTE: We use pgx queries directly here rather than the sqlc-generated code,
// to keep this example readable. In production, the recommended path is to
// generate sqlc code from sqlc/queries.sql and call those functions instead.
// Both approaches use pgx underneath; switching is mechanical.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tinta/identity/internal/user/domain"
	"github.com/tinta/identity/internal/user/ports"
)

// UserRepository persists users in PostgreSQL.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository wires the pgx pool.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// userRow mirrors the columns of the users table.
type userRow struct {
	ID            uuid.UUID
	Email         string
	PasswordHash  string
	Name          string
	Role          string
	EmailVerified bool
	AvatarURL     *string
	Language      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (r userRow) toDomain() *domain.User {
	avatar := ""
	if r.AvatarURL != nil {
		avatar = *r.AvatarURL
	}
	return &domain.User{
		ID:            r.ID,
		Email:         r.Email,
		PasswordHash:  r.PasswordHash,
		Name:          r.Name,
		Role:          domain.Role(r.Role),
		EmailVerified: r.EmailVerified,
		AvatarURL:     avatar,
		Language:      r.Language,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

// Create inserts a new user.
func (r *UserRepository) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	const q = `
		INSERT INTO users (email, password_hash, name, role, email_verified, language)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, password_hash, name, role, email_verified, avatar_url, language, created_at, updated_at`

	var row userRow
	err := r.db.QueryRow(ctx, q,
		u.Email, u.PasswordHash, u.Name, string(u.Role), u.EmailVerified, u.Language,
	).Scan(
		&row.ID, &row.Email, &row.PasswordHash, &row.Name, &row.Role,
		&row.EmailVerified, &row.AvatarURL, &row.Language, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "users_email_key") {
			return nil, domain.ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return row.toDomain(), nil
}

// GetByID returns a user by id or ErrUserNotFound.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const q = `
		SELECT id, email, password_hash, name, role, email_verified, avatar_url, language, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`

	var row userRow
	err := r.db.QueryRow(ctx, q, id).Scan(
		&row.ID, &row.Email, &row.PasswordHash, &row.Name, &row.Role,
		&row.EmailVerified, &row.AvatarURL, &row.Language, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return row.toDomain(), nil
}

// GetByEmail returns a user by email or ErrUserNotFound.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, email, password_hash, name, role, email_verified, avatar_url, language, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`

	var row userRow
	err := r.db.QueryRow(ctx, q, email).Scan(
		&row.ID, &row.Email, &row.PasswordHash, &row.Name, &row.Role,
		&row.EmailVerified, &row.AvatarURL, &row.Language, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return row.toDomain(), nil
}

// EmailExists is a cheap existence check.
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`
	var exists bool
	if err := r.db.QueryRow(ctx, q, email).Scan(&exists); err != nil {
		return false, fmt.Errorf("email exists check: %w", err)
	}
	return exists, nil
}

// Update applies a partial update.
func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, updates ports.UserUpdates) (*domain.User, error) {
	const q = `
		UPDATE users SET
			name       = COALESCE($2, name),
			avatar_url = COALESCE($3, avatar_url),
			language   = COALESCE($4, language),
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, password_hash, name, role, email_verified, avatar_url, language, created_at, updated_at`

	var row userRow
	err := r.db.QueryRow(ctx, q, id, updates.Name, updates.AvatarURL, updates.Language).Scan(
		&row.ID, &row.Email, &row.PasswordHash, &row.Name, &row.Role,
		&row.EmailVerified, &row.AvatarURL, &row.Language, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}
	return row.toDomain(), nil
}

// SoftDelete marks the user as deleted.
func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
