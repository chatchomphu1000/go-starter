// Package dto contains HTTP request/response data transfer objects.
package dto

import (
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

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

// UpdateRequest is the HTTP request body for updating a user.
type UpdateRequest struct {
	Name *string `json:"name" validate:"omitempty,min=2,max=80"`
}

// UserResponse is the HTTP response body for a single user.
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListResponse is the HTTP response body for a list of users.
type ListResponse struct {
	Data  []UserResponse `json:"data"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// LoginResponse is the HTTP response body for a successful login.
type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// ErrorResponse is the HTTP response body returned on errors.
type ErrorResponse struct {
	Code      string   `json:"code"`
	Message   string   `json:"message"`
	Details   []string `json:"details,omitempty"`
	RequestID string   `json:"request_id,omitempty"`
}

// ToUserResponse maps a domain User to a UserResponse DTO.
func ToUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email.String(),
		Role:      u.Role.String(),
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ToUserListResponse maps a slice of domain Users to a ListResponse DTO.
func ToUserListResponse(users []*domain.User, total int64, page, limit int) ListResponse {
	data := make([]UserResponse, 0, len(users))
	for _, u := range users {
		data = append(data, ToUserResponse(u))
	}
	return ListResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	}
}
