// Package domain contains the Club aggregate.
package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Club is the core aggregate of the Community service.
type Club struct {
	ID          uuid.UUID
	CreatorID   uuid.UUID
	BookID      *uuid.UUID // optional reference to a book in the catalog service
	Name        string
	Description string
	IsPrivate   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CanBeManagedBy returns true if the given user is the creator.
// (Moderators support will come in a later iteration with the members table.)
func (c *Club) CanBeManagedBy(userID uuid.UUID) bool {
	return c.CreatorID == userID
}

// Sentinel errors.
var (
	ErrClubNotFound     = errors.New("club not found")
	ErrNotAuthorized    = errors.New("not authorized to manage this club")
	ErrInvalidClubName  = errors.New("invalid club name")
	ErrInvalidPageSize  = errors.New("invalid page size")
)

// ValidateName ensures the club name is usable.
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 3 || len(name) > 120 {
		return ErrInvalidClubName
	}
	return nil
}
