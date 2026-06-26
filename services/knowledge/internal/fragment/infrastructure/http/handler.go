// Package http exposes Fragment and Document HTTP handlers.
package http

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tinta/knowledge/internal/fragment/application"
	"github.com/tinta/knowledge/internal/fragment/domain"
	"github.com/tinta/knowledge/internal/fragment/ports"
	"github.com/tinta/shared/httpx"
)

type FragmentResponse struct {
	ID         string          `json:"id"`
	DocumentID string          `json:"document_id"`
	TopicID    string          `json:"topic_id"`
	TextChunk  string          `json:"text_chunk"`
	Position   int             `json:"position"`
	Tokens     int             `json:"tokens"`
	Embedding  json.RawMessage `json:"embedding,omitempty"`
	HashChunk  string          `json:"hash_chunk"`
	CreatedAt  time.Time       `json:"created_at"`
}

type PaginatedFragmentsResponse struct {
	Items    []FragmentResponse `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

type DocumentResponse struct {
	ID          string    `json:"id"`
	TopicID     string    `json:"topic_id"`
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	License     string    `json:"license"`
	URLOriginal string    `json:"url_original,omitempty"`
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateDocumentRequest struct {
	TopicID     string `json:"topic_id"`
	Title       string `json:"title"`
	Source      string `json:"source"`
	License     string `json:"license,omitempty"`
	URLOriginal string `json:"url_original,omitempty"`
	Version     string `json:"version,omitempty"`
}

type CreateFragmentRequest struct {
	DocumentID string    `json:"document_id"`
	TopicID    string    `json:"topic_id"`
	TextChunk  string    `json:"text_chunk"`
	Position   int       `json:"position,omitempty"`
	Tokens     int       `json:"tokens,omitempty"`
	Embedding  []float32 `json:"embedding,omitempty"`
	HashChunk  string    `json:"hash_chunk"`
}

type Handler struct {
	listFragUC   *application.ListFragmentsUseCase
	createFragUC *application.CreateFragmentUseCase
	listDocUC    *application.ListDocumentsUseCase
	createDocUC  *application.CreateDocumentUseCase
}

func NewHandler(lf *application.ListFragmentsUseCase, cf *application.CreateFragmentUseCase,
	ld *application.ListDocumentsUseCase, cd *application.CreateDocumentUseCase) *Handler {
	return &Handler{listFragUC: lf, createFragUC: cf, listDocUC: ld, createDocUC: cd}
}

func (h *Handler) Register(router fiber.Router, authMW fiber.Handler) {
	// Endpoints autenticados (el celular descarga los fragmentos por tema)
	router.Get("/topics/:topic_id/fragments", authMW, h.listFragments)
	router.Get("/topics/:topic_id/documents", authMW, h.listDocuments)

	// Endpoints de administración (poblar la base de conocimientos desde el pipeline Python)
	router.Post("/documents", authMW, h.createDocument)
	router.Post("/fragments", authMW, h.createFragment)
}

func (h *Handler) listFragments(c *fiber.Ctx) error {
	topicID, err := uuid.Parse(c.Params("topic_id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid topic id")
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "100"))
	res, err := h.listFragUC.Execute(c.Context(), ports.FragmentListFilter{
		TopicID: topicID, Page: page, PageSize: pageSize,
	})
	if err != nil {
		return mapErr(c, err)
	}
	items := make([]FragmentResponse, 0, len(res.Items))
	for _, f := range res.Items {
		items = append(items, toFragResp(f))
	}
	return httpx.OK(c, PaginatedFragmentsResponse{Items: items, Total: res.Total, Page: res.Page, PageSize: res.PageSize})
}

func (h *Handler) listDocuments(c *fiber.Ctx) error {
	topicID, err := uuid.Parse(c.Params("topic_id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid topic id")
	}
	items, err := h.listDocUC.Execute(c.Context(), topicID)
	if err != nil {
		return mapErr(c, err)
	}
	out := make([]DocumentResponse, 0, len(items))
	for _, d := range items {
		out = append(out, toDocResp(d))
	}
	return httpx.OK(c, fiber.Map{"items": out, "total": len(out)})
}

func (h *Handler) createDocument(c *fiber.Ctx) error {
	var b CreateDocumentRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid body")
	}
	tid, err := uuid.Parse(b.TopicID)
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_TOPIC_ID", "invalid topic id")
	}
	d, err := h.createDocUC.Execute(c.Context(), application.CreateDocumentInput{
		TopicID: tid, Title: b.Title, Source: b.Source, License: b.License,
		URLOriginal: b.URLOriginal, Version: b.Version,
	})
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.Created(c, toDocResp(d))
}

func (h *Handler) createFragment(c *fiber.Ctx) error {
	var b CreateFragmentRequest
	if err := c.BodyParser(&b); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid body")
	}
	did, err := uuid.Parse(b.DocumentID)
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_DOCUMENT_ID", "invalid document id")
	}
	tid, err := uuid.Parse(b.TopicID)
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_TOPIC_ID", "invalid topic id")
	}
	f, err := h.createFragUC.Execute(c.Context(), application.CreateFragmentInput{
		DocumentID: did, TopicID: tid, TextChunk: b.TextChunk, Position: b.Position,
		Tokens: b.Tokens, Embedding: b.Embedding, HashChunk: b.HashChunk,
	})
	if err != nil {
		return mapErr(c, err)
	}
	return httpx.Created(c, toFragResp(f))
}

func toFragResp(f *domain.Fragment) FragmentResponse {
	return FragmentResponse{
		ID: f.ID.String(), DocumentID: f.DocumentID.String(), TopicID: f.TopicID.String(),
		TextChunk: f.TextChunk, Position: f.Position, Tokens: f.Tokens,
		Embedding: f.Embedding, HashChunk: f.HashChunk, CreatedAt: f.CreatedAt,
	}
}

func toDocResp(d *domain.KnowledgeDocument) DocumentResponse {
	return DocumentResponse{
		ID: d.ID.String(), TopicID: d.TopicID.String(), Title: d.Title, Source: d.Source,
		License: d.License, URLOriginal: d.URLOriginal, Version: d.Version, CreatedAt: d.CreatedAt,
	}
}

func mapErr(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrFragmentNotFound), errors.Is(err, domain.ErrDocumentNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrInvalidPageSize):
		return httpx.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
