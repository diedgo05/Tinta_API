package domain

import "errors"

// Sentinel errors used by the application layer.
// Wrap (don't replace) so the layer above can use errors.Is.
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrPasswordTooWeak    = errors.New("password too weak")
	ErrNameTooShort       = errors.New("name too short")
	ErrInvalidLanguage    = errors.New("invalid language code")
)
