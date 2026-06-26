// Package application contains the Discussion use cases.
package application

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/tinta/community/internal/discussion/domain"
	"github.com/tinta/community/internal/discussion/ports"
	memberDomain "github.com/tinta/community/internal/member/domain"
	memberPorts "github.com/tinta/community/internal/member/ports"
)

// MembershipChecker validates a user belongs to a club before posting/reading.
// We accept a function adapter rather than the whole MemberRepository to keep
// modules loosely coupled; main.go wires it.
type MembershipChecker interface {
	GetMembership(ctx context.Context, clubID, userID uuid.UUID) (*memberDomain.ClubMember, error)
}

// ---------- Post (create) ----------

type PostDiscussionInput struct {
	ClubID        uuid.UUID
	UserID        uuid.UUID
	ChapterNumber *int
	Content       string
}

type PostDiscussionUseCase struct {
	repo    ports.DiscussionRepository
	members memberPorts.MemberRepository
}

func NewPostDiscussionUseCase(r ports.DiscussionRepository, m memberPorts.MemberRepository) *PostDiscussionUseCase {
	return &PostDiscussionUseCase{repo: r, members: m}
}

func (uc *PostDiscussionUseCase) Execute(ctx context.Context, in PostDiscussionInput) (*domain.Discussion, error) {
	if err := domain.ValidateContent(in.Content); err != nil {
		return nil, err
	}
	// Membership check: only members of the club can post.
	if _, err := uc.members.GetMembership(ctx, in.ClubID, in.UserID); err != nil {
		if errors.Is(err, memberDomain.ErrMemberNotFound) {
			return nil, domain.ErrNotAuthorized
		}
		return nil, err
	}
	return uc.repo.Create(ctx, &domain.Discussion{
		ClubID: in.ClubID, UserID: in.UserID,
		ChapterNumber: in.ChapterNumber,
		Content:       strings.TrimSpace(in.Content),
	})
}

// ---------- List ----------

type ListDiscussionsUseCase struct {
	repo    ports.DiscussionRepository
	members memberPorts.MemberRepository
}

func NewListDiscussionsUseCase(r ports.DiscussionRepository, m memberPorts.MemberRepository) *ListDiscussionsUseCase {
	return &ListDiscussionsUseCase{repo: r, members: m}
}

func (uc *ListDiscussionsUseCase) Execute(ctx context.Context, requesterID uuid.UUID, f ports.ListFilter) (*ports.ListResult, error) {
	// Membership check: only members can read.
	if _, err := uc.members.GetMembership(ctx, f.ClubID, requesterID); err != nil {
		if errors.Is(err, memberDomain.ErrMemberNotFound) {
			return nil, domain.ErrNotAuthorized
		}
		return nil, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 50
	}
	if f.PageSize > 200 {
		f.PageSize = 200
	}
	return uc.repo.List(ctx, f)
}

// ---------- Update ----------

type UpdateDiscussionInput struct {
	DiscussionID uuid.UUID
	RequesterID  uuid.UUID
	Content      string
}

type UpdateDiscussionUseCase struct{ repo ports.DiscussionRepository }

func NewUpdateDiscussionUseCase(r ports.DiscussionRepository) *UpdateDiscussionUseCase {
	return &UpdateDiscussionUseCase{repo: r}
}

func (uc *UpdateDiscussionUseCase) Execute(ctx context.Context, in UpdateDiscussionInput) (*domain.Discussion, error) {
	if err := domain.ValidateContent(in.Content); err != nil {
		return nil, err
	}
	existing, err := uc.repo.GetByID(ctx, in.DiscussionID)
	if err != nil {
		return nil, err
	}
	if !existing.CanBeManagedBy(in.RequesterID) {
		return nil, domain.ErrNotAuthorized
	}
	return uc.repo.UpdateContent(ctx, in.DiscussionID, strings.TrimSpace(in.Content))
}

// ---------- Delete ----------

type DeleteDiscussionUseCase struct{ repo ports.DiscussionRepository }

func NewDeleteDiscussionUseCase(r ports.DiscussionRepository) *DeleteDiscussionUseCase {
	return &DeleteDiscussionUseCase{repo: r}
}

func (uc *DeleteDiscussionUseCase) Execute(ctx context.Context, id, requesterID uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !existing.CanBeManagedBy(requesterID) {
		return domain.ErrNotAuthorized
	}
	return uc.repo.SoftDelete(ctx, id)
}
