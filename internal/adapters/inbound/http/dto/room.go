package dto

import "github.com/chatchomphu1000/go-starter/internal/core/domain"

// CreateRoomRequest is the request body for creating a room.
type CreateRoomRequest struct {
	Number      string   `json:"number" validate:"required,min=1,max=20"`
	Name        string   `json:"name" validate:"required,min=1,max=100"`
	Type        string   `json:"type" validate:"required,oneof=studio 1br 2br suite loft"`
	Floor       int      `json:"floor" validate:"min=0,max=200"`
	SizeSqm     float64  `json:"size_sqm" validate:"min=0"`
	RentPrice   float64  `json:"rent_price" validate:"required,min=0"`
	Deposit     float64  `json:"deposit" validate:"min=0"`
	Amenities   []string `json:"amenities"`
	Photos      []string `json:"photos"`
	Description string   `json:"description" validate:"max=2000"`
}

// UpdateRoomRequest is the request body for updating a room.
type UpdateRoomRequest struct {
	Name        *string  `json:"name" validate:"omitempty,min=1,max=100"`
	Type        *string  `json:"type" validate:"omitempty,oneof=studio 1br 2br suite loft"`
	Floor       *int     `json:"floor" validate:"omitempty,min=0,max=200"`
	SizeSqm     *float64 `json:"size_sqm" validate:"omitempty,min=0"`
	RentPrice   *float64 `json:"rent_price" validate:"omitempty,min=0"`
	Deposit     *float64 `json:"deposit" validate:"omitempty,min=0"`
	Amenities   []string `json:"amenities"`
	Photos      []string `json:"photos"`
	Description *string  `json:"description" validate:"omitempty,max=2000"`
	Status      *string  `json:"status" validate:"omitempty,oneof=available occupied maintenance"`
}

// RoomResponse is the API representation of a room.
type RoomResponse struct {
	ID          string   `json:"id"`
	OwnerID     string   `json:"owner_id"`
	Number      string   `json:"number"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Floor       int      `json:"floor"`
	SizeSqm     float64  `json:"size_sqm"`
	RentPrice   float64  `json:"rent_price"`
	Deposit     float64  `json:"deposit"`
	Status      string   `json:"status"`
	Amenities   []string `json:"amenities"`
	Photos      []string `json:"photos"`
	Description string   `json:"description"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// RoomListResponse wraps a paginated list of rooms.
type RoomListResponse struct {
	Data  []RoomResponse `json:"data"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// ToRoomResponse maps a domain Room to a RoomResponse DTO.
func ToRoomResponse(r *domain.Room) RoomResponse {
	amenities := r.Amenities
	if amenities == nil {
		amenities = []string{}
	}
	photos := r.Photos
	if photos == nil {
		photos = []string{}
	}
	return RoomResponse{
		ID:          r.ID,
		OwnerID:     r.OwnerID,
		Number:      r.Number,
		Name:        r.Name,
		Type:        string(r.Type),
		Floor:       r.Floor,
		SizeSqm:     r.SizeSqm,
		RentPrice:   r.RentPrice,
		Deposit:     r.Deposit,
		Status:      string(r.Status),
		Amenities:   amenities,
		Photos:      photos,
		Description: r.Description,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   r.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToRoomListResponse builds a paginated room list response.
func ToRoomListResponse(rooms []*domain.Room, total int64, page, limit int) RoomListResponse {
	data := make([]RoomResponse, 0, len(rooms))
	for _, r := range rooms {
		data = append(data, ToRoomResponse(r))
	}
	return RoomListResponse{Data: data, Total: total, Page: page, Limit: limit}
}
