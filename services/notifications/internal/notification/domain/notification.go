// Package domain contains the Notification aggregate.
package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      string
	Title     string
	Body      string
	Data      json.RawMessage
	ReadAt    *time.Time
	CreatedAt time.Time
}

func (n *Notification) IsRead() bool { return n.ReadAt != nil }

func (n *Notification) CanBeManagedBy(userID uuid.UUID) bool { return n.UserID == userID }

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrNotAuthorized        = errors.New("not authorized")
	ErrInvalidTitle         = errors.New("invalid title")
	ErrInvalidType          = errors.New("invalid type")
)

func ValidateTitle(t string) error {
	t = strings.TrimSpace(t)
	if len(t) < 1 || len(t) > 255 {
		return ErrInvalidTitle
	}
	return nil
}

func ValidateType(t string) error {
	if strings.TrimSpace(t) == "" {
		return ErrInvalidType
	}
	return nil
}
