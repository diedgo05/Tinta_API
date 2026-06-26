// Package postgres implements topic and user-topic repositories.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tinta/knowledge/internal/topic/domain"
)

// ---------- TopicRepository ----------

type TopicRepository struct{ db *pgxpool.Pool }

func NewTopicRepository(db *pgxpool.Pool) *TopicRepository { return &TopicRepository{db: db} }

const topicCols = `id, name, slug, description, icon, size_mb, version, created_at, updated_at`

type topicRow struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description *string
	Icon        *string
	SizeMB      int32
	Version     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r topicRow) toDomain() *domain.Topic {
	t := &domain.Topic{
		ID: r.ID, Name: r.Name, Slug: r.Slug, SizeMB: int(r.SizeMB), Version: r.Version,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
	if r.Description != nil {
		t.Description = *r.Description
	}
	if r.Icon != nil {
		t.Icon = *r.Icon
	}
	return t
}

func (r *TopicRepository) List(ctx context.Context) ([]*domain.Topic, error) {
	rows, err := r.db.Query(ctx, `SELECT `+topicCols+` FROM topics ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.Topic, 0)
	for rows.Next() {
		var rr topicRow
		if err := rows.Scan(&rr.ID, &rr.Name, &rr.Slug, &rr.Description, &rr.Icon, &rr.SizeMB, &rr.Version, &rr.CreatedAt, &rr.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, rr.toDomain())
	}
	return out, nil
}

func (r *TopicRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Topic, error) {
	var rr topicRow
	err := r.db.QueryRow(ctx, `SELECT `+topicCols+` FROM topics WHERE id=$1`, id).Scan(
		&rr.ID, &rr.Name, &rr.Slug, &rr.Description, &rr.Icon, &rr.SizeMB, &rr.Version, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTopicNotFound
		}
		return nil, err
	}
	return rr.toDomain(), nil
}

func (r *TopicRepository) GetBySlug(ctx context.Context, slug string) (*domain.Topic, error) {
	var rr topicRow
	err := r.db.QueryRow(ctx, `SELECT `+topicCols+` FROM topics WHERE slug=$1`, slug).Scan(
		&rr.ID, &rr.Name, &rr.Slug, &rr.Description, &rr.Icon, &rr.SizeMB, &rr.Version, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTopicNotFound
		}
		return nil, err
	}
	return rr.toDomain(), nil
}

// ---------- UserTopicRepository ----------

type UserTopicRepository struct{ db *pgxpool.Pool }

func NewUserTopicRepository(db *pgxpool.Pool) *UserTopicRepository {
	return &UserTopicRepository{db: db}
}

const userTopicCols = `id, user_id, topic_id, downloaded, selected_at, downloaded_at, version`

type utRow struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	TopicID      uuid.UUID
	Downloaded   bool
	SelectedAt   time.Time
	DownloadedAt *time.Time
	Version      string
}

func (r utRow) toDomain() *domain.UserTopic {
	return &domain.UserTopic{
		ID: r.ID, UserID: r.UserID, TopicID: r.TopicID, Downloaded: r.Downloaded,
		SelectedAt: r.SelectedAt, DownloadedAt: r.DownloadedAt, Version: r.Version,
	}
}

func (r *UserTopicRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.UserTopic, error) {
	rows, err := r.db.Query(ctx, `SELECT `+userTopicCols+` FROM user_topics WHERE user_id=$1 ORDER BY selected_at ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.UserTopic, 0)
	for rows.Next() {
		var rr utRow
		if err := rows.Scan(&rr.ID, &rr.UserID, &rr.TopicID, &rr.Downloaded, &rr.SelectedAt, &rr.DownloadedAt, &rr.Version); err != nil {
			return nil, err
		}
		out = append(out, rr.toDomain())
	}
	return out, nil
}

// ReplaceUserSelection wipes existing rows and inserts the new selection in one transaction.
func (r *UserTopicRepository) ReplaceUserSelection(ctx context.Context, userID uuid.UUID, topicIDs []uuid.UUID) ([]*domain.UserTopic, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM user_topics WHERE user_id=$1`, userID); err != nil {
		return nil, fmt.Errorf("clear selection: %w", err)
	}

	out := make([]*domain.UserTopic, 0, len(topicIDs))
	for _, tid := range topicIDs {
		var rr utRow
		err := tx.QueryRow(ctx, `INSERT INTO user_topics (user_id, topic_id) VALUES ($1,$2) RETURNING `+userTopicCols,
			userID, tid).Scan(&rr.ID, &rr.UserID, &rr.TopicID, &rr.Downloaded, &rr.SelectedAt, &rr.DownloadedAt, &rr.Version)
		if err != nil {
			return nil, fmt.Errorf("insert user_topic: %w", err)
		}
		out = append(out, rr.toDomain())
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *UserTopicRepository) MarkDownloaded(ctx context.Context, userID, topicID uuid.UUID) (*domain.UserTopic, error) {
	const q = `UPDATE user_topics SET downloaded=TRUE, downloaded_at=NOW()
		WHERE user_id=$1 AND topic_id=$2 RETURNING ` + userTopicCols
	var rr utRow
	err := r.db.QueryRow(ctx, q, userID, topicID).Scan(
		&rr.ID, &rr.UserID, &rr.TopicID, &rr.Downloaded, &rr.SelectedAt, &rr.DownloadedAt, &rr.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserTopicNotFound
		}
		return nil, err
	}
	return rr.toDomain(), nil
}

func (r *UserTopicRepository) Remove(ctx context.Context, userID, topicID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM user_topics WHERE user_id=$1 AND topic_id=$2`, userID, topicID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserTopicNotFound
	}
	return nil
}
