// Package postgres implements ports.ProgressRepository.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tinta/reading/internal/progress/domain"
	"github.com/tinta/reading/internal/progress/ports"
)

type ProgressRepository struct{ db *pgxpool.Pool }

func NewProgressRepository(db *pgxpool.Pool) *ProgressRepository { return &ProgressRepository{db: db} }

type row struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	BookID       uuid.UUID
	CurrentPage  int32
	TotalTimeMin int32
	Status       string
	StartedAt    time.Time
	LastReadAt   time.Time
	FinishedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (r row) toDomain() *domain.Progress {
	return &domain.Progress{
		ID: r.ID, UserID: r.UserID, BookID: r.BookID,
		CurrentPage: int(r.CurrentPage), TotalTimeMin: int(r.TotalTimeMin),
		Status: domain.Status(r.Status), StartedAt: r.StartedAt, LastReadAt: r.LastReadAt,
		FinishedAt: r.FinishedAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

const cols = `id, user_id, book_id, current_page, total_time_min, status, started_at, last_read_at, finished_at, created_at, updated_at`

func (r *ProgressRepository) UpsertProgress(ctx context.Context, p *domain.Progress) (*domain.Progress, error) {
	const q = `INSERT INTO reading_progress (user_id, book_id, current_page, status)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (user_id, book_id) DO UPDATE
		SET current_page = GREATEST(reading_progress.current_page, EXCLUDED.current_page),
		    last_read_at = NOW(), updated_at = NOW()
		RETURNING ` + cols
	var rr row
	err := r.db.QueryRow(ctx, q, p.UserID, p.BookID, p.CurrentPage, string(p.Status)).Scan(
		&rr.ID, &rr.UserID, &rr.BookID, &rr.CurrentPage, &rr.TotalTimeMin, &rr.Status,
		&rr.StartedAt, &rr.LastReadAt, &rr.FinishedAt, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert progress: %w", err)
	}
	return rr.toDomain(), nil
}

func (r *ProgressRepository) GetByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (*domain.Progress, error) {
	const q = `SELECT ` + cols + ` FROM reading_progress WHERE user_id=$1 AND book_id=$2`
	var rr row
	err := r.db.QueryRow(ctx, q, userID, bookID).Scan(
		&rr.ID, &rr.UserID, &rr.BookID, &rr.CurrentPage, &rr.TotalTimeMin, &rr.Status,
		&rr.StartedAt, &rr.LastReadAt, &rr.FinishedAt, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProgressNotFound
		}
		return nil, err
	}
	return rr.toDomain(), nil
}

func (r *ProgressRepository) ListByUser(ctx context.Context, userID uuid.UUID, status *domain.Status) ([]*domain.Progress, error) {
	var rows pgx.Rows
	var err error
	if status != nil {
		rows, err = r.db.Query(ctx, `SELECT `+cols+` FROM reading_progress WHERE user_id=$1 AND status=$2 ORDER BY last_read_at DESC`, userID, string(*status))
	} else {
		rows, err = r.db.Query(ctx, `SELECT `+cols+` FROM reading_progress WHERE user_id=$1 ORDER BY last_read_at DESC`, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.Progress, 0)
	for rows.Next() {
		var rr row
		if err := rows.Scan(&rr.ID, &rr.UserID, &rr.BookID, &rr.CurrentPage, &rr.TotalTimeMin, &rr.Status,
			&rr.StartedAt, &rr.LastReadAt, &rr.FinishedAt, &rr.CreatedAt, &rr.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, rr.toDomain())
	}
	return out, nil
}

func (r *ProgressRepository) UpdateProgress(ctx context.Context, userID, bookID uuid.UUID, u ports.ProgressUpdates) (*domain.Progress, error) {
	const q = `UPDATE reading_progress SET
		current_page   = COALESCE($3, current_page),
		total_time_min = COALESCE($4, total_time_min),
		status         = COALESCE($5, status),
		finished_at    = CASE WHEN $5::varchar = 'finished' AND finished_at IS NULL THEN NOW() ELSE finished_at END,
		last_read_at   = NOW(),
		updated_at     = NOW()
	WHERE user_id=$1 AND book_id=$2 RETURNING ` + cols
	var st *string
	if u.Status != nil {
		s := string(*u.Status)
		st = &s
	}
	var cp, tt *int32
	if u.CurrentPage != nil {
		v := int32(*u.CurrentPage)
		cp = &v
	}
	if u.TotalTimeMin != nil {
		v := int32(*u.TotalTimeMin)
		tt = &v
	}
	var rr row
	err := r.db.QueryRow(ctx, q, userID, bookID, cp, tt, st).Scan(
		&rr.ID, &rr.UserID, &rr.BookID, &rr.CurrentPage, &rr.TotalTimeMin, &rr.Status,
		&rr.StartedAt, &rr.LastReadAt, &rr.FinishedAt, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProgressNotFound
		}
		return nil, err
	}
	return rr.toDomain(), nil
}

func (r *ProgressRepository) Delete(ctx context.Context, userID, bookID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM reading_progress WHERE user_id=$1 AND book_id=$2`, userID, bookID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrProgressNotFound
	}
	return nil
}
