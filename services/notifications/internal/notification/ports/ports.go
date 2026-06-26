// Package ports declares the NotificationRepository interface.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/notifications/internal/notification/domain"
)

type ListFilter struct {
	UserID     uuid.UUID
	UnreadOnly bool
	Page       int
	PageSize   int
}

type ListResult struct {
	Items       []*domain.Notification
	Total       int64
	UnreadCount int64
	Page        int
	PageSize    int
}

type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) (*domain.Notification, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	List(ctx context.Context, f ListFilter) (*ListResult, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
