// Package application contains Fragment use cases.
package application

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/tinta/knowledge/internal/fragment/domain"
	"github.com/tinta/knowledge/internal/fragment/ports"
)

const (
	defaultPageSize = 100
	maxPageSize     = 500
)

// ---------- Fragments ----------

type ListFragmentsUseCase struct{ repo ports.FragmentRepository }

func NewListFragmentsUseCase(r ports.FragmentRepository) *ListFragmentsUseCase {
	return &ListFragmentsUseCase{repo: r}
}

func (uc *ListFragmentsUseCase) Execute(ctx context.Context, f ports.FragmentListFilter) (*ports.FragmentListResult, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = defaultPageSize
	}
	if f.PageSize > maxPageSize {
		return nil, domain.ErrInvalidPageSize
	}
	return uc.repo.ListByTopic(ctx, f)
}

type CreateFragmentInput struct {
	DocumentID uuid.UUID
	TopicID    uuid.UUID
	TextChunk  string
	Position   int
	Tokens     int
	Embedding  []float32
	HashChunk  string
}

type CreateFragmentUseCase struct{ repo ports.FragmentRepository }

func NewCreateFragmentUseCase(r ports.FragmentRepository) *CreateFragmentUseCase {
	return &CreateFragmentUseCase{repo: r}
}

func (uc *CreateFragmentUseCase) Execute(ctx context.Context, in CreateFragmentInput) (*domain.Fragment, error) {
	emb, err := json.Marshal(in.Embedding)
	if err != nil {
		return nil, err
	}
	return uc.repo.Create(ctx, &domain.Fragment{
		DocumentID: in.DocumentID, TopicID: in.TopicID, TextChunk: in.TextChunk,
		Position: in.Position, Tokens: in.Tokens, Embedding: emb, HashChunk: in.HashChunk,
	})
}

// ---------- Documents ----------

type ListDocumentsUseCase struct{ repo ports.DocumentRepository }

func NewListDocumentsUseCase(r ports.DocumentRepository) *ListDocumentsUseCase {
	return &ListDocumentsUseCase{repo: r}
}

func (uc *ListDocumentsUseCase) Execute(ctx context.Context, topicID uuid.UUID) ([]*domain.KnowledgeDocument, error) {
	return uc.repo.ListByTopic(ctx, topicID)
}

type CreateDocumentInput struct {
	TopicID     uuid.UUID
	Title       string
	Source      string
	License     string
	URLOriginal string
	Version     string
}

type CreateDocumentUseCase struct{ repo ports.DocumentRepository }

func NewCreateDocumentUseCase(r ports.DocumentRepository) *CreateDocumentUseCase {
	return &CreateDocumentUseCase{repo: r}
}

func (uc *CreateDocumentUseCase) Execute(ctx context.Context, in CreateDocumentInput) (*domain.KnowledgeDocument, error) {
	if in.Version == "" {
		in.Version = "v1"
	}
	if in.License == "" {
		in.License = "cc-by"
	}
	return uc.repo.Create(ctx, &domain.KnowledgeDocument{
		TopicID: in.TopicID, Title: in.Title, Source: in.Source, License: in.License,
		URLOriginal: in.URLOriginal, Version: in.Version,
	})
}
