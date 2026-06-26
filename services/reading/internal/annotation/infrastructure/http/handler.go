// Package http exposes the Annotation HTTP handlers.
package http

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tinta/reading/internal/annotation/application"
	"github.com/tinta/reading/internal/annotation/domain"
	"github.com/tinta/reading/internal/annotation/ports"
	"github.com/tinta/shared/httpx"
	"github.com/tinta/shared/middleware"
)

type CreateAnnotationRequest struct {
	BookID          *string `json:"book_id,omitempty"`
	PersonalDocID   *string `json:"personal_doc_id,omitempty"`
	Page            int     `json:"page,omitempty"`
	HighlightedText string  `json:"highlighted_text"`
	PersonalNote    string  `json:"personal_note,omitempty"`
	Color           string  `json:"color,omitempty"`
}

type UpdateAnnotationRequest struct {
	PersonalNote *string `json:"personal_note,omitempty"`
	Color        *string `json:"color,omitempty"`
	Page         *int    `json:"page,omitempty"`
}

type AnnotationResponse struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	BookID          *string   `json:"book_id,omitempty"`
	PersonalDocID   *string   `json:"personal_doc_id,omitempty"`
	Page            int       `json:"page"`
	HighlightedText string    `json:"highlighted_text"`
	PersonalNote    string    `json:"personal_note,omitempty"`
	Color           string    `json:"color"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Handler struct {
	createUC *application.CreateAnnotationUseCase
	listUC   *application.ListAnnotationsUseCase
	updateUC *application.UpdateAnnotationUseCase
	deleteUC *application.DeleteAnnotationUseCase
}

func NewHandler(c *application.CreateAnnotationUseCase, l *application.ListAnnotationsUseCase,
	u *application.UpdateAnnotationUseCase, d *application.DeleteAnnotationUseCase) *Handler {
	return &Handler{createUC: c, listUC: l, updateUC: u, deleteUC: d}
}

func (h *Handler) Register(router fiber.Router, authMW fiber.Handler) {
	a := router.Group("/annotations", authMW)
	a.Post("/", h.create)
	a.Get("/", h.list)
	a.Patch("/:id", h.update)
	a.Delete("/:id", h.delete)
}

func (h *Handler) create(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	var b CreateAnnotationRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid body")
	}
	var bookID, docID *uuid.UUID
	if b.BookID != nil && *b.BookID != "" {
		id, err := uuid.Parse(*b.BookID)
		if err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BOOK_ID", "invalid book id")
		}
		bookID = &id
	}
	if b.PersonalDocID != nil && *b.PersonalDocID != "" {
		id, err := uuid.Parse(*b.PersonalDocID)
		if err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "INVALID_DOC_ID", "invalid doc id")
		}
		docID = &id
	}
	a, err := h.createUC.Execute(c.Context(), application.CreateAnnotationInput{
		UserID: userID, BookID: bookID, PersonalDocID: docID, Page: b.Page,
		HighlightedText: b.HighlightedText, PersonalNote: b.PersonalNote, Color: b.Color,
	})
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.Created(c, toResp(a))
}

func (h *Handler) list(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	f := ports.ListFilter{UserID: userID}
	if v := c.Query("book_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BOOK_ID", "invalid book id")
		}
		f.BookID = &id
	}
	if v := c.Query("personal_doc_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "INVALID_DOC_ID", "invalid doc id")
		}
		f.PersonalDocID = &id
	}
	items, err := h.listUC.Execute(c.Context(), f)
	if err != nil {
		return mapErr(c, err)
	}
	out := make([]AnnotationResponse, 0, len(items))
	for _, a := range items {
		out = append(out, toResp(a))
	}
	return httpx.OK(c, fiber.Map{"items": out, "total": len(out)})
}

func (h *Handler) update(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid annotation id")
	}
	var b UpdateAnnotationRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid body")
	}
	a, err := h.updateUC.Execute(c.Context(), application.UpdateAnnotationInput{
		AnnotationID: id, RequesterID: userID, PersonalNote: b.PersonalNote, Color: b.Color, Page: b.Page,
	})
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.OK(c, toResp(a))
}

func (h *Handler) delete(c *fiber.Ctx) error {
	userID, _ := middleware.UserIDFromContext(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid annotation id")
	}
	if err := h.deleteUC.Execute(c.Context(), id, userID); err != nil {
		return mapErr(c, err)
	}
	return httpx.NoContent(c)
}

func toResp(a *domain.Annotation) AnnotationResponse {
	var bID, dID *string
	if a.BookID != nil {
		s := a.BookID.String()
		bID = &s
	}
	if a.PersonalDocID != nil {
		s := a.PersonalDocID.String()
		dID = &s
	}
	return AnnotationResponse{
		ID: a.ID.String(), UserID: a.UserID.String(), BookID: bID, PersonalDocID: dID,
		Page: a.Page, HighlightedText: a.HighlightedText, PersonalNote: a.PersonalNote, Color: a.Color,
		CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
	}
}

func mapErr(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrAnnotationNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "ANNOTATION_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrNotAuthorized):
		return httpx.Error(c, fiber.StatusForbidden, "FORBIDDEN", err.Error())
	case errors.Is(err, domain.ErrEmptyHighlight), errors.Is(err, domain.ErrNoTarget):
		return httpx.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
