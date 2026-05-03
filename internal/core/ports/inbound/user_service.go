// Package inbound defines the inbound port interfaces and DTOs for the application.
package inbound

import (
	"context"
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2 --name=UserService

// RegisterInput holds the data required to register a new user.
type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

// LoginInput holds the data required to authenticate a user.
type LoginInput struct {
	Email    string
	Password string
}

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

// AuthToken represents an authentication token returned on successful login.
type AuthToken struct {
	AccessToken string
	ExpiresAt   time.Time
	TokenType   string
}

// UserService defines the inbound port for user-related business operations.
type UserService interface {
	Register(ctx context.Context, in RegisterInput) (*domain.User, error)
	Login(ctx context.Context, in LoginInput) (*AuthToken, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	List(ctx context.Context, f ListFilter) ([]*domain.User, int64, error)
	Update(ctx context.Context, id string, in UpdateInput) (*domain.User, error)
	Delete(ctx context.Context, id string) error
}
