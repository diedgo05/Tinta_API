// Package ports declares the interfaces (ports) used by the application layer.
// Implementations live in the infrastructure layer.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/tinta/identity/internal/user/domain"
)

// UserRepository is the persistence port for the User aggregate.
// The application layer depends on this interface, not on Postgres directly.
type UserRepository interface {
	Create(ctx context.Context, u *domain.User) (*domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	Update(ctx context.Context, id uuid.UUID, updates UserUpdates) (*domain.User, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

// UserUpdates carries optional fields for partial updates.
// nil pointer = "do not change this field".
type UserUpdates struct {
	Name      *string
	AvatarURL *string
	Language  *string
}

// PasswordHasher abstracts password hashing for the application layer.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}
