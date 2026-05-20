package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/internal/mocks"
)

type roomMocks struct {
	repo  *mocks.MockRoomRepository
	clock *mocks.MockClock
	ids   *mocks.MockIDGenerator
}

func newTestRoomService(t *testing.T) (*roomService, roomMocks) {
	t.Helper()
	m := roomMocks{
		repo:  new(mocks.MockRoomRepository),
		clock: new(mocks.MockClock),
		ids:   new(mocks.MockIDGenerator),
	}
	svc := NewRoomService(m.repo, m.clock, m.ids, nopLogger()).(*roomService)
	return svc, m
}

func validRoom(t *testing.T) *domain.Room {
	t.Helper()
	r, _ := domain.NewRoom(
		"room-1", "owner-1", "101", "Room 101",
		domain.RoomTypeStudio, 1, 25.0, 5000.0, 10000.0,
		[]string{"wifi"}, []string{}, "Nice room",
		fixedNow,
	)
	return r
}

func validCreateRoomInput() inbound.CreateRoomInput {
	return inbound.CreateRoomInput{
		OwnerID:     "owner-1",
		Number:      "101",
		Name:        "Room 101",
		Type:        "studio",
		Floor:       1,
		SizeSqm:     25.0,
		RentPrice:   5000.0,
		Deposit:     10000.0,
		Amenities:   []string{"wifi"},
		Photos:      []string{},
		Description: "Nice room",
	}
}

// ─── Create ──────────────────────────────────────────────────────────────────

func TestRoomService_Create(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		input   inbound.CreateRoomInput
		setup   func(m roomMocks)
		wantErr error
	}{
		{
			name:  "success",
			ctx:   context.Background(),
			input: validCreateRoomInput(),
			setup: func(m roomMocks) {
				m.repo.EXPECT().ExistsByNumber(mock.Anything, "owner-1", "101").Return(false, nil)
				m.ids.EXPECT().New().Return("room-1")
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			input:   validCreateRoomInput(),
			setup:   func(m roomMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:  "room_number_already_exists",
			ctx:   context.Background(),
			input: validCreateRoomInput(),
			setup: func(m roomMocks) {
				m.repo.EXPECT().ExistsByNumber(mock.Anything, "owner-1", "101").Return(true, nil)
			},
			wantErr: domain.ErrRoomNumberExists,
		},
		{
			name: "domain_validation_error_negative_price",
			ctx:  context.Background(),
			input: func() inbound.CreateRoomInput {
				in := validCreateRoomInput()
				in.RentPrice = -1.0
				return in
			}(),
			setup: func(m roomMocks) {
				m.repo.EXPECT().ExistsByNumber(mock.Anything, "owner-1", "101").Return(false, nil)
				m.ids.EXPECT().New().Return("room-1")
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: errors.New("rent_price cannot be negative"),
		},
		{
			name:  "repo_insert_error",
			ctx:   context.Background(),
			input: validCreateRoomInput(),
			setup: func(m roomMocks) {
				m.repo.EXPECT().ExistsByNumber(mock.Anything, "owner-1", "101").Return(false, nil)
				m.ids.EXPECT().New().Return("room-1")
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestRoomService(t)
			tc.setup(m)

			got, err := svc.Create(tc.ctx, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tc.wantErr, domain.ErrRoomNumberExists) {
					assert.True(t, errors.Is(err, domain.ErrRoomNumberExists))
				} else if tc.wantErr == context.Canceled {
					assert.True(t, errors.Is(err, context.Canceled))
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, "room-1", got.ID)
			}

			m.repo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
			m.ids.AssertExpectations(t)
		})
	}
}

// ─── GetByID ─────────────────────────────────────────────────────────────────

func TestRoomService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		setup   func(m roomMocks)
		wantErr error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			id:   "room-1",
			setup: func(m roomMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "room-1").Return(validRoom(t), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "room-1",
			setup:   func(m roomMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "not_found",
			ctx:  context.Background(),
			id:   "missing",
			setup: func(m roomMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrRoomNotFound)
			},
			wantErr: domain.ErrRoomNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestRoomService(t)
			tc.setup(m)

			got, err := svc.GetByID(tc.ctx, tc.id)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}
			m.repo.AssertExpectations(t)
		})
	}
}

// ─── List ─────────────────────────────────────────────────────────────────────

func TestRoomService_List(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		filter  inbound.RoomFilter
		setup   func(m roomMocks)
		wantErr error
	}{
		{
			name:   "success_with_clamping",
			ctx:    context.Background(),
			filter: inbound.RoomFilter{Page: 0, Limit: 0},
			setup: func(m roomMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.MatchedBy(func(f inbound.RoomFilter) bool {
					return f.Page == 1 && f.Limit == 20
				})).Return([]*domain.Room{validRoom(t)}, int64(1), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			filter:  inbound.RoomFilter{},
			setup:   func(m roomMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:   "repo_error",
			ctx:    context.Background(),
			filter: inbound.RoomFilter{Page: 1, Limit: 10},
			setup: func(m roomMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestRoomService(t)
			tc.setup(m)

			got, total, err := svc.List(tc.ctx, tc.filter)

			if tc.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				_ = total
			}
			m.repo.AssertExpectations(t)
		})
	}
}

// ─── Update ──────────────────────────────────────────────────────────────────

func TestRoomService_Update(t *testing.T) {
	newName := "Updated Room"

	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		ownerID string
		input   inbound.UpdateRoomInput
		setup   func(m roomMocks)
		wantErr error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			id:      "room-1",
			ownerID: "owner-1",
			input:   inbound.UpdateRoomInput{Name: &newName},
			setup: func(m roomMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "room-1").Return(validRoom(t), nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "room-1",
			ownerID: "owner-1",
			input:   inbound.UpdateRoomInput{Name: &newName},
			setup:   func(m roomMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:    "not_found",
			ctx:     context.Background(),
			id:      "missing",
			ownerID: "owner-1",
			input:   inbound.UpdateRoomInput{Name: &newName},
			setup: func(m roomMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrRoomNotFound)
			},
			wantErr: domain.ErrRoomNotFound,
		},
		{
			name:    "wrong_owner",
			ctx:     context.Background(),
			id:      "room-1",
			ownerID: "other-owner",
			input:   inbound.UpdateRoomInput{Name: &newName},
			setup: func(m roomMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "room-1").Return(validRoom(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestRoomService(t)
			tc.setup(m)

			got, err := svc.Update(tc.ctx, tc.id, tc.ownerID, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}
			m.repo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── Delete ──────────────────────────────────────────────────────────────────

func TestRoomService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		ownerID string
		setup   func(m roomMocks)
		wantErr error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			id:      "room-1",
			ownerID: "owner-1",
			setup: func(m roomMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "room-1").Return(validRoom(t), nil)
				m.repo.EXPECT().Delete(mock.Anything, "room-1").Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "room-1",
			ownerID: "owner-1",
			setup:   func(m roomMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:    "not_found",
			ctx:     context.Background(),
			id:      "missing",
			ownerID: "owner-1",
			setup: func(m roomMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrRoomNotFound)
			},
			wantErr: domain.ErrRoomNotFound,
		},
		{
			name:    "wrong_owner",
			ctx:     context.Background(),
			id:      "room-1",
			ownerID: "other-owner",
			setup: func(m roomMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "room-1").Return(validRoom(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestRoomService(t)
			tc.setup(m)

			err := svc.Delete(tc.ctx, tc.id, tc.ownerID)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
			}
			m.repo.AssertExpectations(t)
		})
	}
}
