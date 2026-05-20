package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/internal/mocks"
)

type bookingMocks struct {
	repo     *mocks.MockBookingRepository
	roomRepo *mocks.MockRoomRepository
	clock    *mocks.MockClock
	ids      *mocks.MockIDGenerator
}

func newTestBookingService(t *testing.T) (*bookingService, bookingMocks) {
	t.Helper()
	m := bookingMocks{
		repo:     new(mocks.MockBookingRepository),
		roomRepo: new(mocks.MockRoomRepository),
		clock:    new(mocks.MockClock),
		ids:      new(mocks.MockIDGenerator),
	}
	svc := NewBookingService(m.repo, m.roomRepo, m.clock, m.ids, nopLogger()).(*bookingService)
	return svc, m
}

func validBooking(t *testing.T) *domain.Booking {
	t.Helper()
	b, _ := domain.NewBooking(
		"booking-1", "room-1", "tenant-1", "owner-1",
		fixedNow, fixedNow.AddDate(0, 1, 0),
		5000.0, 10000.0, "notes", fixedNow,
	)
	return b
}

func validCreateBookingInput() inbound.CreateBookingInput {
	return inbound.CreateBookingInput{
		RoomID:    "room-1",
		TenantID:  "tenant-1",
		StartDate: fixedNow,
		EndDate:   fixedNow.AddDate(0, 1, 0),
		Notes:     "notes",
	}
}

// ─── List ────────────────────────────────────────────────────────────────────

func TestBookingService_List(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		filter  inbound.BookingFilter
		setup   func(m bookingMocks)
		wantErr error
	}{
		{
			name:   "success_default_pagination",
			ctx:    context.Background(),
			filter: inbound.BookingFilter{Page: 0, Limit: 0},
			setup: func(m bookingMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.MatchedBy(func(f inbound.BookingFilter) bool {
					return f.Page == 1 && f.Limit == 20
				})).Return([]*domain.Booking{}, int64(0), nil)
			},
		},
		{
			name:   "success_limit_clamped",
			ctx:    context.Background(),
			filter: inbound.BookingFilter{Page: 1, Limit: 200},
			setup: func(m bookingMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.MatchedBy(func(f inbound.BookingFilter) bool {
					return f.Limit == 20
				})).Return([]*domain.Booking{}, int64(0), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			filter:  inbound.BookingFilter{},
			setup:   func(m bookingMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:   "repo_error",
			ctx:    context.Background(),
			filter: inbound.BookingFilter{Page: 1, Limit: 10},
			setup: func(m bookingMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestBookingService(t)
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

// ─── Create ──────────────────────────────────────────────────────────────────

func TestBookingService_Create(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		input   inbound.CreateBookingInput
		setup   func(m bookingMocks)
		wantErr error
	}{
		{
			name:  "success",
			ctx:   context.Background(),
			input: validCreateBookingInput(),
			setup: func(m bookingMocks) {
				r := validRoom(t)
				m.roomRepo.EXPECT().FindByID(mock.Anything, "room-1").Return(r, nil)
				m.repo.EXPECT().HasActiveBooking(mock.Anything, "room-1", mock.Anything, mock.Anything).Return(false, nil)
				m.ids.EXPECT().New().Return("booking-1")
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			input:   validCreateBookingInput(),
			setup:   func(m bookingMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:  "room_not_found",
			ctx:   context.Background(),
			input: validCreateBookingInput(),
			setup: func(m bookingMocks) {
				m.roomRepo.EXPECT().FindByID(mock.Anything, "room-1").Return(nil, domain.ErrRoomNotFound)
			},
			wantErr: domain.ErrRoomNotFound,
		},
		{
			name:  "room_unavailable",
			ctx:   context.Background(),
			input: validCreateBookingInput(),
			setup: func(m bookingMocks) {
				r := validRoom(t)
				r.SetStatus(domain.RoomStatusOccupied)
				m.roomRepo.EXPECT().FindByID(mock.Anything, "room-1").Return(r, nil)
			},
			wantErr: domain.ErrRoomUnavailable,
		},
		{
			name:  "booking_conflict",
			ctx:   context.Background(),
			input: validCreateBookingInput(),
			setup: func(m bookingMocks) {
				r := validRoom(t)
				m.roomRepo.EXPECT().FindByID(mock.Anything, "room-1").Return(r, nil)
				m.repo.EXPECT().HasActiveBooking(mock.Anything, "room-1", mock.Anything, mock.Anything).Return(true, nil)
			},
			wantErr: domain.ErrBookingConflict,
		},
		{
			name: "domain_validation_error_end_before_start",
			ctx:  context.Background(),
			input: inbound.CreateBookingInput{
				RoomID:    "room-1",
				TenantID:  "tenant-1",
				StartDate: fixedNow,
				EndDate:   fixedNow.Add(-time.Hour), // end before start
				Notes:     "notes",
			},
			setup: func(m bookingMocks) {
				r := validRoom(t)
				m.roomRepo.EXPECT().FindByID(mock.Anything, "room-1").Return(r, nil)
				m.repo.EXPECT().HasActiveBooking(mock.Anything, "room-1", mock.Anything, mock.Anything).Return(false, nil)
				m.ids.EXPECT().New().Return("booking-1")
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: errors.New("end_date must be after start_date"),
		},
		{
			name:  "repo_insert_error",
			ctx:   context.Background(),
			input: validCreateBookingInput(),
			setup: func(m bookingMocks) {
				r := validRoom(t)
				m.roomRepo.EXPECT().FindByID(mock.Anything, "room-1").Return(r, nil)
				m.repo.EXPECT().HasActiveBooking(mock.Anything, "room-1", mock.Anything, mock.Anything).Return(false, nil)
				m.ids.EXPECT().New().Return("booking-1")
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestBookingService(t)
			tc.setup(m)

			got, err := svc.Create(tc.ctx, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tc.wantErr, domain.ErrRoomNotFound) {
					assert.True(t, errors.Is(err, domain.ErrRoomNotFound))
				} else if errors.Is(tc.wantErr, domain.ErrRoomUnavailable) {
					assert.True(t, errors.Is(err, domain.ErrRoomUnavailable))
				} else if errors.Is(tc.wantErr, domain.ErrBookingConflict) {
					assert.True(t, errors.Is(err, domain.ErrBookingConflict))
				} else if tc.wantErr == context.Canceled {
					assert.True(t, errors.Is(err, context.Canceled))
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}

			m.repo.AssertExpectations(t)
			m.roomRepo.AssertExpectations(t)
			m.ids.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── GetByID ─────────────────────────────────────────────────────────────────

func TestBookingService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		setup   func(m bookingMocks)
		wantErr error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			id:   "booking-1",
			setup: func(m bookingMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(validBooking(t), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "booking-1",
			setup:   func(m bookingMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "not_found",
			ctx:  context.Background(),
			id:   "missing",
			setup: func(m bookingMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrBookingNotFound)
			},
			wantErr: domain.ErrBookingNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestBookingService(t)
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

// ─── Approve ─────────────────────────────────────────────────────────────────

func TestBookingService_Approve(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		ownerID string
		setup   func(m bookingMocks)
		wantErr error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			id:      "booking-1",
			ownerID: "owner-1",
			setup: func(m bookingMocks) {
				b := validBooking(t)
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(b, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "booking-1",
			ownerID: "owner-1",
			setup:   func(m bookingMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:    "not_found",
			ctx:     context.Background(),
			id:      "missing",
			ownerID: "owner-1",
			setup: func(m bookingMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrBookingNotFound)
			},
			wantErr: domain.ErrBookingNotFound,
		},
		{
			name:    "wrong_owner",
			ctx:     context.Background(),
			id:      "booking-1",
			ownerID: "other-owner",
			setup: func(m bookingMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(validBooking(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
		{
			name:    "invalid_transition_already_approved",
			ctx:     context.Background(),
			id:      "booking-1",
			ownerID: "owner-1",
			setup: func(m bookingMocks) {
				b := validBooking(t)
				b.Status = domain.BookingStatusApproved
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(b, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: domain.ErrInvalidBookingTransition,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestBookingService(t)
			tc.setup(m)

			got, err := svc.Approve(tc.ctx, tc.id, tc.ownerID)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, domain.BookingStatusApproved, got.Status)
			}
			m.repo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── Reject ──────────────────────────────────────────────────────────────────

func TestBookingService_Reject(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		ownerID string
		reason  string
		setup   func(m bookingMocks)
		wantErr error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			id:      "booking-1",
			ownerID: "owner-1",
			reason:  "no vacancy",
			setup: func(m bookingMocks) {
				b := validBooking(t)
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(b, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "booking-1",
			ownerID: "owner-1",
			reason:  "no vacancy",
			setup:   func(m bookingMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:    "wrong_owner",
			ctx:     context.Background(),
			id:      "booking-1",
			ownerID: "other",
			reason:  "no vacancy",
			setup: func(m bookingMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(validBooking(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
		{
			name:    "invalid_transition",
			ctx:     context.Background(),
			id:      "booking-1",
			ownerID: "owner-1",
			reason:  "no vacancy",
			setup: func(m bookingMocks) {
				b := validBooking(t)
				b.Status = domain.BookingStatusRejected
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(b, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: domain.ErrInvalidBookingTransition,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestBookingService(t)
			tc.setup(m)

			got, err := svc.Reject(tc.ctx, tc.id, tc.ownerID, tc.reason)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, domain.BookingStatusRejected, got.Status)
			}
			m.repo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── Cancel ──────────────────────────────────────────────────────────────────

func TestBookingService_Cancel(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		id          string
		requesterID string
		setup       func(m bookingMocks)
		wantErr     error
	}{
		{
			name:        "success_by_tenant",
			ctx:         context.Background(),
			id:          "booking-1",
			requesterID: "tenant-1",
			setup: func(m bookingMocks) {
				b := validBooking(t)
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(b, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:        "context_already_cancelled",
			ctx:         cancelledCtx(),
			id:          "booking-1",
			requesterID: "tenant-1",
			setup:       func(m bookingMocks) {},
			wantErr:     context.Canceled,
		},
		{
			name:        "wrong_actor",
			ctx:         context.Background(),
			id:          "booking-1",
			requesterID: "stranger",
			setup: func(m bookingMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(validBooking(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
		{
			name:        "invalid_transition_active",
			ctx:         context.Background(),
			id:          "booking-1",
			requesterID: "tenant-1",
			setup: func(m bookingMocks) {
				b := validBooking(t)
				b.Status = domain.BookingStatusActive
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(b, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: domain.ErrInvalidBookingTransition,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestBookingService(t)
			tc.setup(m)

			got, err := svc.Cancel(tc.ctx, tc.id, tc.requesterID)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, domain.BookingStatusCancelled, got.Status)
			}
			m.repo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── Activate ────────────────────────────────────────────────────────────────

func TestBookingService_Activate(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		ownerID string
		setup   func(m bookingMocks)
		wantErr error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			id:      "booking-1",
			ownerID: "owner-1",
			setup: func(m bookingMocks) {
				b := validBooking(t)
				b.Status = domain.BookingStatusApproved
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(b, nil)
				m.clock.EXPECT().Now().Return(fixedNow).Maybe()
				m.roomRepo.EXPECT().FindByID(mock.Anything, b.RoomID).Return(validRoom(t), nil).Maybe()
				m.roomRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Maybe()
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "booking-1",
			ownerID: "owner-1",
			setup:   func(m bookingMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:    "wrong_owner",
			ctx:     context.Background(),
			id:      "booking-1",
			ownerID: "other",
			setup: func(m bookingMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(validBooking(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
		{
			name:    "invalid_transition_pending",
			ctx:     context.Background(),
			id:      "booking-1",
			ownerID: "owner-1",
			setup: func(m bookingMocks) {
				b := validBooking(t) // status = pending
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(b, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: domain.ErrInvalidBookingTransition,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestBookingService(t)
			tc.setup(m)

			got, err := svc.Activate(tc.ctx, tc.id, tc.ownerID)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, domain.BookingStatusActive, got.Status)
			}
			m.repo.AssertExpectations(t)
		})
	}
}

// ─── Complete ────────────────────────────────────────────────────────────────

func TestBookingService_Complete(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		ownerID string
		setup   func(m bookingMocks)
		wantErr error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			id:      "booking-1",
			ownerID: "owner-1",
			setup: func(m bookingMocks) {
				b := validBooking(t)
				b.Status = domain.BookingStatusActive
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(b, nil)
				m.clock.EXPECT().Now().Return(fixedNow).Maybe()
				m.roomRepo.EXPECT().FindByID(mock.Anything, b.RoomID).Return(validRoom(t), nil).Maybe()
				m.roomRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Maybe()
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "booking-1",
			ownerID: "owner-1",
			setup:   func(m bookingMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:    "wrong_owner",
			ctx:     context.Background(),
			id:      "booking-1",
			ownerID: "other",
			setup: func(m bookingMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(validBooking(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
		{
			name:    "invalid_transition_pending",
			ctx:     context.Background(),
			id:      "booking-1",
			ownerID: "owner-1",
			setup: func(m bookingMocks) {
				b := validBooking(t) // status = pending, cannot complete
				m.repo.EXPECT().FindByID(mock.Anything, "booking-1").Return(b, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: domain.ErrInvalidBookingTransition,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestBookingService(t)
			tc.setup(m)

			got, err := svc.Complete(tc.ctx, tc.id, tc.ownerID)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, domain.BookingStatusCompleted, got.Status)
			}
			m.repo.AssertExpectations(t)
		})
	}
}
