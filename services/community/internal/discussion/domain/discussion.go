// Package domain contains the Discussion (chat message) entity.
package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Discussion struct {
	ID            uuid.UUID
	ClubID        uuid.UUID
	UserID        uuid.UUID
	ChapterNumber *int
	Content       string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (d *Discussion) CanBeManagedBy(userID uuid.UUID) bool { return d.UserID == userID }

var (
	ErrDiscussionNotFound = errors.New("discussion not found")
	ErrNotAuthorized      = errors.New("not authorized")
	ErrEmptyContent       = errors.New("content is required")
	ErrContentTooLong     = errors.New("content is too long (max 4000 chars)")
)

const MaxContentLen = 4000

func ValidateContent(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return ErrEmptyContent
	}
	if len(s) > MaxContentLen {
		return ErrContentTooLong
	}
	return nil
}
