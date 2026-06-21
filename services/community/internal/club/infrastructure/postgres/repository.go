// Package postgres implements ports.ClubRepository using pgx.
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
	"github.com/tinta/community/internal/club/domain"
	"github.com/tinta/community/internal/club/ports"
)

// ClubRepository persists clubs in PostgreSQL.
type ClubRepository struct {
	db *pgxpool.Pool
}

// NewClubRepository wires the pgx pool.
func NewClubRepository(db *pgxpool.Pool) *ClubRepository {
	return &ClubRepository{db: db}
}

type clubRow struct {
	ID          uuid.UUID
	CreatorID   uuid.UUID
	BookID      *uuid.UUID
	Name        string
	Description *string
	IsPrivate   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r clubRow) toDomain() *domain.Club {
	desc := ""
	if r.Description != nil {
		desc = *r.Description
	}
	return &domain.Club{
		ID:          r.ID,
		CreatorID:   r.CreatorID,
		BookID:      r.BookID,
		Name:        r.Name,
		Description: desc,
		IsPrivate:   r.IsPrivate,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

const selectClubColumns = `
	id, creator_id, book_id, name, description, is_private, created_at, updated_at`

// Create inserts a new club.
func (r *ClubRepository) Create(ctx context.Context, c *domain.Club) (*domain.Club, error) {
	const q = `
		INSERT INTO clubs (creator_id, book_id, name, description, is_private)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + selectClubColumns

	var row clubRow
	err := r.db.QueryRow(ctx, q,
		c.CreatorID, c.BookID, c.Name, nullableString(c.Description), c.IsPrivate,
	).Scan(
		&row.ID, &row.CreatorID, &row.BookID, &row.Name, &row.Description,
		&row.IsPrivate, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert club: %w", err)
	}
	return row.toDomain(), nil
}

// GetByID returns a club by id or ErrClubNotFound.
func (r *ClubRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Club, error) {
	const q = `SELECT ` + selectClubColumns + `
		FROM clubs WHERE id = $1 AND deleted_at IS NULL`

	var row clubRow
	err := r.db.QueryRow(ctx, q, id).Scan(
		&row.ID, &row.CreatorID, &row.BookID, &row.Name, &row.Description,
		&row.IsPrivate, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrClubNotFound
		}
		return nil, fmt.Errorf("get club: %w", err)
	}
	return row.toDomain(), nil
}

// List returns a paginated list of clubs with optional filters.
func (r *ClubRepository) List(ctx context.Context, f ports.ListFilter) (*ports.ListResult, error) {
	// Build the WHERE clause dynamically.
	conds := []string{"deleted_at IS NULL"}
	args := []any{}
	idx := 1

	if !f.IncludeAll {
		conds = append(conds, "is_private = FALSE")
	}
	if f.CreatorID != nil {
		conds = append(conds, fmt.Sprintf("creator_id = $%d", idx))
		args = append(args, *f.CreatorID)
		idx++
	}
	if f.BookID != nil {
		conds = append(conds, fmt.Sprintf("book_id = $%d", idx))
		args = append(args, *f.BookID)
		idx++
	}

	where := "WHERE " + strings.Join(conds, " AND ")

	// 1. Total count.
	var total int64
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM clubs "+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count clubs: %w", err)
	}

	// 2. Page query.
	offset := (f.Page - 1) * f.PageSize
	listQ := fmt.Sprintf(
		"SELECT %s FROM clubs %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		selectClubColumns, where, idx, idx+1,
	)
	args = append(args, f.PageSize, offset)

	rows, err := r.db.Query(ctx, listQ, args...)
	if err != nil {
		return nil, fmt.Errorf("list clubs: %w", err)
	}
	defer rows.Close()

	items := make([]*domain.Club, 0)
	for rows.Next() {
		var row clubRow
		if err := rows.Scan(
			&row.ID, &row.CreatorID, &row.BookID, &row.Name, &row.Description,
			&row.IsPrivate, &row.CreatedAt, &row.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan club: %w", err)
		}
		items = append(items, row.toDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return &ports.ListResult{
		Items:    items,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
}

// Update applies a partial update.
func (r *ClubRepository) Update(ctx context.Context, id uuid.UUID, u ports.ClubUpdates) (*domain.Club, error) {
	const q = `
		UPDATE clubs SET
			name        = COALESCE($2, name),
			description = COALESCE($3, description),
			book_id     = COALESCE($4, book_id),
			is_private  = COALESCE($5, is_private),
			updated_at  = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING ` + selectClubColumns

	var row clubRow
	err := r.db.QueryRow(ctx, q, id, u.Name, u.Description, u.BookID, u.IsPrivate).Scan(
		&row.ID, &row.CreatorID, &row.BookID, &row.Name, &row.Description,
		&row.IsPrivate, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrClubNotFound
		}
		return nil, fmt.Errorf("update club: %w", err)
	}
	return row.toDomain(), nil
}

// SoftDelete marks the club as deleted.
func (r *ClubRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE clubs SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("soft delete club: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrClubNotFound
	}
	return nil
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
