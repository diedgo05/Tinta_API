// Package ports declares the Topic and UserTopic repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/knowledge/internal/topic/domain"
)

type TopicRepository interface {
	List(ctx context.Context) ([]*domain.Topic, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Topic, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Topic, error)
}

type UserTopicRepository interface {
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.UserTopic, error)
	ReplaceUserSelection(ctx context.Context, userID uuid.UUID, topicIDs []uuid.UUID) ([]*domain.UserTopic, error)
	MarkDownloaded(ctx context.Context, userID, topicID uuid.UUID) (*domain.UserTopic, error)
	Remove(ctx context.Context, userID, topicID uuid.UUID) error
}
