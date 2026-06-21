// Package application contains the use cases of the User module.
// Each use case is a small struct with a single Execute method.
package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/tinta/identity/internal/user/domain"
	"github.com/tinta/identity/internal/user/ports"
)

// CreateUserInput is the input DTO for the CreateUser use case.
type CreateUserInput struct {
	Email    string
	Password string
	Name     string
	Language string
}

// CreateUserUseCase handles user registration.
type CreateUserUseCase struct {
	repo   ports.UserRepository
	hasher ports.PasswordHasher
}

// NewCreateUserUseCase wires up the dependencies.
func NewCreateUserUseCase(repo ports.UserRepository, hasher ports.PasswordHasher) *CreateUserUseCase {
	return &CreateUserUseCase{repo: repo, hasher: hasher}
}

// Execute registers a new user with role 'user' and email unverified.
func (uc *CreateUserUseCase) Execute(ctx context.Context, in CreateUserInput) (*domain.User, error) {
	// 1. Validate inputs (pure domain logic, no DB call yet)
	if err := domain.ValidateEmail(in.Email); err != nil {
		return nil, err
	}
	if err := domain.ValidatePassword(in.Password); err != nil {
		return nil, err
	}
	if err := domain.ValidateName(in.Name); err != nil {
		return nil, err
	}
	if err := domain.ValidateLanguage(in.Language); err != nil {
		return nil, err
	}

	email := domain.NormalizeEmail(in.Email)

	// 2. Check uniqueness
	exists, err := uc.repo.EmailExists(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("check email exists: %w", err)
	}
	if exists {
		return nil, domain.ErrEmailAlreadyExists
	}

	// 3. Hash the password
	hash, err := uc.hasher.Hash(in.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// 4. Build the entity
	lang := in.Language
	if lang == "" {
		lang = "es"
	}
	user := &domain.User{
		Email:         email,
		PasswordHash:  hash,
		Name:          in.Name,
		Role:          domain.RoleUser,
		EmailVerified: false,
		Language:      lang,
	}

	// 5. Persist
	created, err := uc.repo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			return nil, err
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return created, nil
}
