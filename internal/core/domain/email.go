package domain

import (
	"fmt"
	"net/mail"
	"strings"
)

// Email is a value object representing a validated email address.
type Email string

// NewEmail creates a validated Email value object from a raw string.
// It trims whitespace, lowercases, validates RFC 5322 format, and enforces a max length of 254.
func NewEmail(raw string) (Email, error) {
	trimmed := strings.TrimSpace(raw)
	lower := strings.ToLower(trimmed)

	if len(lower) == 0 {
		return "", fmt.Errorf("%w: email is empty", ErrInvalidEmail)
	}

	if len(lower) > 254 {
		return "", fmt.Errorf("%w: email exceeds 254 characters", ErrInvalidEmail)
	}

	addr, err := mail.ParseAddress(lower)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrInvalidEmail, err.Error())
	}

	// Use addr.Address (the parsed mailbox only, stripping any display name).
	return Email(addr.Address), nil
}

// String returns the string representation of the email.
func (e Email) String() string {
	return string(e)
}
