package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/identity/internal/user/domain"
	"github.com/tinta/identity/internal/user/ports"
)

// UpdateUserInput carries optional fields. nil = "do not update".
type UpdateUserInput struct {
	Name      *string
	AvatarURL *string
	Language  *string
}

// UpdateUserUseCase updates the authenticated user's profile.
type UpdateUserUseCase struct {
	repo ports.UserRepository
}

// NewUpdateUserUseCase wires the dependency.
func NewUpdateUserUseCase(repo ports.UserRepository) *UpdateUserUseCase {
	return &UpdateUserUseCase{repo: repo}
}

// Execute applies a partial update to the user.
func (uc *UpdateUserUseCase) Execute(ctx context.Context, userID uuid.UUID, in UpdateUserInput) (*domain.User, error) {
	if in.Name != nil {
		if err := domain.ValidateName(*in.Name); err != nil {
			return nil, err
		}
	}
	if in.Language != nil {
		if err := domain.ValidateLanguage(*in.Language); err != nil {
			return nil, err
		}
	}

	return uc.repo.Update(ctx, userID, ports.UserUpdates{
		Name:      in.Name,
		AvatarURL: in.AvatarURL,
		Language:  in.Language,
	})
}
