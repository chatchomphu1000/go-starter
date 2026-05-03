// Package domain contains the core business entities and sentinel errors.
package domain

import "errors"

// Sentinel errors for the domain layer.
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidRole        = errors.New("invalid role")
	ErrWeakPassword       = errors.New("weak password")
	ErrUserInactive       = errors.New("user inactive")
)
