package domain

import (
	"fmt"
	"time"
)

// User is the core domain entity representing a user in the system.
type User struct {
	ID             string
	Name           string
	Email          Email
	HashedPassword string
	Role           Role
	Active         bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewUser creates a new User with the given values and validates it.
// The hashedPassword should already be hashed by the PasswordHasher port in the service layer.
func NewUser(id string, name string, email Email, hashedPassword string, role Role, now time.Time) (*User, error) {
	u := &User{
		ID:             id,
		Name:           name,
		Email:          email,
		HashedPassword: hashedPassword,
		Role:           role,
		Active:         true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := u.Validate(); err != nil {
		return nil, err
	}

	return u, nil
}

// Validate checks the invariants of the user entity.
func (u *User) Validate() error {
	if u.ID == "" {
		return fmt.Errorf("user: id is required")
	}
	if u.Name == "" {
		return fmt.Errorf("user: name is required")
	}
	if u.Email == "" {
		return fmt.Errorf("user: email is required")
	}
	if !u.Role.IsValid() {
		return fmt.Errorf("user: %w: %q", ErrInvalidRole, u.Role)
	}
	return nil
}

// Rename updates the user's name.
func (u *User) Rename(name string) error {
	if name == "" {
		return fmt.Errorf("user: name cannot be empty")
	}
	u.Name = name
	return nil
}

// Deactivate marks the user as inactive.
func (u *User) Deactivate() {
	u.Active = false
}

// Activate marks the user as active.
func (u *User) Activate() {
	u.Active = true
}

// Touch updates the UpdatedAt timestamp.
func (u *User) Touch(now time.Time) {
	u.UpdatedAt = now
}
