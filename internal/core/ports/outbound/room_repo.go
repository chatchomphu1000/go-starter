package outbound

import (
	"context"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
)

//go:generate go run github.com/vektra/mockery/v2 --name=RoomRepository

// RoomRepository defines the outbound port for room persistence.
type RoomRepository interface {
	Insert(ctx context.Context, r *domain.Room) error
	FindByID(ctx context.Context, id string) (*domain.Room, error)
	FindAll(ctx context.Context, f inbound.RoomFilter) ([]*domain.Room, int64, error)
	Update(ctx context.Context, r *domain.Room) error
	Delete(ctx context.Context, id string) error
	ExistsByNumber(ctx context.Context, ownerID, number string) (bool, error)
}
