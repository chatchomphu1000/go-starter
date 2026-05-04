// Package inbound defines the inbound port interfaces and DTOs for the application.
package inbound

import (
	"context"
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2 --name=AuthService

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

// AuthToken represents an authentication token pair returned on successful login or refresh.
type AuthToken struct {
	AccessToken          string
	ExpiresAt            time.Time
	TokenType            string
	RefreshToken         string
	RefreshExpiresAt     time.Time
}

// RefreshTokenInput holds the data required to refresh an authentication token.
type RefreshTokenInput struct {
	RefreshToken string
}

// AuthService defines the inbound port for authentication-related business operations.
type AuthService interface {
	Register(ctx context.Context, in RegisterInput) (*domain.User, error)
	Login(ctx context.Context, in LoginInput) (*AuthToken, error)
	RefreshToken(ctx context.Context, in RefreshTokenInput) (*AuthToken, error)	
}