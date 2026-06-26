// Package postgres implements ports.DiscussionRepository.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tinta/community/internal/discussion/domain"
	"github.com/tinta/community/internal/discussion/ports"
)

type DiscussionRepository struct{ db *pgxpool.Pool }

func NewDiscussionRepository(db *pgxpool.Pool) *DiscussionRepository {
	return &DiscussionRepository{db: db}
}

type row struct {
	ID            uuid.UUID
	ClubID        uuid.UUID
	UserID        uuid.UUID
	ChapterNumber *int32
	Content       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (r row) toDomain() *domain.Discussion {
	d := &domain.Discussion{
		ID: r.ID, ClubID: r.ClubID, UserID: r.UserID,
		Content: r.Content, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
	if r.ChapterNumber != nil {
		v := int(*r.ChapterNumber)
		d.ChapterNumber = &v
	}
	return d
}

const cols = `id, club_id, user_id, chapter_number, content, created_at, updated_at`

func (r *DiscussionRepository) Create(ctx context.Context, d *domain.Discussion) (*domain.Discussion, error) {
	const q = `INSERT INTO discussions (club_id, user_id, chapter_number, content)
		VALUES ($1,$2,$3,$4) RETURNING ` + cols
	var ch *int32
	if d.ChapterNumber != nil {
		v := int32(*d.ChapterNumber)
		ch = &v
	}
	var rr row
	err := r.db.QueryRow(ctx, q, d.ClubID, d.UserID, ch, d.Content).Scan(
		&rr.ID, &rr.ClubID, &rr.UserID, &rr.ChapterNumber, &rr.Content, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert discussion: %w", err)
	}
	return rr.toDomain(), nil
}

func (r *DiscussionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Discussion, error) {
	const q = `SELECT ` + cols + ` FROM discussions WHERE id=$1 AND deleted_at IS NULL`
	var rr row
	err := r.db.QueryRow(ctx, q, id).Scan(&rr.ID, &rr.ClubID, &rr.UserID, &rr.ChapterNumber,
		&rr.Content, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDiscussionNotFound
		}
		return nil, err
	}
	return rr.toDomain(), nil
}

func (r *DiscussionRepository) List(ctx context.Context, f ports.ListFilter) (*ports.ListResult, error) {
	args := []any{f.ClubID}
	whereExtra := ""
	if f.ChapterNumber != nil {
		whereExtra = " AND chapter_number=$2"
		args = append(args, *f.ChapterNumber)
	}

	var total int64
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM discussions WHERE club_id=$1 AND deleted_at IS NULL`+whereExtra, args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := (f.Page - 1) * f.PageSize
	limOff := fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, f.PageSize, offset)

	q := `SELECT ` + cols + ` FROM discussions WHERE club_id=$1 AND deleted_at IS NULL` + whereExtra + limOff
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]*domain.Discussion, 0)
	for rows.Next() {
		var rr row
		if err := rows.Scan(&rr.ID, &rr.ClubID, &rr.UserID, &rr.ChapterNumber,
			&rr.Content, &rr.CreatedAt, &rr.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, rr.toDomain())
	}
	return &ports.ListResult{Items: items, Total: total, Page: f.Page, PageSize: f.PageSize}, nil
}

func (r *DiscussionRepository) UpdateContent(ctx context.Context, id uuid.UUID, content string) (*domain.Discussion, error) {
	const q = `UPDATE discussions SET content=$2, updated_at=NOW()
		WHERE id=$1 AND deleted_at IS NULL RETURNING ` + cols
	var rr row
	err := r.db.QueryRow(ctx, q, id, content).Scan(&rr.ID, &rr.ClubID, &rr.UserID, &rr.ChapterNumber,
		&rr.Content, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDiscussionNotFound
		}
		return nil, err
	}
	return rr.toDomain(), nil
}

func (r *DiscussionRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `UPDATE discussions SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrDiscussionNotFound
	}
	return nil
}
