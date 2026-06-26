// Package http exposes the Topic HTTP handlers.
package http

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tinta/knowledge/internal/topic/application"
	"github.com/tinta/knowledge/internal/topic/domain"
	"github.com/tinta/shared/httpx"
	"github.com/tinta/shared/middleware"
)

type TopicResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	SizeMB      int    `json:"size_mb"`
	Version     string `json:"version"`
}

type UserTopicResponse struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	TopicID      string     `json:"topic_id"`
	Downloaded   bool       `json:"downloaded"`
	SelectedAt   time.Time  `json:"selected_at"`
	DownloadedAt *time.Time `json:"downloaded_at,omitempty"`
	Version      string     `json:"version"`
}

type SelectTopicsRequest struct {
	TopicIDs []string `json:"topic_ids"`
}

type Handler struct {
	listUC         *application.ListTopicsUseCase
	getUC          *application.GetTopicUseCase
	selectUC       *application.SelectTopicsUseCase
	listMyUC       *application.ListMyTopicsUseCase
	markDownloadUC *application.MarkDownloadedUseCase
	removeUC       *application.RemoveTopicUseCase
}

func NewHandler(l *application.ListTopicsUseCase, g *application.GetTopicUseCase,
	s *application.SelectTopicsUseCase, lm *application.ListMyTopicsUseCase,
	md *application.MarkDownloadedUseCase, rm *application.RemoveTopicUseCase) *Handler {
	return &Handler{listUC: l, getUC: g, selectUC: s, listMyUC: lm, markDownloadUC: md, removeUC: rm}
}

func (h *Handler) Register(router fiber.Router, authMW fiber.Handler) {
	// Catálogo público (no requiere auth)
	router.Get("/topics", h.list)
	router.Get("/topics/:id", h.get)

	// Selección del usuario (requiere auth)
	router.Put("/topics/me", authMW, h.selectMyTopics)
	router.Get("/topics/me", authMW, h.listMyTopics)
	router.Post("/topics/me/:topic_id/downloaded", authMW, h.markDownloaded)
	router.Delete("/topics/me/:topic_id", authMW, h.removeMyTopic)
}

func (h *Handler) list(c *fiber.Ctx) error {
	items, err := h.listUC.Execute(c.Context())
	if err != nil {
		return mapErr(c, err)
	}
	out := make([]TopicResponse, 0, len(items))
	for _, t := range items {
		out = append(out, toTopicResp(t))
	}
	return httpx.OK(c, fiber.Map{"items": out, "total": len(out)})
}

func (h *Handler) get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid topic id")
	}
	t, err := h.getUC.Execute(c.Context(), id)
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, toTopicResp(t))
}

func (h *Handler) selectMyTopics(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	var b SelectTopicsRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid body")
	}
	ids := make([]uuid.UUID, 0, len(b.TopicIDs))
	for _, s := range b.TopicIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "INVALID_TOPIC_ID", "invalid topic id: "+s)
		}
		ids = append(ids, id)
	}
	items, err := h.selectUC.Execute(c.Context(), userID, ids)
	if err != nil {
		return mapErr(c, err)
	}
	out := make([]UserTopicResponse, 0, len(items))
	for _, ut := range items {
		out = append(out, toUTResp(ut))
	}
	return httpx.OK(c, fiber.Map{"items": out, "total": len(out)})
}

func (h *Handler) listMyTopics(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	items, err := h.listMyUC.Execute(c.Context(), userID)
	if err != nil {
		return mapErr(c, err)
	}
	out := make([]UserTopicResponse, 0, len(items))
	for _, ut := range items {
		out = append(out, toUTResp(ut))
	}
	return httpx.OK(c, fiber.Map{"items": out, "total": len(out)})
}

func (h *Handler) markDownloaded(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	tid, err := uuid.Parse(c.Params("topic_id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid topic id")
	}
	ut, err := h.markDownloadUC.Execute(c.Context(), userID, tid)
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, toUTResp(ut))
}

func (h *Handler) removeMyTopic(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	tid, err := uuid.Parse(c.Params("topic_id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid topic id")
	}
	if err := h.removeUC.Execute(c.Context(), userID, tid); err != nil {
		return mapErr(c, err)
	}
	return httpx.NoContent(c)
}

func toTopicResp(t *domain.Topic) TopicResponse {
	return TopicResponse{
		ID: t.ID.String(), Name: t.Name, Slug: t.Slug, Description: t.Description,
		Icon: t.Icon, SizeMB: t.SizeMB, Version: t.Version,
	}
}

func toUTResp(ut *domain.UserTopic) UserTopicResponse {
	return UserTopicResponse{
		ID: ut.ID.String(), UserID: ut.UserID.String(), TopicID: ut.TopicID.String(),
		Downloaded: ut.Downloaded, SelectedAt: ut.SelectedAt, DownloadedAt: ut.DownloadedAt, Version: ut.Version,
	}
}

func mapErr(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrTopicNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "TOPIC_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrUserTopicNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "USER_TOPIC_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrTooFewTopics), errors.Is(err, domain.ErrTooManyTopics):
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_SELECTION", err.Error())
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
