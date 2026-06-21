// Package domain contains the User aggregate and its business rules.
// This layer has NO dependencies on databases, frameworks, or external libraries
// other than the standard library and small primitives like uuid.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a user role inside Tinta.
type Role string

const (
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
	RoleSystem    Role = "system"
)

// IsValid reports whether the role is one of the recognized values.
func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleModerator, RoleAdmin, RoleSystem:
		return true
	}
	return false
}

// User is the core aggregate of the Identity service.
// All authentication and profile logic revolves around this entity.
type User struct {
	ID            uuid.UUID
	Email         string
	PasswordHash  string
	Name          string
	Role          Role
	EmailVerified bool
	AvatarURL     string
	Language      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// IsActive reports whether the user can authenticate.
// (Soft-deleted users are not returned by the repository, so this is mostly
// for explicit role checks in the future.)
func (u *User) IsActive() bool {
	return u.Role.IsValid()
}

// CanModerate reports whether the user has moderator-level privileges.
func (u *User) CanModerate() bool {
	return u.Role == RoleModerator || u.Role == RoleAdmin || u.Role == RoleSystem
}

// CanAdmin reports whether the user has admin-level privileges.
func (u *User) CanAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleSystem
}
