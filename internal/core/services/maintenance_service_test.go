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

type maintenanceMocks struct {
	repo  *mocks.MockMaintenanceRepository
	clock *mocks.MockClock
	ids   *mocks.MockIDGenerator
}

func newTestMaintenanceService(t *testing.T) (*maintenanceService, maintenanceMocks) {
	t.Helper()
	m := maintenanceMocks{
		repo:  new(mocks.MockMaintenanceRepository),
		clock: new(mocks.MockClock),
		ids:   new(mocks.MockIDGenerator),
	}
	svc := NewMaintenanceService(m.repo, m.clock, m.ids, nopLogger()).(*maintenanceService)
	return svc, m
}

func validTicket(t *testing.T) *domain.MaintenanceTicket {
	t.Helper()
	ticket, _ := domain.NewMaintenanceTicket(
		"ticket-1", "room-1", "tenant-1", "owner-1",
		"Broken AC", "AC not cooling", "ac",
		domain.PriorityHigh, []string{}, fixedNow,
	)
	return ticket
}

func validCreateTicketInput() inbound.CreateTicketInput {
	return inbound.CreateTicketInput{
		RoomID:      "room-1",
		TenantID:    "tenant-1",
		OwnerID:     "owner-1",
		Title:       "Broken AC",
		Description: "AC not cooling",
		Category:    "ac",
		Priority:    "high",
		Photos:      []string{},
	}
}

// ─── List ────────────────────────────────────────────────────────────────────

func TestMaintenanceService_List(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		filter  inbound.TicketFilter
		setup   func(m maintenanceMocks)
		wantErr error
	}{
		{
			name:   "success_default_pagination",
			ctx:    context.Background(),
			filter: inbound.TicketFilter{Page: 0, Limit: 0},
			setup: func(m maintenanceMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.MatchedBy(func(f inbound.TicketFilter) bool {
					return f.Page == 1 && f.Limit == 20
				})).Return([]*domain.MaintenanceTicket{}, int64(0), nil)
			},
		},
		{
			name:   "success_limit_clamped",
			ctx:    context.Background(),
			filter: inbound.TicketFilter{Page: 1, Limit: 500},
			setup: func(m maintenanceMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.MatchedBy(func(f inbound.TicketFilter) bool {
					return f.Limit == 20
				})).Return([]*domain.MaintenanceTicket{}, int64(0), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			filter:  inbound.TicketFilter{},
			setup:   func(_ maintenanceMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:   "repo_error",
			ctx:    context.Background(),
			filter: inbound.TicketFilter{Page: 1, Limit: 10},
			setup: func(m maintenanceMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestMaintenanceService(t)
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

func TestMaintenanceService_Create(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		input   inbound.CreateTicketInput
		setup   func(m maintenanceMocks)
		wantErr error
	}{
		{
			name:  "success",
			ctx:   context.Background(),
			input: validCreateTicketInput(),
			setup: func(m maintenanceMocks) {
				m.ids.EXPECT().New().Return("ticket-1")
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			input:   validCreateTicketInput(),
			setup:   func(_ maintenanceMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "domain_validation_empty_title",
			ctx:  context.Background(),
			input: inbound.CreateTicketInput{
				RoomID:   "room-1",
				TenantID: "tenant-1",
				OwnerID:  "owner-1",
				Title:    "", // empty
				Category: "ac",
				Priority: "high",
			},
			setup: func(m maintenanceMocks) {
				m.ids.EXPECT().New().Return("ticket-1")
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: errors.New("title is required"),
		},
		{
			name:  "repo_insert_error",
			ctx:   context.Background(),
			input: validCreateTicketInput(),
			setup: func(m maintenanceMocks) {
				m.ids.EXPECT().New().Return("ticket-1")
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestMaintenanceService(t)
			tc.setup(m)

			got, err := svc.Create(tc.ctx, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, "ticket-1", got.ID)
				assert.Equal(t, domain.TicketStatusOpen, got.Status)
			}
			m.repo.AssertExpectations(t)
			m.ids.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── GetByID ─────────────────────────────────────────────────────────────────

func TestMaintenanceService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		setup   func(m maintenanceMocks)
		wantErr error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			id:   "ticket-1",
			setup: func(m maintenanceMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "ticket-1").Return(validTicket(t), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "ticket-1",
			setup:   func(_ maintenanceMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "not_found",
			ctx:  context.Background(),
			id:   "missing",
			setup: func(m maintenanceMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrTicketNotFound)
			},
			wantErr: domain.ErrTicketNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestMaintenanceService(t)
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

// ─── StartWork ───────────────────────────────────────────────────────────────

func TestMaintenanceService_StartWork(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		ownerID string
		setup   func(m maintenanceMocks)
		wantErr error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			id:      "ticket-1",
			ownerID: "owner-1",
			setup: func(m maintenanceMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "ticket-1").Return(validTicket(t), nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "ticket-1",
			ownerID: "owner-1",
			setup:   func(_ maintenanceMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:    "wrong_owner",
			ctx:     context.Background(),
			id:      "ticket-1",
			ownerID: "other",
			setup: func(m maintenanceMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "ticket-1").Return(validTicket(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
		{
			name:    "invalid_transition_already_in_progress",
			ctx:     context.Background(),
			id:      "ticket-1",
			ownerID: "owner-1",
			setup: func(m maintenanceMocks) {
				ticket := validTicket(t)
				ticket.Status = domain.TicketStatusInProgress
				m.repo.EXPECT().FindByID(mock.Anything, "ticket-1").Return(ticket, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: errors.New("only open tickets can be started"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestMaintenanceService(t)
			tc.setup(m)

			got, err := svc.StartWork(tc.ctx, tc.id, tc.ownerID)

			if tc.wantErr != nil {
				require.Error(t, err)
				// For sentinel errors use errors.Is; transition errors are string-wrapped
				if errors.Is(tc.wantErr, domain.ErrUnauthorizedAccess) || errors.Is(tc.wantErr, context.Canceled) {
					assert.True(t, errors.Is(err, tc.wantErr))
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, domain.TicketStatusInProgress, got.Status)
			}
			m.repo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── Resolve ─────────────────────────────────────────────────────────────────

func TestMaintenanceService_Resolve(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		ownerID string
		notes   string
		setup   func(m maintenanceMocks)
		wantErr error
	}{
		{
			name:    "success_from_in_progress",
			ctx:     context.Background(),
			id:      "ticket-1",
			ownerID: "owner-1",
			notes:   "Fixed the AC",
			setup: func(m maintenanceMocks) {
				ticket := validTicket(t)
				ticket.Status = domain.TicketStatusInProgress
				m.repo.EXPECT().FindByID(mock.Anything, "ticket-1").Return(ticket, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "ticket-1",
			ownerID: "owner-1",
			notes:   "Fixed",
			setup:   func(_ maintenanceMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:    "wrong_owner",
			ctx:     context.Background(),
			id:      "ticket-1",
			ownerID: "other",
			notes:   "Fixed",
			setup: func(m maintenanceMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "ticket-1").Return(validTicket(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
		{
			name:    "invalid_transition_closed",
			ctx:     context.Background(),
			id:      "ticket-1",
			ownerID: "owner-1",
			notes:   "notes",
			setup: func(m maintenanceMocks) {
				ticket := validTicket(t)
				ticket.Status = domain.TicketStatusClosed
				m.repo.EXPECT().FindByID(mock.Anything, "ticket-1").Return(ticket, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: errors.New("cannot resolve"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestMaintenanceService(t)
			tc.setup(m)

			got, err := svc.Resolve(tc.ctx, tc.id, tc.ownerID, tc.notes)

			if tc.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tc.wantErr, domain.ErrUnauthorizedAccess) {
					assert.True(t, errors.Is(err, domain.ErrUnauthorizedAccess))
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, domain.TicketStatusResolved, got.Status)
				assert.Equal(t, "Fixed the AC", got.Notes)
			}
			m.repo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── Close ───────────────────────────────────────────────────────────────────

func TestMaintenanceService_Close(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		id       string
		tenantID string
		setup    func(m maintenanceMocks)
		wantErr  error
	}{
		{
			name:     "success",
			ctx:      context.Background(),
			id:       "ticket-1",
			tenantID: "tenant-1",
			setup: func(m maintenanceMocks) {
				ticket := validTicket(t)
				ticket.Status = domain.TicketStatusResolved
				m.repo.EXPECT().FindByID(mock.Anything, "ticket-1").Return(ticket, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:     "context_already_cancelled",
			ctx:      cancelledCtx(),
			id:       "ticket-1",
			tenantID: "tenant-1",
			setup:    func(_ maintenanceMocks) {},
			wantErr:  context.Canceled,
		},
		{
			name:     "wrong_tenant",
			ctx:      context.Background(),
			id:       "ticket-1",
			tenantID: "other-tenant",
			setup: func(m maintenanceMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "ticket-1").Return(validTicket(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
		{
			name:     "invalid_transition_open",
			ctx:      context.Background(),
			id:       "ticket-1",
			tenantID: "tenant-1",
			setup: func(m maintenanceMocks) {
				ticket := validTicket(t) // status = open, not resolved
				m.repo.EXPECT().FindByID(mock.Anything, "ticket-1").Return(ticket, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: errors.New("only resolved tickets can be closed"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestMaintenanceService(t)
			tc.setup(m)

			got, err := svc.Close(tc.ctx, tc.id, tc.tenantID)

			if tc.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tc.wantErr, domain.ErrUnauthorizedAccess) {
					assert.True(t, errors.Is(err, domain.ErrUnauthorizedAccess))
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, domain.TicketStatusClosed, got.Status)
			}
			m.repo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}
