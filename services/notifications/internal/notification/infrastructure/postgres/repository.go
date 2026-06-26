// Package postgres implements ports.NotificationRepository.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tinta/notifications/internal/notification/domain"
	"github.com/tinta/notifications/internal/notification/ports"
)

type NotificationRepository struct{ db *pgxpool.Pool }

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

type row struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      string
	Title     string
	Body      *string
	Data      []byte
	ReadAt    *time.Time
	CreatedAt time.Time
}

func (r row) toDomain() *domain.Notification {
	n := &domain.Notification{
		ID: r.ID, UserID: r.UserID, Type: r.Type, Title: r.Title,
		ReadAt: r.ReadAt, CreatedAt: r.CreatedAt,
	}
	if r.Body != nil {
		n.Body = *r.Body
	}
	if r.Data != nil {
		n.Data = json.RawMessage(r.Data)
	}
	return n
}

const cols = `id, user_id, type, title, body, data, read_at, created_at`

func (r *NotificationRepository) Create(ctx context.Context, n *domain.Notification) (*domain.Notification, error) {
	const q = `INSERT INTO notifications (user_id, type, title, body, data)
		VALUES ($1,$2,$3,$4,$5) RETURNING ` + cols
	var body *string
	if n.Body != "" {
		body = &n.Body
	}
	var data []byte
	if n.Data != nil {
		data = []byte(n.Data)
	}
	var rr row
	err := r.db.QueryRow(ctx, q, n.UserID, n.Type, n.Title, body, data).Scan(
		&rr.ID, &rr.UserID, &rr.Type, &rr.Title, &rr.Body, &rr.Data, &rr.ReadAt, &rr.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert notification: %w", err)
	}
	return rr.toDomain(), nil
}

func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	var rr row
	err := r.db.QueryRow(ctx, `SELECT `+cols+` FROM notifications WHERE id=$1`, id).Scan(
		&rr.ID, &rr.UserID, &rr.Type, &rr.Title, &rr.Body, &rr.Data, &rr.ReadAt, &rr.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotificationNotFound
		}
		return nil, err
	}
	return rr.toDomain(), nil
}

func (r *NotificationRepository) List(ctx context.Context, f ports.ListFilter) (*ports.ListResult, error) {
	args := []any{f.UserID}
	whereExtra := ""
	if f.UnreadOnly {
		whereExtra = " AND read_at IS NULL"
	}

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id=$1`+whereExtra, args...).Scan(&total); err != nil {
		return nil, err
	}
	var unread int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND read_at IS NULL`, f.UserID).Scan(&unread); err != nil {
		return nil, err
	}
	offset := (f.Page - 1) * f.PageSize
	args = append(args, f.PageSize, offset)
	q := fmt.Sprintf("SELECT %s FROM notifications WHERE user_id=$1%s ORDER BY created_at DESC LIMIT $2 OFFSET $3", cols, whereExtra)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]*domain.Notification, 0)
	for rows.Next() {
		var rr row
		if err := rows.Scan(&rr.ID, &rr.UserID, &rr.Type, &rr.Title, &rr.Body, &rr.Data, &rr.ReadAt, &rr.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, rr.toDomain())
	}
	return &ports.ListResult{Items: items, Total: total, UnreadCount: unread, Page: f.Page, PageSize: f.PageSize}, nil
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	const q = `UPDATE notifications SET read_at = COALESCE(read_at, NOW()) WHERE id=$1 RETURNING ` + cols
	var rr row
	err := r.db.QueryRow(ctx, q, id).Scan(
		&rr.ID, &rr.UserID, &rr.Type, &rr.Title, &rr.Body, &rr.Data, &rr.ReadAt, &rr.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotificationNotFound
		}
		return nil, err
	}
	return rr.toDomain(), nil
}

func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	tag, err := r.db.Exec(ctx, `UPDATE notifications SET read_at=NOW() WHERE user_id=$1 AND read_at IS NULL`, userID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *NotificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM notifications WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotificationNotFound
	}
	return nil
}
