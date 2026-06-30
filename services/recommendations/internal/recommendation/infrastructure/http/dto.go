// Package http exposes the Recommendation HTTP handlers (Fiber).
package http

import "time"

// BookInfo is the embedded book metadata returned alongside a recommendation.
type BookInfo struct {
	GoogleVolumeID string   `json:"google_volume_id,omitempty"`
	Title          string   `json:"title"`
	Authors        []string `json:"authors"`
	Thumbnail      string   `json:"thumbnail,omitempty"`
	InfoLink       string   `json:"info_link,omitempty"`
	Description    string   `json:"description,omitempty"`
}

// RecommendationResponse is the JSON sent back when returning a recommendation.
// `Book` is included when the book metadata exists in `recommended_books`.
type RecommendationResponse struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	BookID      string     `json:"book_id"`
	Score       float64    `json:"score"`
	ClusterID   *int       `json:"cluster_id,omitempty"`
	Source      string     `json:"source"`
	Feedback    *string    `json:"feedback,omitempty"`
	FeedbackAt  *time.Time `json:"feedback_at,omitempty"`
	GeneratedAt time.Time  `json:"generated_at"`
	Book        *BookInfo  `json:"book,omitempty"`
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
