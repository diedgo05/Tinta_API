// Package postgres implements Fragment and Document repositories.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tinta/knowledge/internal/fragment/domain"
	"github.com/tinta/knowledge/internal/fragment/ports"
)

// ---------- FragmentRepository ----------

type FragmentRepository struct{ db *pgxpool.Pool }

func NewFragmentRepository(db *pgxpool.Pool) *FragmentRepository { return &FragmentRepository{db: db} }

type fragRow struct {
	ID         uuid.UUID
	DocumentID uuid.UUID
	TopicID    uuid.UUID
	TextChunk  string
	Position   int32
	Tokens     int32
	Embedding  []byte
	HashChunk  string
	CreatedAt  time.Time
}

func (r fragRow) toDomain() *domain.Fragment {
	return &domain.Fragment{
		ID: r.ID, DocumentID: r.DocumentID, TopicID: r.TopicID, TextChunk: r.TextChunk,
		Position: int(r.Position), Tokens: int(r.Tokens),
		Embedding: json.RawMessage(r.Embedding), HashChunk: r.HashChunk, CreatedAt: r.CreatedAt,
	}
}

const fragCols = `id, document_id, topic_id, text_chunk, position, tokens, embedding, hash_chunk, created_at`

func (r *FragmentRepository) Create(ctx context.Context, f *domain.Fragment) (*domain.Fragment, error) {
	const q = `INSERT INTO rag_fragments (document_id, topic_id, text_chunk, position, tokens, embedding, hash_chunk)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING ` + fragCols
	var rr fragRow
	err := r.db.QueryRow(ctx, q, f.DocumentID, f.TopicID, f.TextChunk, f.Position, f.Tokens, []byte(f.Embedding), f.HashChunk).Scan(
		&rr.ID, &rr.DocumentID, &rr.TopicID, &rr.TextChunk, &rr.Position, &rr.Tokens, &rr.Embedding, &rr.HashChunk, &rr.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert fragment: %w", err)
	}
	return rr.toDomain(), nil
}

func (r *FragmentRepository) ListByTopic(ctx context.Context, f ports.FragmentListFilter) (*ports.FragmentListResult, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM rag_fragments WHERE topic_id=$1`, f.TopicID).Scan(&total); err != nil {
		return nil, err
	}
	offset := (f.Page - 1) * f.PageSize
	rows, err := r.db.Query(ctx, `SELECT `+fragCols+` FROM rag_fragments
		WHERE topic_id=$1 ORDER BY position ASC LIMIT $2 OFFSET $3`, f.TopicID, f.PageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.Fragment, 0)
	for rows.Next() {
		var rr fragRow
		if err := rows.Scan(&rr.ID, &rr.DocumentID, &rr.TopicID, &rr.TextChunk, &rr.Position, &rr.Tokens,
			&rr.Embedding, &rr.HashChunk, &rr.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rr.toDomain())
	}
	return &ports.FragmentListResult{Items: out, Total: total, Page: f.Page, PageSize: f.PageSize}, nil
}

// ---------- DocumentRepository ----------

type DocumentRepository struct{ db *pgxpool.Pool }

func NewDocumentRepository(db *pgxpool.Pool) *DocumentRepository { return &DocumentRepository{db: db} }

type docRow struct {
	ID          uuid.UUID
	TopicID     uuid.UUID
	Title       string
	Source      string
	License     string
	URLOriginal *string
	Version     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (r docRow) toDomain() *domain.KnowledgeDocument {
	d := &domain.KnowledgeDocument{
		ID: r.ID, TopicID: r.TopicID, Title: r.Title, Source: r.Source,
		License: r.License, Version: r.Version, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
	if r.URLOriginal != nil {
		d.URLOriginal = *r.URLOriginal
	}
	return d
}

const docCols = `id, topic_id, title, source, license, url_original, version, created_at, updated_at`

func (r *DocumentRepository) Create(ctx context.Context, d *domain.KnowledgeDocument) (*domain.KnowledgeDocument, error) {
	const q = `INSERT INTO knowledge_documents (topic_id, title, source, license, url_original, version)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING ` + docCols
	var url *string
	if d.URLOriginal != "" {
		url = &d.URLOriginal
	}
	var rr docRow
	err := r.db.QueryRow(ctx, q, d.TopicID, d.Title, d.Source, d.License, url, d.Version).Scan(
		&rr.ID, &rr.TopicID, &rr.Title, &rr.Source, &rr.License, &rr.URLOriginal, &rr.Version, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert document: %w", err)
	}
	return rr.toDomain(), nil
}

func (r *DocumentRepository) ListByTopic(ctx context.Context, topicID uuid.UUID) ([]*domain.KnowledgeDocument, error) {
	rows, err := r.db.Query(ctx, `SELECT `+docCols+` FROM knowledge_documents WHERE topic_id=$1 ORDER BY created_at DESC`, topicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*domain.KnowledgeDocument{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	out := make([]*domain.KnowledgeDocument, 0)
	for rows.Next() {
		var rr docRow
		if err := rows.Scan(&rr.ID, &rr.TopicID, &rr.Title, &rr.Source, &rr.License, &rr.URLOriginal,
			&rr.Version, &rr.CreatedAt, &rr.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, rr.toDomain())
	}
	return out, nil
}
