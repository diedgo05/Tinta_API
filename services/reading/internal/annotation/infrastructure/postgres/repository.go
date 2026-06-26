// Package postgres implements ports.AnnotationRepository.
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
	"github.com/tinta/reading/internal/annotation/domain"
	"github.com/tinta/reading/internal/annotation/ports"
)

type AnnotationRepository struct{ db *pgxpool.Pool }

func NewAnnotationRepository(db *pgxpool.Pool) *AnnotationRepository {
	return &AnnotationRepository{db: db}
}

type row struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	BookID          *uuid.UUID
	PersonalDocID   *uuid.UUID
	Page            int32
	HighlightedText string
	PersonalNote    *string
	Color           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (r row) toDomain() *domain.Annotation {
	a := &domain.Annotation{
		ID: r.ID, UserID: r.UserID, BookID: r.BookID, PersonalDocID: r.PersonalDocID,
		Page: int(r.Page), HighlightedText: r.HighlightedText, Color: r.Color,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
	if r.PersonalNote != nil {
		a.PersonalNote = *r.PersonalNote
	}
	return a
}

const cols = `id, user_id, book_id, personal_doc_id, page, highlighted_text, personal_note, color, created_at, updated_at`

func (r *AnnotationRepository) Create(ctx context.Context, a *domain.Annotation) (*domain.Annotation, error) {
	const q = `INSERT INTO annotations (user_id, book_id, personal_doc_id, page, highlighted_text, personal_note, color)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING ` + cols
	var note *string
	if a.PersonalNote != "" {
		note = &a.PersonalNote
	}
	var rr row
	err := r.db.QueryRow(ctx, q, a.UserID, a.BookID, a.PersonalDocID, a.Page, a.HighlightedText, note, a.Color).Scan(
		&rr.ID, &rr.UserID, &rr.BookID, &rr.PersonalDocID, &rr.Page, &rr.HighlightedText, &rr.PersonalNote, &rr.Color,
		&rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert annotation: %w", err)
	}
	return rr.toDomain(), nil
}

func (r *AnnotationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Annotation, error) {
	const q = `SELECT ` + cols + ` FROM annotations WHERE id=$1 AND deleted_at IS NULL`
	var rr row
	err := r.db.QueryRow(ctx, q, id).Scan(&rr.ID, &rr.UserID, &rr.BookID, &rr.PersonalDocID, &rr.Page,
		&rr.HighlightedText, &rr.PersonalNote, &rr.Color, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAnnotationNotFound
		}
		return nil, err
	}
	return rr.toDomain(), nil
}

func (r *AnnotationRepository) List(ctx context.Context, f ports.ListFilter) ([]*domain.Annotation, error) {
	conds := []string{"user_id=$1", "deleted_at IS NULL"}
	args := []any{f.UserID}
	i := 2
	if f.BookID != nil {
		conds = append(conds, fmt.Sprintf("book_id=$%d", i))
		args = append(args, *f.BookID)
		i++
	}
	if f.PersonalDocID != nil {
		conds = append(conds, fmt.Sprintf("personal_doc_id=$%d", i))
		args = append(args, *f.PersonalDocID)
		i++
	}
	q := `SELECT ` + cols + ` FROM annotations WHERE ` + strings.Join(conds, " AND ") + ` ORDER BY page ASC, created_at ASC`
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.Annotation, 0)
	for rows.Next() {
		var rr row
		if err := rows.Scan(&rr.ID, &rr.UserID, &rr.BookID, &rr.PersonalDocID, &rr.Page,
			&rr.HighlightedText, &rr.PersonalNote, &rr.Color, &rr.CreatedAt, &rr.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, rr.toDomain())
	}
	return out, nil
}

func (r *AnnotationRepository) Update(ctx context.Context, id uuid.UUID, u ports.AnnotationUpdates) (*domain.Annotation, error) {
	const q = `UPDATE annotations SET
		personal_note = COALESCE($2, personal_note),
		color         = COALESCE($3, color),
		page          = COALESCE($4, page),
		updated_at    = NOW()
	WHERE id=$1 AND deleted_at IS NULL RETURNING ` + cols
	var p *int32
	if u.Page != nil {
		v := int32(*u.Page)
		p = &v
	}
	var rr row
	err := r.db.QueryRow(ctx, q, id, u.PersonalNote, u.Color, p).Scan(
		&rr.ID, &rr.UserID, &rr.BookID, &rr.PersonalDocID, &rr.Page,
		&rr.HighlightedText, &rr.PersonalNote, &rr.Color, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAnnotationNotFound
		}
		return nil, err
	}
	return rr.toDomain(), nil
}

func (r *AnnotationRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `UPDATE annotations SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAnnotationNotFound
	}
	return nil
}
