// Package ports declares the Fragment and KnowledgeDocument repository interfaces.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/knowledge/internal/fragment/domain"
)

type FragmentListFilter struct {
	TopicID  uuid.UUID
	Page     int
	PageSize int
}

type FragmentListResult struct {
	Items    []*domain.Fragment
	Total    int64
	Page     int
	PageSize int
}

type FragmentRepository interface {
	Create(ctx context.Context, f *domain.Fragment) (*domain.Fragment, error)
	ListByTopic(ctx context.Context, f FragmentListFilter) (*FragmentListResult, error)
}

type DocumentRepository interface {
	Create(ctx context.Context, d *domain.KnowledgeDocument) (*domain.KnowledgeDocument, error)
	ListByTopic(ctx context.Context, topicID uuid.UUID) ([]*domain.KnowledgeDocument, error)
}
