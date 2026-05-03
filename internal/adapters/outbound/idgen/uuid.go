// Package idgen provides a UUID v7 implementation of the IDGenerator port.
package idgen

import "github.com/google/uuid"

// UUIDGenerator implements outbound.IDGenerator using UUID v7.
type UUIDGenerator struct{}

// NewUUIDGenerator creates a new UUIDGenerator.
func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

// New generates a new UUID v7 string. Falls back to v4 if v7 is unavailable.
func (g *UUIDGenerator) New() string {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New().String()
	}
	return id.String()
}
