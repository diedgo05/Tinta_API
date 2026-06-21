// Package http exposes the Club HTTP handlers (Fiber).
package http

import "time"

// CreateClubRequest is the body of POST /api/v1/clubs.
type CreateClubRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	BookID      *string `json:"book_id,omitempty"`
	IsPrivate   bool    `json:"is_private,omitempty"`
}

// UpdateClubRequest is the body of PATCH /api/v1/clubs/{id}.
type UpdateClubRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	BookID      *string `json:"book_id,omitempty"`
	IsPrivate   *bool   `json:"is_private,omitempty"`
}

// ClubResponse is the JSON sent back when returning a club.
type ClubResponse struct {
	ID          string    `json:"id"`
	CreatorID   string    `json:"creator_id"`
	BookID      *string   `json:"book_id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsPrivate   bool      `json:"is_private"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PaginatedClubsResponse is the JSON for GET /api/v1/clubs.
type PaginatedClubsResponse struct {
	Items    []ClubResponse `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}
