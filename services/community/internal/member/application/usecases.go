// Package application contains the ClubMember use cases.
package application

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tinta/community/internal/member/domain"
	"github.com/tinta/community/internal/member/ports"
)

// ---------- Join ----------

type JoinClubUseCase struct{ repo ports.MemberRepository }

func NewJoinClubUseCase(r ports.MemberRepository) *JoinClubUseCase { return &JoinClubUseCase{repo: r} }

func (uc *JoinClubUseCase) Execute(ctx context.Context, clubID, userID uuid.UUID) (*domain.ClubMember, error) {
	// Check duplicate first to return a friendly error instead of a Postgres UNIQUE violation.
	if existing, err := uc.repo.GetMembership(ctx, clubID, userID); err == nil && existing != nil {
		return nil, domain.ErrAlreadyMember
	} else if err != nil && !errors.Is(err, domain.ErrMemberNotFound) {
		return nil, err
	}
	return uc.repo.Join(ctx, clubID, userID, domain.RoleMember)
}

// ---------- Leave ----------

type LeaveClubUseCase struct{ repo ports.MemberRepository }

func NewLeaveClubUseCase(r ports.MemberRepository) *LeaveClubUseCase {
	return &LeaveClubUseCase{repo: r}
}

func (uc *LeaveClubUseCase) Execute(ctx context.Context, clubID, userID uuid.UUID) error {
	existing, err := uc.repo.GetMembership(ctx, clubID, userID)
	if err != nil {
		return err
	}
	if existing.Role == domain.RoleOwner {
		return domain.ErrCannotLeaveAsOwner
	}
	return uc.repo.Leave(ctx, clubID, userID)
}

// ---------- Listings ----------

type ListClubMembersUseCase struct{ repo ports.MemberRepository }

func NewListClubMembersUseCase(r ports.MemberRepository) *ListClubMembersUseCase {
	return &ListClubMembersUseCase{repo: r}
}

func (uc *ListClubMembersUseCase) Execute(ctx context.Context, clubID uuid.UUID) ([]*domain.ClubMember, error) {
	return uc.repo.ListByClub(ctx, clubID)
}

type ListMyClubsUseCase struct{ repo ports.MemberRepository }

func NewListMyClubsUseCase(r ports.MemberRepository) *ListMyClubsUseCase {
	return &ListMyClubsUseCase{repo: r}
}

func (uc *ListMyClubsUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*domain.ClubMember, error) {
	return uc.repo.ListByUser(ctx, userID)
}

// ---------- Membership check ----------

type CheckMembershipUseCase struct{ repo ports.MemberRepository }

func NewCheckMembershipUseCase(r ports.MemberRepository) *CheckMembershipUseCase {
	return &CheckMembershipUseCase{repo: r}
}

func (uc *CheckMembershipUseCase) Execute(ctx context.Context, clubID, userID uuid.UUID) (*domain.ClubMember, error) {
	return uc.repo.GetMembership(ctx, clubID, userID)
}
