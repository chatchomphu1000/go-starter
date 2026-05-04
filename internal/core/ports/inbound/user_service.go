// Package inbound defines the inbound port interfaces and DTOs for the application.
package inbound

import (
	"context"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2 --name=UserService

// UpdateInput holds the data for updating a user. Nil fields are not changed.
type UpdateInput struct {
	Name *string
}

// ListFilter defines pagination, sorting, and filtering options for listing users.
type ListFilter struct {
	Role     *domain.Role
	Active   *bool
	Search   string // name/email substring
	Page     int    // 1-based
	Limit    int    // clamped [1,100]
	SortBy   string // "created_at" | "name" | "email"
	SortDesc bool
}

// UserService defines the inbound port for user-related business operations.
type UserService interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	List(ctx context.Context, f ListFilter) ([]*domain.User, int64, error)
	Update(ctx context.Context, id string, in UpdateInput) (*domain.User, error)
	Delete(ctx context.Context, id string) error
}
