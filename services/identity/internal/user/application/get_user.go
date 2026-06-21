package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/identity/internal/user/domain"
	"github.com/tinta/identity/internal/user/ports"
)

// GetUserUseCase fetches a user by ID.
type GetUserUseCase struct {
	repo ports.UserRepository
}

// NewGetUserUseCase wires the dependency.
func NewGetUserUseCase(repo ports.UserRepository) *GetUserUseCase {
	return &GetUserUseCase{repo: repo}
}

// Execute returns the user with the given ID or ErrUserNotFound.
func (uc *GetUserUseCase) Execute(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return uc.repo.GetByID(ctx, id)
}
