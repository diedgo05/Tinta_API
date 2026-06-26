// Package postgres implements ports.GenreRepository.
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
	"github.com/tinta/catalog/internal/genre/domain"
	"github.com/tinta/catalog/internal/genre/ports"
)

type GenreRepository struct{ db *pgxpool.Pool }

func NewGenreRepository(db *pgxpool.Pool) *GenreRepository { return &GenreRepository{db: db} }

type row struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r row) toDomain() *domain.Genre {
	g := &domain.Genre{ID: r.ID, Name: r.Name, Slug: r.Slug, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
	if r.Description != nil {
		g.Description = *r.Description
	}
	return g
}

const cols = `id, name, slug, description, created_at, updated_at`

func (r *GenreRepository) Create(ctx context.Context, g *domain.Genre) (*domain.Genre, error) {
	const q = `INSERT INTO genres (name, slug, description) VALUES ($1,$2,$3) RETURNING ` + cols
	var desc *string
	if g.Description != "" {
		desc = &g.Description
	}
	var rr row
	err := r.db.QueryRow(ctx, q, g.Name, g.Slug, desc).Scan(&rr.ID, &rr.Name, &rr.Slug, &rr.Description, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "genres_name_key") || strings.Contains(err.Error(), "genres_slug_key") {
			return nil, domain.ErrGenreAlreadyExists
		}
		return nil, fmt.Errorf("insert genre: %w", err)
	}
	return rr.toDomain(), nil
}

func (r *GenreRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Genre, error) {
	const q = `SELECT ` + cols + ` FROM genres WHERE id=$1`
	var rr row
	err := r.db.QueryRow(ctx, q, id).Scan(&rr.ID, &rr.Name, &rr.Slug, &rr.Description, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrGenreNotFound
		}
		return nil, err
	}
	return rr.toDomain(), nil
}

func (r *GenreRepository) List(ctx context.Context) ([]*domain.Genre, error) {
	const q = `SELECT ` + cols + ` FROM genres ORDER BY name ASC`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.Genre, 0)
	for rows.Next() {
		var rr row
		if err := rows.Scan(&rr.ID, &rr.Name, &rr.Slug, &rr.Description, &rr.CreatedAt, &rr.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, rr.toDomain())
	}
	return out, nil
}

func (r *GenreRepository) Update(ctx context.Context, id uuid.UUID, u ports.GenreUpdates) (*domain.Genre, error) {
	const q = `UPDATE genres SET
		name=COALESCE($2, name),
		slug=CASE WHEN $2::varchar IS NOT NULL THEN $3 ELSE slug END,
		description=COALESCE($4, description),
		updated_at=NOW()
	WHERE id=$1 RETURNING ` + cols
	var slug *string
	if u.Name != nil {
		s := domain.Slugify(*u.Name)
		slug = &s
	}
	var rr row
	err := r.db.QueryRow(ctx, q, id, u.Name, slug, u.Description).Scan(&rr.ID, &rr.Name, &rr.Slug, &rr.Description, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrGenreNotFound
		}
		return nil, err
	}
	return rr.toDomain(), nil
}

func (r *GenreRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM genres WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrGenreNotFound
	}
	return nil
}
