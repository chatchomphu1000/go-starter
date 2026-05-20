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

type invoiceMocks struct {
	repo     *mocks.MockInvoiceRepository
	userRepo *mocks.MockUserRepository
	clock    *mocks.MockClock
	ids      *mocks.MockIDGenerator
}

func newTestInvoiceService(t *testing.T) (*invoiceService, invoiceMocks) {
	t.Helper()
	m := invoiceMocks{
		repo:     new(mocks.MockInvoiceRepository),
		userRepo: new(mocks.MockUserRepository),
		clock:    new(mocks.MockClock),
		ids:      new(mocks.MockIDGenerator),
	}
	svc := NewInvoiceService(m.repo, m.userRepo, m.clock, m.ids, nopLogger()).(*invoiceService)
	return svc, m
}

func validInvoice(t *testing.T) *domain.Invoice {
	t.Helper()
	items := []domain.InvoiceItem{
		{Description: "Rent", Quantity: 1, UnitPrice: 5000.0, Total: 5000.0},
	}
	inv, _ := domain.NewInvoice(
		"inv-1", "booking-1", "tenant-1", "owner-1",
		items, fixedNow.AddDate(0, 0, 30), 1, 2025, "notes", fixedNow,
	)
	return inv
}

func validCreateInvoiceInput() inbound.CreateInvoiceInput {
	return inbound.CreateInvoiceInput{
		BookingID: "booking-1",
		TenantID:  "tenant-1",
		OwnerID:   "owner-1",
		Items: []inbound.InvoiceItemInput{
			{Description: "Rent", Quantity: 1, UnitPrice: 5000.0},
		},
		DueDate: fixedNow.AddDate(0, 0, 30),
		Month:   1,
		Year:    2025,
		Notes:   "notes",
	}
}

// ─── Create ──────────────────────────────────────────────────────────────────

func TestInvoiceService_Create(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		input   inbound.CreateInvoiceInput
		setup   func(m invoiceMocks)
		wantErr error
	}{
		{
			name:  "success",
			ctx:   context.Background(),
			input: validCreateInvoiceInput(),
			setup: func(m invoiceMocks) {
				m.ids.EXPECT().New().Return("inv-1")
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			input:   validCreateInvoiceInput(),
			setup:   func(m invoiceMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "items_summed_correctly",
			ctx:  context.Background(),
			input: inbound.CreateInvoiceInput{
				BookingID: "booking-1",
				TenantID:  "tenant-1",
				OwnerID:   "owner-1",
				Items: []inbound.InvoiceItemInput{
					{Description: "Rent", Quantity: 1, UnitPrice: 5000.0},
					{Description: "Utilities", Quantity: 2, UnitPrice: 200.0},
				},
				DueDate: fixedNow.AddDate(0, 0, 30),
				Month:   1,
				Year:    2025,
			},
			setup: func(m invoiceMocks) {
				m.ids.EXPECT().New().Return("inv-1")
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Insert(mock.Anything, mock.MatchedBy(func(inv *domain.Invoice) bool {
					return inv.TotalAmount == 5400.0 // 5000 + 2*200
				})).Return(nil)
			},
		},
		{
			name:  "repo_insert_error",
			ctx:   context.Background(),
			input: validCreateInvoiceInput(),
			setup: func(m invoiceMocks) {
				m.ids.EXPECT().New().Return("inv-1")
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestInvoiceService(t)
			tc.setup(m)

			got, err := svc.Create(tc.ctx, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}
			m.repo.AssertExpectations(t)
			m.ids.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── GetByID ─────────────────────────────────────────────────────────────────

func TestInvoiceService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		setup   func(m invoiceMocks)
		wantErr error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			id:   "inv-1",
			setup: func(m invoiceMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "inv-1").Return(validInvoice(t), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "inv-1",
			setup:   func(m invoiceMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "not_found",
			ctx:  context.Background(),
			id:   "missing",
			setup: func(m invoiceMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrInvoiceNotFound)
			},
			wantErr: domain.ErrInvoiceNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestInvoiceService(t)
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

// ─── Send ────────────────────────────────────────────────────────────────────

func TestInvoiceService_Send(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		ownerID string
		setup   func(m invoiceMocks)
		wantErr error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			id:      "inv-1",
			ownerID: "owner-1",
			setup: func(m invoiceMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "inv-1").Return(validInvoice(t), nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "inv-1",
			ownerID: "owner-1",
			setup:   func(m invoiceMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:    "wrong_owner",
			ctx:     context.Background(),
			id:      "inv-1",
			ownerID: "other",
			setup: func(m invoiceMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "inv-1").Return(validInvoice(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
		{
			name:    "already_sent_transition_error",
			ctx:     context.Background(),
			id:      "inv-1",
			ownerID: "owner-1",
			setup: func(m invoiceMocks) {
				inv := validInvoice(t)
				inv.Status = domain.InvoiceStatusSent
				m.repo.EXPECT().FindByID(mock.Anything, "inv-1").Return(inv, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: errors.New("only draft invoices can be sent"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestInvoiceService(t)
			tc.setup(m)

			got, err := svc.Send(tc.ctx, tc.id, tc.ownerID)

			if tc.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tc.wantErr, domain.ErrUnauthorizedAccess) {
					assert.True(t, errors.Is(err, domain.ErrUnauthorizedAccess))
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, domain.InvoiceStatusSent, got.Status)
			}
			m.repo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── MarkPaid ────────────────────────────────────────────────────────────────

func TestInvoiceService_MarkPaid(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		ownerID string
		setup   func(m invoiceMocks)
		wantErr error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			id:      "inv-1",
			ownerID: "owner-1",
			setup: func(m invoiceMocks) {
				inv := validInvoice(t)
				inv.Status = domain.InvoiceStatusSent
				m.repo.EXPECT().FindByID(mock.Anything, "inv-1").Return(inv, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "inv-1",
			ownerID: "owner-1",
			setup:   func(m invoiceMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:    "wrong_owner",
			ctx:     context.Background(),
			id:      "inv-1",
			ownerID: "other",
			setup: func(m invoiceMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "inv-1").Return(validInvoice(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
		{
			name:    "already_paid",
			ctx:     context.Background(),
			id:      "inv-1",
			ownerID: "owner-1",
			setup: func(m invoiceMocks) {
				inv := validInvoice(t)
				inv.Status = domain.InvoiceStatusPaid
				m.repo.EXPECT().FindByID(mock.Anything, "inv-1").Return(inv, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: domain.ErrPaymentAlreadyPaid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestInvoiceService(t)
			tc.setup(m)

			got, err := svc.MarkPaid(tc.ctx, tc.id, tc.ownerID)

			if tc.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tc.wantErr, domain.ErrUnauthorizedAccess) || errors.Is(tc.wantErr, domain.ErrPaymentAlreadyPaid) {
					assert.True(t, errors.Is(err, tc.wantErr))
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, domain.InvoiceStatusPaid, got.Status)
				assert.NotNil(t, got.PaidAt)
			}
			m.repo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── Cancel ──────────────────────────────────────────────────────────────────

func TestInvoiceService_Cancel(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		ownerID string
		setup   func(m invoiceMocks)
		wantErr error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			id:      "inv-1",
			ownerID: "owner-1",
			setup: func(m invoiceMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "inv-1").Return(validInvoice(t), nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "inv-1",
			ownerID: "owner-1",
			setup:   func(m invoiceMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:    "wrong_owner",
			ctx:     context.Background(),
			id:      "inv-1",
			ownerID: "other",
			setup: func(m invoiceMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "inv-1").Return(validInvoice(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestInvoiceService(t)
			tc.setup(m)

			got, err := svc.Cancel(tc.ctx, tc.id, tc.ownerID)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, domain.InvoiceStatusCancelled, got.Status)
			}
			m.repo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── GeneratePDF ─────────────────────────────────────────────────────────────

func TestInvoiceService_GeneratePDF(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		setup   func(m invoiceMocks)
		wantErr error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			id:   "inv-1",
			setup: func(m invoiceMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "inv-1").Return(validInvoice(t), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "inv-1",
			setup:   func(m invoiceMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "not_found",
			ctx:  context.Background(),
			id:   "missing",
			setup: func(m invoiceMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrInvoiceNotFound)
			},
			wantErr: domain.ErrInvoiceNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestInvoiceService(t)
			tc.setup(m)

			data, contentType, err := svc.GeneratePDF(tc.ctx, tc.id)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, data)
				assert.Equal(t, "text/html; charset=utf-8", contentType)
			}
			m.repo.AssertExpectations(t)
		})
	}
}

// ─── List ─────────────────────────────────────────────────────────────────────

func TestInvoiceService_List(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		filter  inbound.InvoiceFilter
		setup   func(m invoiceMocks)
		wantErr error
	}{
		{
			name:   "success",
			ctx:    context.Background(),
			filter: inbound.InvoiceFilter{Page: 1, Limit: 10},
			setup: func(m invoiceMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.Anything).Return([]*domain.Invoice{validInvoice(t)}, int64(1), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			filter:  inbound.InvoiceFilter{},
			setup:   func(m invoiceMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:   "repo_error",
			ctx:    context.Background(),
			filter: inbound.InvoiceFilter{Page: 1, Limit: 10},
			setup: func(m invoiceMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestInvoiceService(t)
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

// silence unused import
var _ = time.Now
