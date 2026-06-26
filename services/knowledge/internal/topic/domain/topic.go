// Package domain contains the Topic and UserTopic aggregates.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Topic struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description string
	Icon        string
	SizeMB      int
	Version     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserTopic represents the user's choice (2-5 topics at onboarding).
type UserTopic struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	TopicID      uuid.UUID
	Downloaded   bool
	SelectedAt   time.Time
	DownloadedAt *time.Time
	Version      string
}

var (
	ErrTopicNotFound     = errors.New("topic not found")
	ErrUserTopicNotFound = errors.New("user has not selected this topic")
	ErrTooFewTopics      = errors.New("must select at least 2 topics")
	ErrTooManyTopics     = errors.New("cannot select more than 5 topics")
)

// MinTopics and MaxTopics define how many topics a user can pick at once.
const (
	MinTopics = 2
	MaxTopics = 5
)
