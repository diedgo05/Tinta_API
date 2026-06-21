// Package postgres implements ports.RecommendationRepository using pgx.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tinta/recommendations/internal/recommendation/domain"
	"github.com/tinta/recommendations/internal/recommendation/ports"
)

// RecommendationRepository persists recommendations.
type RecommendationRepository struct {
	db *pgxpool.Pool
}

// NewRecommendationRepository wires the pgx pool.
func NewRecommendationRepository(db *pgxpool.Pool) *RecommendationRepository {
	return &RecommendationRepository{db: db}
}

type recoRow struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	BookID       uuid.UUID
	Score        float64
	ClusterID    *int32
	Source       string
	Feedback     *string
	FeedbackAt   *time.Time
	DismissedAt  *time.Time
	GeneratedAt  time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (r recoRow) toDomain() *domain.Recommendation {
	var clusterID *int
	if r.ClusterID != nil {
		v := int(*r.ClusterID)
		clusterID = &v
	}
	var feedback *domain.Feedback
	if r.Feedback != nil {
		f := domain.Feedback(*r.Feedback)
		feedback = &f
	}
	return &domain.Recommendation{
		ID:          r.ID,
		UserID:      r.UserID,
		BookID:      r.BookID,
		Score:       r.Score,
		ClusterID:   clusterID,
		Source:      domain.Source(r.Source),
		Feedback:    feedback,
		FeedbackAt:  r.FeedbackAt,
		DismissedAt: r.DismissedAt,
		GeneratedAt: r.GeneratedAt,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

const selectRecoColumns = `
	id, user_id, book_id, score, cluster_id, source, feedback, feedback_at,
	dismissed_at, generated_at, created_at, updated_at`

// GetByID returns a recommendation by id.
func (r *RecommendationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Recommendation, error) {
	const q = `SELECT ` + selectRecoColumns + ` FROM recommendations WHERE id = $1`

	var row recoRow
	err := r.db.QueryRow(ctx, q, id).Scan(
		&row.ID, &row.UserID, &row.BookID, &row.Score, &row.ClusterID, &row.Source,
		&row.Feedback, &row.FeedbackAt, &row.DismissedAt, &row.GeneratedAt,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRecommendationNotFound
		}
		return nil, fmt.Errorf("get recommendation: %w", err)
	}
	return row.toDomain(), nil
}

// List returns a user's active recommendations sorted by score desc.
func (r *RecommendationRepository) List(ctx context.Context, f ports.ListFilter) (*ports.ListResult, error) {
	const countQ = `SELECT COUNT(*) FROM recommendations WHERE user_id = $1 AND dismissed_at IS NULL`
	var total int64
	if err := r.db.QueryRow(ctx, countQ, f.UserID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count recommendations: %w", err)
	}

	const listQ = `SELECT ` + selectRecoColumns + `
		FROM recommendations
		WHERE user_id = $1 AND dismissed_at IS NULL
		ORDER BY score DESC, generated_at DESC
		LIMIT $2 OFFSET $3`

	offset := (f.Page - 1) * f.PageSize
	rows, err := r.db.Query(ctx, listQ, f.UserID, f.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list recommendations: %w", err)
	}
	defer rows.Close()

	items := make([]*domain.Recommendation, 0)
	for rows.Next() {
		var row recoRow
		if err := rows.Scan(
			&row.ID, &row.UserID, &row.BookID, &row.Score, &row.ClusterID, &row.Source,
			&row.Feedback, &row.FeedbackAt, &row.DismissedAt, &row.GeneratedAt,
			&row.CreatedAt, &row.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan recommendation: %w", err)
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

// SetFeedback records the user's reaction.
func (r *RecommendationRepository) SetFeedback(ctx context.Context, id uuid.UUID, feedback domain.Feedback) (*domain.Recommendation, error) {
	const q = `
		UPDATE recommendations
		SET feedback = $2, feedback_at = NOW(), updated_at = NOW()
		WHERE id = $1
		RETURNING ` + selectRecoColumns

	var row recoRow
	err := r.db.QueryRow(ctx, q, id, string(feedback)).Scan(
		&row.ID, &row.UserID, &row.BookID, &row.Score, &row.ClusterID, &row.Source,
		&row.Feedback, &row.FeedbackAt, &row.DismissedAt, &row.GeneratedAt,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRecommendationNotFound
		}
		return nil, fmt.Errorf("set feedback: %w", err)
	}
	return row.toDomain(), nil
}

// Dismiss marks the recommendation as dismissed.
func (r *RecommendationRepository) Dismiss(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE recommendations SET dismissed_at = NOW(), updated_at = NOW() WHERE id = $1 AND dismissed_at IS NULL`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("dismiss recommendation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrRecommendationNotFound
	}
	return nil
}

// DeleteAllForUser hard-deletes every recommendation of a user.
// Used by the regeneration pipeline before producing a fresh set.
func (r *RecommendationRepository) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	const q = `DELETE FROM recommendations WHERE user_id = $1`
	if _, err := r.db.Exec(ctx, q, userID); err != nil {
		return fmt.Errorf("delete recommendations: %w", err)
	}
	return nil
}
