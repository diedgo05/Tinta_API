// Package http exposes the User HTTP handlers (Fiber).
package http

import "time"

// RegisterRequest is the body of POST /api/v1/users.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Language string `json:"language,omitempty"`
}

// UpdateUserRequest is the body of PATCH /api/v1/users/me.
// nil fields are not updated.
type UpdateUserRequest struct {
	Name      *string `json:"name,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Language  *string `json:"language,omitempty"`
}

// UserResponse is the JSON sent back when returning a user.
type UserResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	Role          string    `json:"role"`
	EmailVerified bool      `json:"email_verified"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	Language      string    `json:"language"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PublicUserResponse is a redacted version for /users/{id} (other users).
// Hides email and dates of joining.
type PublicUserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}
