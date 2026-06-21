// Package http exposes the Recommendation HTTP handlers (Fiber).
package http

import "time"

// RecommendationResponse is the JSON sent back when returning a recommendation.
type RecommendationResponse struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	BookID       string    `json:"book_id"`
	Score        float64   `json:"score"`
	ClusterID    *int      `json:"cluster_id,omitempty"`
	Source       string    `json:"source"`
	Feedback     *string   `json:"feedback,omitempty"`
	FeedbackAt   *time.Time `json:"feedback_at,omitempty"`
	GeneratedAt  time.Time `json:"generated_at"`
}

// PaginatedRecommendationsResponse is the JSON for GET /api/v1/recommendations.
type PaginatedRecommendationsResponse struct {
	Items    []RecommendationResponse `json:"items"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

// SubmitFeedbackRequest is the body of POST /api/v1/recommendations/{id}/feedback.
type SubmitFeedbackRequest struct {
	Feedback string `json:"feedback"` // "like" or "dislike"
}

// RegenerateResponse confirms that a regeneration job was queued.
type RegenerateResponse struct {
	Message string `json:"message"`
}
