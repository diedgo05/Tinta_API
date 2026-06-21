package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/identity/internal/user/ports"
)

// DeleteUserUseCase implements the ARCO cancellation right.
// Soft-deletes the account; refresh tokens are revoked at the auth layer.
type DeleteUserUseCase struct {
	repo ports.UserRepository
}

// NewDeleteUserUseCase wires the dependency.
func NewDeleteUserUseCase(repo ports.UserRepository) *DeleteUserUseCase {
	return &DeleteUserUseCase{repo: repo}
}

// Execute soft-deletes the user.
func (uc *DeleteUserUseCase) Execute(ctx context.Context, userID uuid.UUID) error {
	return uc.repo.SoftDelete(ctx, userID)
}
