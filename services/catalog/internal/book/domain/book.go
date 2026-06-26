// Package domain contains the Book aggregate.
package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// License represents the legal status of a book.
type License string

const (
	LicensePublicDomain    License = "public_domain"
	LicenseCreativeCommons License = "creative_commons"
	LicenseCopyrighted     License = "copyrighted"
	LicenseUserOwned       License = "user_owned"
	LicenseUnknown         License = "unknown"
)

// IsValid reports whether the license is recognized.
func (l License) IsValid() bool {
	switch l {
	case LicensePublicDomain, LicenseCreativeCommons, LicenseCopyrighted, LicenseUserOwned, LicenseUnknown:
		return true
	}
	return false
}

// Book is the core aggregate of the Catalog service.
type Book struct {
	ID            uuid.UUID
	GenreID       *uuid.UUID
	Title         string
	Author        string
	ISBN          string
	Synopsis      string
	CoverURL      string
	TotalPages    int
	License       License
	Language      string
	PublishedYear *int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Sentinel errors.
var (
	ErrBookNotFound    = errors.New("book not found")
	ErrInvalidTitle    = errors.New("invalid title")
	ErrInvalidAuthor   = errors.New("invalid author")
	ErrInvalidLicense  = errors.New("invalid license")
	ErrInvalidPageSize = errors.New("invalid page size")
)

// ValidateTitle ensures a usable book title.
func ValidateTitle(t string) error {
	t = strings.TrimSpace(t)
	if len(t) < 1 || len(t) > 255 {
		return ErrInvalidTitle
	}
	return nil
}

// ValidateAuthor ensures a usable author name.
func ValidateAuthor(a string) error {
	a = strings.TrimSpace(a)
	if len(a) < 1 || len(a) > 255 {
		return ErrInvalidAuthor
	}
	return nil
}
