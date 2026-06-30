// Package postgres implements ports.RecommendedBookRepository.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tinta/recommendations/internal/recommendedbook/domain"
)

type RecommendedBookRepository struct{ db *pgxpool.Pool }

func NewRecommendedBookRepository(db *pgxpool.Pool) *RecommendedBookRepository {
	return &RecommendedBookRepository{db: db}
}

type row struct {
	BookID         uuid.UUID
	GoogleVolumeID string
	Title          string
	Authors        []string
	Thumbnail      *string
	InfoLink       *string
	Description    *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (r row) toDomain() *domain.RecommendedBook {
	b := &domain.RecommendedBook{
		BookID:         r.BookID,
		GoogleVolumeID: r.GoogleVolumeID,
		Title:          r.Title,
		Authors:        r.Authors,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
	if r.Authors == nil {
		b.Authors = []string{}
	}
	if r.Thumbnail != nil {
		b.Thumbnail = *r.Thumbnail
	}
	if r.InfoLink != nil {
		b.InfoLink = *r.InfoLink
	}
	if r.Description != nil {
		b.Description = *r.Description
	}
	return b
}

const cols = `book_id, google_volume_id, title, authors, thumbnail, info_link, description, created_at, updated_at`

func (r *RecommendedBookRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.RecommendedBook, error) {
	const q = `SELECT ` + cols + ` FROM recommended_books WHERE book_id=$1`
	var rr row
	err := r.db.QueryRow(ctx, q, id).Scan(&rr.BookID, &rr.GoogleVolumeID, &rr.Title, &rr.Authors,
		&rr.Thumbnail, &rr.InfoLink, &rr.Description, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRecommendedBookNotFound
		}
		return nil, fmt.Errorf("get recommended book: %w", err)
	}
	return rr.toDomain(), nil
}

// GetByIDs returns metadata for many book IDs in a single round trip.
// Missing books are silently skipped (no error). Used by the list handler
// to enrich recommendations efficiently.
func (r *RecommendedBookRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domain.RecommendedBook, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]*domain.RecommendedBook{}, nil
	}
	const q = `SELECT ` + cols + ` FROM recommended_books WHERE book_id = ANY($1::uuid[])`
	rows, err := r.db.Query(ctx, q, ids)
	if err != nil {
		return nil, fmt.Errorf("query recommended books: %w", err)
	}
	defer rows.Close()
	out := make(map[uuid.UUID]*domain.RecommendedBook, len(ids))
	for rows.Next() {
		var rr row
		if err := rows.Scan(&rr.BookID, &rr.GoogleVolumeID, &rr.Title, &rr.Authors,
			&rr.Thumbnail, &rr.InfoLink, &rr.Description, &rr.CreatedAt, &rr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan recommended book: %w", err)
		}
		out[rr.BookID] = rr.toDomain()
	}
	return out, nil
}
