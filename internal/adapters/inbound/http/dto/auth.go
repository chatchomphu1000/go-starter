// Package dto contains HTTP request/response data transfer objects.
package dto

// RegisterRequest is the HTTP request body for user registration.
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=80"`
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,min=10,max=128"`
}

// LoginRequest is the HTTP request body for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenRequest is the HTTP request body for refreshing a JWT token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
