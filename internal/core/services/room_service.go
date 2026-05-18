package services

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/outbound"
	"github.com/chatchomphu1000/go-starter/pkg/apperrors"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

type roomService struct {
	repo  outbound.RoomRepository
	clock outbound.Clock
	ids   outbound.IDGenerator
	log   logger.Logger
}

// NewRoomService creates a new RoomService.
func NewRoomService(
	repo outbound.RoomRepository,
	clock outbound.Clock,
	ids outbound.IDGenerator,
	log logger.Logger,
) inbound.RoomService {
	return &roomService{repo: repo, clock: clock, ids: ids, log: log}
}

// Create adds a new room for an owner.
func (s *roomService) Create(ctx context.Context, in inbound.CreateRoomInput) (*domain.Room, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("roomService.Create: %w", err)
	}

	exists, err := s.repo.ExistsByNumber(ctx, in.OwnerID, in.Number)
	if err != nil {
		return nil, fmt.Errorf("roomService.Create: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("roomService.Create: %w", domain.ErrRoomNumberExists)
	}

	roomType := domain.RoomType(in.Type)
	room, err := domain.NewRoom(
		s.ids.New(), in.OwnerID, in.Number, in.Name,
		roomType, in.Floor, in.SizeSqm, in.RentPrice, in.Deposit,
		in.Amenities, in.Photos, in.Description, s.clock.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("roomService.Create: %w", apperrors.BadRequest(err.Error(), err))
	}

	if err := s.repo.Insert(ctx, room); err != nil {
		return nil, fmt.Errorf("roomService.Create: %w", err)
	}

	s.log.Info("room created", zap.String("room_id", room.ID), zap.String("owner_id", in.OwnerID))
	return room, nil
}

// GetByID retrieves a room by ID.
func (s *roomService) GetByID(ctx context.Context, id string) (*domain.Room, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("roomService.GetByID: %w", err)
	}
	room, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("roomService.GetByID: %w", err)
	}
	return room, nil
}

// List returns filtered rooms with pagination.
func (s *roomService) List(ctx context.Context, f inbound.RoomFilter) ([]*domain.Room, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, fmt.Errorf("roomService.List: %w", err)
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	rooms, total, err := s.repo.FindAll(ctx, f)
	if err != nil {
		return nil, 0, fmt.Errorf("roomService.List: %w", err)
	}
	return rooms, total, nil
}

// Update modifies a room — only the owner may update their room.
func (s *roomService) Update(ctx context.Context, id, ownerID string, in inbound.UpdateRoomInput) (*domain.Room, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("roomService.Update: %w", err)
	}

	room, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("roomService.Update: %w", err)
	}
	if room.OwnerID != ownerID {
		return nil, fmt.Errorf("roomService.Update: %w", domain.ErrUnauthorizedAccess)
	}

	if in.Name != nil {
		room.Name = *in.Name
	}
	if in.Type != nil {
		room.Type = domain.RoomType(*in.Type)
	}
	if in.Floor != nil {
		room.Floor = *in.Floor
	}
	if in.SizeSqm != nil {
		room.SizeSqm = *in.SizeSqm
	}
	if in.RentPrice != nil {
		room.RentPrice = *in.RentPrice
	}
	if in.Deposit != nil {
		room.Deposit = *in.Deposit
	}
	if in.Amenities != nil {
		room.Amenities = in.Amenities
	}
	if in.Photos != nil {
		room.Photos = in.Photos
	}
	if in.Description != nil {
		room.Description = *in.Description
	}
	if in.Status != nil {
		room.Status = domain.RoomStatus(*in.Status)
	}

	room.Touch(s.clock.Now())

	if err := s.repo.Update(ctx, room); err != nil {
		return nil, fmt.Errorf("roomService.Update: %w", err)
	}

	s.log.Info("room updated", zap.String("room_id", id))
	return room, nil
}

// Delete removes a room — only the owner may delete their own rooms.
func (s *roomService) Delete(ctx context.Context, id, ownerID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("roomService.Delete: %w", err)
	}

	room, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("roomService.Delete: %w", err)
	}
	if room.OwnerID != ownerID {
		return fmt.Errorf("roomService.Delete: %w", domain.ErrUnauthorizedAccess)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("roomService.Delete: %w", err)
	}

	s.log.Info("room deleted", zap.String("room_id", id))
	return nil
}
