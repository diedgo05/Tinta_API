package domain

import (
	"regexp"
	"strings"
	"unicode"
)

// emailRegex is a pragmatic email format check (not full RFC 5322).
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// validLanguages contains the ISO 639-1 codes Tinta supports.
var validLanguages = map[string]bool{"es": true, "en": true}

// ValidateEmail checks the basic shape of an email address.
func ValidateEmail(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || !emailRegex.MatchString(email) || len(email) > 255 {
		return ErrInvalidEmail
	}
	return nil
}

// ValidateName ensures the user provided a usable display name.
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 120 {
		return ErrNameTooShort
	}
	return nil
}

// ValidatePassword enforces password rules:
//   - at least 8 characters
//   - at least one letter and one digit
func ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 128 {
		return ErrPasswordTooWeak
	}
	var hasLetter, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsLetter(ch):
			hasLetter = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return ErrPasswordTooWeak
	}
	return nil
}

// ValidateLanguage checks that the language code is one of the supported ones.
func ValidateLanguage(lang string) error {
	if lang == "" {
		return nil // optional
	}
	if !validLanguages[lang] {
		return ErrInvalidLanguage
	}
	return nil
}

// NormalizeEmail returns the canonical (lowercased, trimmed) version of an email.
func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}
