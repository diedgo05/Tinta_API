// Package domain contains the Annotation aggregate.
package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Annotation struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	BookID          *uuid.UUID
	PersonalDocID   *uuid.UUID
	Page            int
	HighlightedText string
	PersonalNote    string
	Color           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (a *Annotation) CanBeManagedBy(userID uuid.UUID) bool {
	return a.UserID == userID
}

var (
	ErrAnnotationNotFound = errors.New("annotation not found")
	ErrNotAuthorized      = errors.New("not authorized")
	ErrEmptyHighlight     = errors.New("highlighted text is required")
	ErrNoTarget           = errors.New("annotation needs a book_id or personal_doc_id")
)

func ValidateHighlight(s string) error {
	if strings.TrimSpace(s) == "" {
		return ErrEmptyHighlight
	}
	return nil
}
