// Package crypto provides a bcrypt implementation of the PasswordHasher port.
package crypto

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

// BcryptHasher implements outbound.PasswordHasher using bcrypt.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new BcryptHasher with the given cost.
// Rejects cost < 10 for security.
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < 10 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

// Hash hashes a plaintext password using bcrypt.
func (h *BcryptHasher) Hash(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", fmt.Errorf("bcryptHasher.Hash: %w", err)
	}
	return string(bytes), nil
}

// Verify compares a hashed password with a plaintext password.
// Returns domain.ErrInvalidCredentials on mismatch.
func (h *BcryptHasher) Verify(hashed, plain string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return domain.ErrInvalidCredentials
		}
		return fmt.Errorf("bcryptHasher.Verify: %w", err)
	}
	return nil
}
