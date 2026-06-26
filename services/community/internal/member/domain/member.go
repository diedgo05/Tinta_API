// Package domain contains the ClubMember entity.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleMember    Role = "member"
	RoleModerator Role = "moderator"
	RoleOwner     Role = "owner"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleMember, RoleModerator, RoleOwner:
		return true
	}
	return false
}

type ClubMember struct {
	ID       uuid.UUID
	ClubID   uuid.UUID
	UserID   uuid.UUID
	Role     Role
	JoinedAt time.Time
}

var (
	ErrMemberNotFound     = errors.New("member not found")
	ErrAlreadyMember      = errors.New("user is already a member of this club")
	ErrNotMember          = errors.New("user is not a member of this club")
	ErrCannotLeaveAsOwner = errors.New("owner cannot leave the club; transfer ownership or delete the club")
	ErrInvalidRole        = errors.New("invalid role")
)
