package mongodb

import (
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

type roomDoc struct {
	ID          string    `bson:"_id"`
	OwnerID     string    `bson:"owner_id"`
	Number      string    `bson:"number"`
	Name        string    `bson:"name"`
	Type        string    `bson:"type"`
	Floor       int       `bson:"floor"`
	SizeSqm     float64   `bson:"size_sqm"`
	RentPrice   float64   `bson:"rent_price"`
	Deposit     float64   `bson:"deposit"`
	Status      string    `bson:"status"`
	Amenities   []string  `bson:"amenities"`
	Photos      []string  `bson:"photos"`
	Description string    `bson:"description"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

func roomFromDomain(r *domain.Room) roomDoc {
	return roomDoc{
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
		Amenities:   r.Amenities,
		Photos:      r.Photos,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func roomToDomain(d roomDoc) *domain.Room {
	return &domain.Room{
		ID:          d.ID,
		OwnerID:     d.OwnerID,
		Number:      d.Number,
		Name:        d.Name,
		Type:        domain.RoomType(d.Type),
		Floor:       d.Floor,
		SizeSqm:     d.SizeSqm,
		RentPrice:   d.RentPrice,
		Deposit:     d.Deposit,
		Status:      domain.RoomStatus(d.Status),
		Amenities:   d.Amenities,
		Photos:      d.Photos,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
