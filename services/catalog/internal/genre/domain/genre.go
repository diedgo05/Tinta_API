// Package domain contains the Genre aggregate.
package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Genre struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var (
	ErrGenreNotFound      = errors.New("genre not found")
	ErrGenreAlreadyExists = errors.New("genre already exists")
	ErrInvalidGenreName   = errors.New("invalid genre name")
)

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func ValidateName(n string) error {
	n = strings.TrimSpace(n)
	if len(n) < 2 || len(n) > 120 {
		return ErrInvalidGenreName
	}
	return nil
}

// Slugify converts a name into a URL-safe slug.
func Slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, "-")
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// ValidateSlug checks slug format.
func ValidateSlug(s string) bool {
	return slugRegex.MatchString(s)
}
