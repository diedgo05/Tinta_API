// Package postgres implements ports.BookRepository using pgx.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tinta/catalog/internal/book/domain"
	"github.com/tinta/catalog/internal/book/ports"
)

type BookRepository struct{ db *pgxpool.Pool }

func NewBookRepository(db *pgxpool.Pool) *BookRepository { return &BookRepository{db: db} }

type row struct {
	ID            uuid.UUID
	GenreID       *uuid.UUID
	Title         string
	Author        string
	ISBN          *string
	Synopsis      *string
	CoverURL      *string
	TotalPages    int32
	License       string
	Language      string
	PublishedYear *int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (r row) toDomain() *domain.Book {
	b := &domain.Book{
		ID:         r.ID,
		GenreID:    r.GenreID,
		Title:      r.Title,
		Author:     r.Author,
		TotalPages: int(r.TotalPages),
		License:    domain.License(r.License),
		Language:   r.Language,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
	if r.ISBN != nil {
		b.ISBN = *r.ISBN
	}
	if r.Synopsis != nil {
		b.Synopsis = *r.Synopsis
	}
	if r.CoverURL != nil {
		b.CoverURL = *r.CoverURL
	}
	if r.PublishedYear != nil {
		v := int(*r.PublishedYear)
		b.PublishedYear = &v
	}
	return b
}

const cols = `id, genre_id, title, author, isbn, synopsis, cover_url, total_pages, license, language, published_year, created_at, updated_at`

func (r *BookRepository) Create(ctx context.Context, b *domain.Book) (*domain.Book, error) {
	const q = `INSERT INTO books (genre_id, title, author, isbn, synopsis, cover_url, total_pages, license, language, published_year)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING ` + cols
	var py *int32
	if b.PublishedYear != nil {
		v := int32(*b.PublishedYear)
		py = &v
	}
	var rr row
	err := r.db.QueryRow(ctx, q,
		b.GenreID, b.Title, b.Author, ns(b.ISBN), ns(b.Synopsis), ns(b.CoverURL),
		b.TotalPages, string(b.License), b.Language, py,
	).Scan(&rr.ID, &rr.GenreID, &rr.Title, &rr.Author, &rr.ISBN, &rr.Synopsis, &rr.CoverURL,
		&rr.TotalPages, &rr.License, &rr.Language, &rr.PublishedYear, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert book: %w", err)
	}
	return rr.toDomain(), nil
}

func (r *BookRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	const q = `SELECT ` + cols + ` FROM books WHERE id=$1 AND deleted_at IS NULL`
	var rr row
	err := r.db.QueryRow(ctx, q, id).Scan(&rr.ID, &rr.GenreID, &rr.Title, &rr.Author, &rr.ISBN, &rr.Synopsis, &rr.CoverURL,
		&rr.TotalPages, &rr.License, &rr.Language, &rr.PublishedYear, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBookNotFound
		}
		return nil, fmt.Errorf("get book: %w", err)
	}
	return rr.toDomain(), nil
}

func (r *BookRepository) List(ctx context.Context, f ports.ListFilter) (*ports.ListResult, error) {
	conds := []string{"deleted_at IS NULL"}
	args := []any{}
	i := 1
	if f.GenreID != nil {
		conds = append(conds, fmt.Sprintf("genre_id=$%d", i))
		args = append(args, *f.GenreID)
		i++
	}
	if f.Search != "" {
		conds = append(conds, fmt.Sprintf("(title ILIKE $%d OR author ILIKE $%d)", i, i))
		args = append(args, "%"+f.Search+"%")
		i++
	}
	where := "WHERE " + strings.Join(conds, " AND ")

	var total int64
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM books "+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count: %w", err)
	}
	offset := (f.Page - 1) * f.PageSize
	args = append(args, f.PageSize, offset)
	q := fmt.Sprintf("SELECT %s FROM books %s ORDER BY title ASC LIMIT $%d OFFSET $%d", cols, where, i, i+1)
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}
	defer rows.Close()
	items := make([]*domain.Book, 0)
	for rows.Next() {
		var rr row
		if err := rows.Scan(&rr.ID, &rr.GenreID, &rr.Title, &rr.Author, &rr.ISBN, &rr.Synopsis, &rr.CoverURL,
			&rr.TotalPages, &rr.License, &rr.Language, &rr.PublishedYear, &rr.CreatedAt, &rr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		items = append(items, rr.toDomain())
	}
	return &ports.ListResult{Items: items, Total: total, Page: f.Page, PageSize: f.PageSize}, nil
}

func (r *BookRepository) Update(ctx context.Context, id uuid.UUID, u ports.BookUpdates) (*domain.Book, error) {
	const q = `UPDATE books SET
		genre_id       = COALESCE($2, genre_id),
		title          = COALESCE($3, title),
		author         = COALESCE($4, author),
		isbn           = COALESCE($5, isbn),
		synopsis       = COALESCE($6, synopsis),
		cover_url      = COALESCE($7, cover_url),
		total_pages    = COALESCE($8, total_pages),
		license        = COALESCE($9, license),
		language       = COALESCE($10, language),
		published_year = COALESCE($11, published_year),
		updated_at     = NOW()
	WHERE id=$1 AND deleted_at IS NULL RETURNING ` + cols

	var lic *string
	if u.License != nil {
		s := string(*u.License)
		lic = &s
	}
	var tp *int32
	if u.TotalPages != nil {
		v := int32(*u.TotalPages)
		tp = &v
	}
	var py *int32
	if u.PublishedYear != nil {
		v := int32(*u.PublishedYear)
		py = &v
	}
	var rr row
	err := r.db.QueryRow(ctx, q, id, u.GenreID, u.Title, u.Author, u.ISBN, u.Synopsis, u.CoverURL,
		tp, lic, u.Language, py).Scan(&rr.ID, &rr.GenreID, &rr.Title, &rr.Author, &rr.ISBN, &rr.Synopsis, &rr.CoverURL,
		&rr.TotalPages, &rr.License, &rr.Language, &rr.PublishedYear, &rr.CreatedAt, &rr.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBookNotFound
		}
		return nil, fmt.Errorf("update: %w", err)
	}
	return rr.toDomain(), nil
}

func (r *BookRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE books SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrBookNotFound
	}
	return nil
}

func ns(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
