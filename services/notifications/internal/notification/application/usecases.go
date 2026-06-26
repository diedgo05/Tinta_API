// Package application contains Notification use cases.
package application

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/tinta/notifications/internal/notification/domain"
	"github.com/tinta/notifications/internal/notification/ports"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type CreateNotificationInput struct {
	UserID uuid.UUID
	Type   string
	Title  string
	Body   string
	Data   any
}

type CreateNotificationUseCase struct{ repo ports.NotificationRepository }

func NewCreateNotificationUseCase(r ports.NotificationRepository) *CreateNotificationUseCase {
	return &CreateNotificationUseCase{repo: r}
}

func (uc *CreateNotificationUseCase) Execute(ctx context.Context, in CreateNotificationInput) (*domain.Notification, error) {
	if err := domain.ValidateTitle(in.Title); err != nil {
		return nil, err
	}
	if err := domain.ValidateType(in.Type); err != nil {
		return nil, err
	}
	var data json.RawMessage
	if in.Data != nil {
		raw, err := json.Marshal(in.Data)
		if err != nil {
			return nil, err
		}
		data = raw
	}
	return uc.repo.Create(ctx, &domain.Notification{
		UserID: in.UserID, Type: in.Type, Title: in.Title, Body: in.Body, Data: data,
	})
}

type ListNotificationsUseCase struct{ repo ports.NotificationRepository }

func NewListNotificationsUseCase(r ports.NotificationRepository) *ListNotificationsUseCase {
	return &ListNotificationsUseCase{repo: r}
}

func (uc *ListNotificationsUseCase) Execute(ctx context.Context, f ports.ListFilter) (*ports.ListResult, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = defaultPageSize
	}
	if f.PageSize > maxPageSize {
		f.PageSize = maxPageSize
	}
	return uc.repo.List(ctx, f)
}

type MarkAsReadUseCase struct{ repo ports.NotificationRepository }

func NewMarkAsReadUseCase(r ports.NotificationRepository) *MarkAsReadUseCase {
	return &MarkAsReadUseCase{repo: r}
}

func (uc *MarkAsReadUseCase) Execute(ctx context.Context, id, requesterID uuid.UUID) (*domain.Notification, error) {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !existing.CanBeManagedBy(requesterID) {
		return nil, domain.ErrNotAuthorized
	}
	return uc.repo.MarkAsRead(ctx, id)
}

type MarkAllAsReadUseCase struct{ repo ports.NotificationRepository }

func NewMarkAllAsReadUseCase(r ports.NotificationRepository) *MarkAllAsReadUseCase {
	return &MarkAllAsReadUseCase{repo: r}
}

func (uc *MarkAllAsReadUseCase) Execute(ctx context.Context, userID uuid.UUID) (int64, error) {
	return uc.repo.MarkAllAsRead(ctx, userID)
}

type DeleteNotificationUseCase struct{ repo ports.NotificationRepository }

func NewDeleteNotificationUseCase(r ports.NotificationRepository) *DeleteNotificationUseCase {
	return &DeleteNotificationUseCase{repo: r}
}

func (uc *DeleteNotificationUseCase) Execute(ctx context.Context, id, requesterID uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !existing.CanBeManagedBy(requesterID) {
		return domain.ErrNotAuthorized
	}
	return uc.repo.Delete(ctx, id)
}
