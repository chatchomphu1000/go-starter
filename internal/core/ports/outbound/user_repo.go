// Package outbound defines the outbound port interfaces for external dependencies.
package outbound

import (
	"context"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
)

//go:generate go run github.com/vektra/mockery/v2 --name=UserRepository

// UserRepository defines the data access operations for users.
type UserRepository interface {
	Insert(ctx context.Context, u *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error)
	FindAll(ctx context.Context, f inbound.ListFilter) ([]*domain.User, int64, error)
	Update(ctx context.Context, u *domain.User) error
	Delete(ctx context.Context, id string) error
	ExistsByEmail(ctx context.Context, email domain.Email) (bool, error)
}
