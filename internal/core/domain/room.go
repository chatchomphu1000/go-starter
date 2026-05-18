package domain

import (
	"fmt"
	"time"
)

// RoomType categorizes the type of a room.
type RoomType string

// Room type constants.
const (
	RoomTypeStudio RoomType = "studio"
	RoomType1BR    RoomType = "1br"
	RoomType2BR    RoomType = "2br"
	RoomTypeSuite  RoomType = "suite"
	RoomTypeLoft   RoomType = "loft"
)

// RoomStatus represents the current availability state of a room.
type RoomStatus string

// Room status constants.
const (
	RoomStatusAvailable   RoomStatus = "available"
	RoomStatusOccupied    RoomStatus = "occupied"
	RoomStatusMaintenance RoomStatus = "maintenance"
)

// Room is the core domain entity for a rentable unit in a dormitory.
type Room struct {
	ID          string
	OwnerID     string
	Number      string // e.g. "101A"
	Name        string
	Type        RoomType
	Floor       int
	SizeSqm     float64
	RentPrice   float64
	Deposit     float64
	Status      RoomStatus
	Amenities   []string
	Photos      []string // signed URLs or paths
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewRoom creates a validated Room entity.
func NewRoom(
	id, ownerID, number, name string,
	roomType RoomType,
	floor int,
	sizeSqm, rentPrice, deposit float64,
	amenities, photos []string,
	description string,
	now time.Time,
) (*Room, error) {
	r := &Room{
		ID:          id,
		OwnerID:     ownerID,
		Number:      number,
		Name:        name,
		Type:        roomType,
		Floor:       floor,
		SizeSqm:     sizeSqm,
		RentPrice:   rentPrice,
		Deposit:     deposit,
		Status:      RoomStatusAvailable,
		Amenities:   amenities,
		Photos:      photos,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	return r, nil
}

// Validate enforces room invariants.
func (r *Room) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("room: id is required")
	}
	if r.OwnerID == "" {
		return fmt.Errorf("room: owner_id is required")
	}
	if r.Number == "" {
		return fmt.Errorf("room: number is required")
	}
	if r.RentPrice < 0 {
		return fmt.Errorf("room: rent_price cannot be negative")
	}
	if r.Deposit < 0 {
		return fmt.Errorf("room: deposit cannot be negative")
	}
	return nil
}

// Touch updates the UpdatedAt timestamp.
func (r *Room) Touch(now time.Time) { r.UpdatedAt = now }

// SetStatus changes the room's availability status.
func (r *Room) SetStatus(s RoomStatus) { r.Status = s }

// IsAvailable returns true if the room can be booked.
func (r *Room) IsAvailable() bool { return r.Status == RoomStatusAvailable }
