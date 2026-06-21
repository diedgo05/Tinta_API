// Package domain contains the Recommendation aggregate.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Source describes which ML pipeline produced the recommendation.
type Source string

const (
	SourceCollaborative Source = "collaborative" // SVD / matrix factorization
	SourceContent       Source = "content"       // similarity over content embeddings
	SourceHybrid        Source = "hybrid"        // weighted combination
	SourceTrending      Source = "trending"      // popularity-based fallback
)

// IsValid reports whether the source is one of the recognized values.
func (s Source) IsValid() bool {
	switch s {
	case SourceCollaborative, SourceContent, SourceHybrid, SourceTrending:
		return true
	}
	return false
}

// Feedback is the user's reaction to a recommendation.
type Feedback string

const (
	FeedbackLike    Feedback = "like"
	FeedbackDislike Feedback = "dislike"
)

// IsValid reports whether the feedback value is recognized.
func (f Feedback) IsValid() bool {
	return f == FeedbackLike || f == FeedbackDislike
}

// Recommendation is the core aggregate of the Recommendations service.
type Recommendation struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	BookID       uuid.UUID
	Score        float64
	ClusterID    *int
	Source       Source
	Feedback     *Feedback
	FeedbackAt   *time.Time
	DismissedAt  *time.Time
	GeneratedAt  time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// IsActive reports whether the recommendation is still shown to the user.
func (r *Recommendation) IsActive() bool {
	return r.DismissedAt == nil
}

// Sentinel errors.
var (
	ErrRecommendationNotFound = errors.New("recommendation not found")
	ErrInvalidFeedback        = errors.New("invalid feedback value")
	ErrNotAuthorized          = errors.New("not authorized")
)
