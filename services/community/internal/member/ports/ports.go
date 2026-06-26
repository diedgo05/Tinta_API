// Package ports declares the ClubMember repository interface.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/community/internal/member/domain"
)

type MemberRepository interface {
	Join(ctx context.Context, clubID, userID uuid.UUID, role domain.Role) (*domain.ClubMember, error)
	Leave(ctx context.Context, clubID, userID uuid.UUID) error
	GetMembership(ctx context.Context, clubID, userID uuid.UUID) (*domain.ClubMember, error)
	ListByClub(ctx context.Context, clubID uuid.UUID) ([]*domain.ClubMember, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.ClubMember, error)
	CountMembers(ctx context.Context, clubID uuid.UUID) (int64, error)
}
