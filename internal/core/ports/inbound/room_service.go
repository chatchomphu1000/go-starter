package inbound

import (
	"context"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2 --name=RoomService

// CreateRoomInput holds data required to create a room.
type CreateRoomInput struct {
	OwnerID     string
	Number      string
	Name        string
	Type        string
	Floor       int
	SizeSqm     float64
	RentPrice   float64
	Deposit     float64
	Amenities   []string
	Photos      []string
	Description string
}

// UpdateRoomInput holds updatable room fields (nil = no change).
type UpdateRoomInput struct {
	Name        *string
	Type        *string
	Floor       *int
	SizeSqm     *float64
	RentPrice   *float64
	Deposit     *float64
	Amenities   []string
	Photos      []string
	Description *string
	Status      *string
}

// RoomFilter defines query parameters for listing rooms.
type RoomFilter struct {
	OwnerID   string
	Status    *domain.RoomStatus
	Type      *domain.RoomType
	MinPrice  *float64
	MaxPrice  *float64
	Search    string
	Page      int
	Limit     int
	SortBy    string
	SortDesc  bool
}

// RoomService defines the inbound port for room management.
type RoomService interface {
	Create(ctx context.Context, in CreateRoomInput) (*domain.Room, error)
	GetByID(ctx context.Context, id string) (*domain.Room, error)
	List(ctx context.Context, f RoomFilter) ([]*domain.Room, int64, error)
	Update(ctx context.Context, id string, ownerID string, in UpdateRoomInput) (*domain.Room, error)
	Delete(ctx context.Context, id string, ownerID string) error
}
