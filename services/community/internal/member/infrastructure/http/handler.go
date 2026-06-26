// Package http exposes the ClubMember HTTP handlers.
package http

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tinta/community/internal/member/application"
	"github.com/tinta/community/internal/member/domain"
	"github.com/tinta/shared/httpx"
	"github.com/tinta/shared/middleware"
)

type MemberResponse struct {
	ID       string    `json:"id"`
	ClubID   string    `json:"club_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type Handler struct {
	joinUC     *application.JoinClubUseCase
	leaveUC    *application.LeaveClubUseCase
	listClubUC *application.ListClubMembersUseCase
	listMyUC   *application.ListMyClubsUseCase
	checkUC    *application.CheckMembershipUseCase
}

func NewHandler(j *application.JoinClubUseCase, l *application.LeaveClubUseCase,
	lc *application.ListClubMembersUseCase, lm *application.ListMyClubsUseCase,
	c *application.CheckMembershipUseCase) *Handler {
	return &Handler{joinUC: j, leaveUC: l, listClubUC: lc, listMyUC: lm, checkUC: c}
}

// Register adds member routes nested under /api/v1.
// All require auth.
func (h *Handler) Register(router fiber.Router, authMW fiber.Handler) {
	g := router.Group("", authMW)
	g.Post("/clubs/:club_id/join", h.join)
	g.Delete("/clubs/:club_id/leave", h.leave)
	g.Get("/clubs/:club_id/members", h.listClubMembers)
	g.Get("/clubs/:club_id/membership", h.checkMyMembership)
	g.Get("/me/clubs", h.listMyClubs)
}

func (h *Handler) join(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	clubID, err := uuid.Parse(c.Params("club_id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_CLUB_ID", "invalid club id")
	}
	m, err := h.joinUC.Execute(c.Context(), clubID, userID)
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.Created(c, toResp(m))
}

func (h *Handler) leave(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	clubID, err := uuid.Parse(c.Params("club_id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_CLUB_ID", "invalid club id")
	}
	if err := h.leaveUC.Execute(c.Context(), clubID, userID); err != nil {
		return mapErr(c, err)
	}
	return httpx.NoContent(c)
}

func (h *Handler) listClubMembers(c *fiber.Ctx) error {
	clubID, err := uuid.Parse(c.Params("club_id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_CLUB_ID", "invalid club id")
	}
	items, err := h.listClubUC.Execute(c.Context(), clubID)
	if err != nil {
		return mapErr(c, err)
	}
	out := make([]MemberResponse, 0, len(items))
	for _, m := range items {
		out = append(out, toResp(m))
	}
	return httpx.OK(c, fiber.Map{"items": out, "total": len(out)})
}

func (h *Handler) checkMyMembership(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	clubID, err := uuid.Parse(c.Params("club_id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_CLUB_ID", "invalid club id")
	}
	m, err := h.checkUC.Execute(c.Context(), clubID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			return httpx.OK(c, fiber.Map{"is_member": false})
		}
		return mapErr(c, err)
	}
	return httpx.OK(c, fiber.Map{"is_member": true, "membership": toResp(m)})
}

func (h *Handler) listMyClubs(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	items, err := h.listMyUC.Execute(c.Context(), userID)
	if err != nil {
		return mapErr(c, err)
	}
	out := make([]MemberResponse, 0, len(items))
	for _, m := range items {
		out = append(out, toResp(m))
	}
	return httpx.OK(c, fiber.Map{"items": out, "total": len(out)})
}

func toResp(m *domain.ClubMember) MemberResponse {
	return MemberResponse{
		ID: m.ID.String(), ClubID: m.ClubID.String(), UserID: m.UserID.String(),
		Role: string(m.Role), JoinedAt: m.JoinedAt,
	}
}

func mapErr(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrMemberNotFound), errors.Is(err, domain.ErrNotMember):
		return httpx.Error(c, fiber.StatusNotFound, "NOT_MEMBER", err.Error())
	case errors.Is(err, domain.ErrAlreadyMember):
		return httpx.Error(c, fiber.StatusConflict, "ALREADY_MEMBER", err.Error())
	case errors.Is(err, domain.ErrCannotLeaveAsOwner):
		return httpx.Error(c, fiber.StatusForbidden, "OWNER_CANNOT_LEAVE", err.Error())
	case errors.Is(err, domain.ErrInvalidRole):
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ROLE", err.Error())
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
