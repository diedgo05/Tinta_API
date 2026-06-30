package http

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tinta/recommendations/internal/recommendation/application"
	"github.com/tinta/recommendations/internal/recommendation/domain"
	bookApp "github.com/tinta/recommendations/internal/recommendedbook/application"
	"github.com/tinta/shared/httpx"
	"github.com/tinta/shared/middleware"
)

// Handler holds the recommendation use cases.
// `getBooksUC` is optional: if nil, responses don't include book metadata.
type Handler struct {
	listUC       *application.ListRecommendationsUseCase
	feedbackUC   *application.SubmitFeedbackUseCase
	dismissUC    *application.DismissRecommendationUseCase
	regenerateUC *application.RegenerateRecommendationsUseCase
	getBooksUC   *bookApp.GetRecommendedBooksUseCase
}

// NewHandler constructs the recommendation HTTP handler.
func NewHandler(
	list *application.ListRecommendationsUseCase,
	feedback *application.SubmitFeedbackUseCase,
	dismiss *application.DismissRecommendationUseCase,
	regenerate *application.RegenerateRecommendationsUseCase,
	getBooks *bookApp.GetRecommendedBooksUseCase,
) *Handler {
	return &Handler{
		listUC:       list,
		feedbackUC:   feedback,
		dismissUC:    dismiss,
		regenerateUC: regenerate,
		getBooksUC:   getBooks,
	}
}

// Register adds the recommendation routes; all of them require auth.
func (h *Handler) Register(router fiber.Router, authMiddleware fiber.Handler) {
	r := router.Group("/recommendations", authMiddleware)
	r.Get("/", h.list)
	r.Post("/regenerate", h.regenerate)
	r.Post("/:id/feedback", h.feedback)
	r.Delete("/:id", h.dismiss)
}

// list handles GET /recommendations.
func (h *Handler) list(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return httpx.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	result, err := h.listUC.Execute(c.Context(), userID, page, pageSize)
	if err != nil {
		return mapRecoError(c, err)
	}

	// Batch-fetch book metadata in a single query (enrichment).
	bookIDs := make([]uuid.UUID, 0, len(result.Items))
	for _, r := range result.Items {
		bookIDs = append(bookIDs, r.BookID)
	}
	books, err := h.getBooksUC.Execute(c.Context(), bookIDs)
	if err != nil {
		// Enrichment failure is non-fatal: log internally and return without `book`.
		books = nil
	}

	items := make([]RecommendationResponse, 0, len(result.Items))
	for _, r := range result.Items {
		resp := toResponse(r)
		if books != nil {
			if b, ok := books[r.BookID]; ok {
				resp.Book = &BookInfo{
					GoogleVolumeID: b.GoogleVolumeID,
					Title:          b.Title,
					Authors:        b.Authors,
					Thumbnail:      b.Thumbnail,
					InfoLink:       b.InfoLink,
					Description:    b.Description,
				}
			}
		}
		items = append(items, resp)
	}
	return httpx.OK(c, PaginatedRecommendationsResponse{
		Items:    items,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// feedback handles POST /recommendations/{id}/feedback.
func (h *Handler) feedback(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return httpx.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid recommendation id")
	}

	var body SubmitFeedbackRequest
	if err := c.BodyParser(&body); err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}

	updated, err := h.feedbackUC.Execute(c.Context(), application.SubmitFeedbackInput{
		RecommendationID: id,
		RequesterID:      userID,
		Feedback:         domain.Feedback(body.Feedback),
	})
	if err != nil {
		return mapRecoError(c, err)
	}
	return httpx.OK(c, toResponse(updated))
}

// dismiss handles DELETE /recommendations/{id}.
func (h *Handler) dismiss(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return httpx.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid recommendation id")
	}
	if err := h.dismissUC.Execute(c.Context(), id, userID); err != nil {
		return mapRecoError(c, err)
	}
	return httpx.NoContent(c)
}

// regenerate handles POST /recommendations/regenerate.
func (h *Handler) regenerate(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return httpx.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
	}
	if err := h.regenerateUC.Execute(c.Context(), userID); err != nil {
		return mapRecoError(c, err)
	}
	return c.Status(fiber.StatusAccepted).JSON(RegenerateResponse{
		Message: "regeneration job queued",
	})
}

func toResponse(r *domain.Recommendation) RecommendationResponse {
	var feedback *string
	if r.Feedback != nil {
		s := string(*r.Feedback)
		feedback = &s
	}
	return RecommendationResponse{
		ID:          r.ID.String(),
		UserID:      r.UserID.String(),
		BookID:      r.BookID.String(),
		Score:       r.Score,
		ClusterID:   r.ClusterID,
		Source:      string(r.Source),
		Feedback:    feedback,
		FeedbackAt:  r.FeedbackAt,
		GeneratedAt: r.GeneratedAt,
	}
}

func mapRecoError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrRecommendationNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "RECOMMENDATION_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrNotAuthorized):
		return httpx.Error(c, fiber.StatusForbidden, "FORBIDDEN", err.Error())
	case errors.Is(err, domain.ErrInvalidFeedback):
		return httpx.Error(c, fiber.StatusBadRequest, "INVALID_FEEDBACK", err.Error())
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}
