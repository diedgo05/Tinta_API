// Package application contains the Topic and UserTopic use cases.
package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/knowledge/internal/topic/domain"
	"github.com/tinta/knowledge/internal/topic/ports"
)

// ---------- Catálogo público de temas ----------

type ListTopicsUseCase struct{ repo ports.TopicRepository }

func NewListTopicsUseCase(r ports.TopicRepository) *ListTopicsUseCase {
	return &ListTopicsUseCase{repo: r}
}

func (uc *ListTopicsUseCase) Execute(ctx context.Context) ([]*domain.Topic, error) {
	return uc.repo.List(ctx)
}

type GetTopicUseCase struct{ repo ports.TopicRepository }

func NewGetTopicUseCase(r ports.TopicRepository) *GetTopicUseCase { return &GetTopicUseCase{repo: r} }

func (uc *GetTopicUseCase) Execute(ctx context.Context, id uuid.UUID) (*domain.Topic, error) {
	return uc.repo.GetByID(ctx, id)
}

// ---------- Selección del usuario ----------

type SelectTopicsUseCase struct {
	topicRepo ports.TopicRepository
	userRepo  ports.UserTopicRepository
}

func NewSelectTopicsUseCase(t ports.TopicRepository, u ports.UserTopicRepository) *SelectTopicsUseCase {
	return &SelectTopicsUseCase{topicRepo: t, userRepo: u}
}

// Execute replaces the user's topic selection with the given list (2-5).
func (uc *SelectTopicsUseCase) Execute(ctx context.Context, userID uuid.UUID, topicIDs []uuid.UUID) ([]*domain.UserTopic, error) {
	if len(topicIDs) < domain.MinTopics {
		return nil, domain.ErrTooFewTopics
	}
	if len(topicIDs) > domain.MaxTopics {
		return nil, domain.ErrTooManyTopics
	}
	// Validate all topics exist
	for _, id := range topicIDs {
		if _, err := uc.topicRepo.GetByID(ctx, id); err != nil {
			return nil, err
		}
	}
	return uc.userRepo.ReplaceUserSelection(ctx, userID, topicIDs)
}

type ListMyTopicsUseCase struct{ repo ports.UserTopicRepository }

func NewListMyTopicsUseCase(r ports.UserTopicRepository) *ListMyTopicsUseCase {
	return &ListMyTopicsUseCase{repo: r}
}

func (uc *ListMyTopicsUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]*domain.UserTopic, error) {
	return uc.repo.ListByUser(ctx, userID)
}

type MarkDownloadedUseCase struct{ repo ports.UserTopicRepository }

func NewMarkDownloadedUseCase(r ports.UserTopicRepository) *MarkDownloadedUseCase {
	return &MarkDownloadedUseCase{repo: r}
}

func (uc *MarkDownloadedUseCase) Execute(ctx context.Context, userID, topicID uuid.UUID) (*domain.UserTopic, error) {
	return uc.repo.MarkDownloaded(ctx, userID, topicID)
}

type RemoveTopicUseCase struct{ repo ports.UserTopicRepository }

func NewRemoveTopicUseCase(r ports.UserTopicRepository) *RemoveTopicUseCase {
	return &RemoveTopicUseCase{repo: r}
}

func (uc *RemoveTopicUseCase) Execute(ctx context.Context, userID, topicID uuid.UUID) error {
	return uc.repo.Remove(ctx, userID, topicID)
}
