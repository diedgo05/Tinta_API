// Package http exposes the Notification HTTP handlers.
package http

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tinta/notifications/internal/notification/application"
	"github.com/tinta/notifications/internal/notification/domain"
	"github.com/tinta/notifications/internal/notification/ports"
	"github.com/tinta/shared/httpx"
	"github.com/tinta/shared/middleware"
)

type CreateNotificationRequest struct {
	UserID string          `json:"user_id"`
	Type   string          `json:"type"`
	Title  string          `json:"title"`
	Body   string          `json:"body,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

type NotificationResponse struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Type      string          `json:"type"`
	Title     string          `json:"title"`
	Body      string          `json:"body,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Read      bool            `json:"read"`
	ReadAt    *time.Time      `json:"read_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type PaginatedNotificationsResponse struct {
	Items       []NotificationResponse `json:"items"`
	Total       int64                  `json:"total"`
	UnreadCount int64                  `json:"unread_count"`
	Page        int                    `json:"page"`
	PageSize    int                    `json:"page_size"`
}

type Handler struct {
	createUC      *application.CreateNotificationUseCase
	listUC        *application.ListNotificationsUseCase
	markReadUC    *application.MarkAsReadUseCase
	markAllReadUC *application.MarkAllAsReadUseCase
	deleteUC      *application.DeleteNotificationUseCase
}

func NewHandler(c *application.CreateNotificationUseCase, l *application.ListNotificationsUseCase,
	mr *application.MarkAsReadUseCase, mar *application.MarkAllAsReadUseCase,
	d *application.DeleteNotificationUseCase) *Handler {
	return &Handler{createUC: c, listUC: l, markReadUC: mr, markAllReadUC: mar, deleteUC: d}
}

func (h *Handler) Register(router fiber.Router, authMW fiber.Handler) {
	n := router.Group("/notifications", authMW)
	n.Get("/", h.list)
	n.Post("/", h.create) // creación interna (admin/sistema)
	n.Post("/:id/read", h.markRead)
	n.Post("/read-all", h.markAllRead)
	n.Delete("/:id", h.delete)
}

func (h *Handler) list(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	unread := c.Query("unread") == "true"
	res, err := h.listUC.Execute(c.Context(), ports.ListFilter{
		UserID: userID, UnreadOnly: unread, Page: page, PageSize: pageSize,
	})
	if err != nil {
		return mapErr(c, err)
	}
	items := make([]NotificationResponse, 0, len(res.Items))
	for _, n := range res.Items {
		items = append(items, toResp(n))
	}
	return httpx.OK(c, PaginatedNotificationsResponse{
		Items: items, Total: res.Total, UnreadCount: res.UnreadCount, Page: res.Page, PageSize: res.PageSize,
	})
}

func (h *Handler) create(c *fiber.Ctx) error {
	var b CreateNotificationRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid body")
	}
	uid, err := uuid.Parse(b.UserID)
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_USER_ID", "invalid user id")
	}
	var data any
	if len(b.Data) > 0 {
		data = b.Data
	}
	n, err := h.createUC.Execute(c.Context(), application.CreateNotificationInput{
		UserID: uid, Type: b.Type, Title: b.Title, Body: b.Body, Data: data,
	})
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.Created(c, toResp(n))
}

func (h *Handler) markRead(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid notification id")
	}
	n, err := h.markReadUC.Execute(c.Context(), id, userID)
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, toResp(n))
}

func (h *Handler) markAllRead(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	updated, err := h.markAllReadUC.Execute(c.Context(), userID)
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, fiber.Map{"updated": updated})
}

func (h *Handler) delete(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid notification id")
	}
	if err := h.deleteUC.Execute(c.Context(), id, userID); err != nil {
		return mapErr(c, err)
	}
	return httpx.NoContent(c)
}

func toResp(n *domain.Notification) NotificationResponse {
	return NotificationResponse{
		ID: n.ID.String(), UserID: n.UserID.String(), Type: n.Type, Title: n.Title,
		Body: n.Body, Data: n.Data, Read: n.IsRead(), ReadAt: n.ReadAt, CreatedAt: n.CreatedAt,
	}
}

func mapErr(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrNotificationNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "NOTIFICATION_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrNotAuthorized):
		return httpx.Error(c, fiber.StatusForbidden, "FORBIDDEN", err.Error())
	case errors.Is(err, domain.ErrInvalidTitle), errors.Is(err, domain.ErrInvalidType):
		return httpx.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
