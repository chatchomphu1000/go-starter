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

type paymentMocks struct {
	repo        *mocks.MockPaymentRepository
	invoiceRepo *mocks.MockInvoiceRepository
	clock       *mocks.MockClock
	ids         *mocks.MockIDGenerator
}

func newTestPaymentService(t *testing.T) (*paymentService, paymentMocks) {
	t.Helper()
	m := paymentMocks{
		repo:        new(mocks.MockPaymentRepository),
		invoiceRepo: new(mocks.MockInvoiceRepository),
		clock:       new(mocks.MockClock),
		ids:         new(mocks.MockIDGenerator),
	}
	svc := NewPaymentService(m.repo, m.invoiceRepo, m.clock, m.ids, nopLogger()).(*paymentService)
	return svc, m
}

func validPayment(t *testing.T) *domain.Payment {
	t.Helper()
	p, _ := domain.NewPayment(
		"pay-1", "booking-1", "inv-1", "tenant-1", "owner-1",
		5000.0, "THB", domain.PaymentMethodBankTransfer,
		"manual", "Monthly rent", fixedNow,
	)
	return p
}

func validCreatePaymentInput() inbound.CreatePaymentInput {
	return inbound.CreatePaymentInput{
		BookingID:   "booking-1",
		InvoiceID:   "inv-1",
		TenantID:    "tenant-1",
		OwnerID:     "owner-1",
		Amount:      5000.0,
		Currency:    "THB",
		Method:      "bank_transfer",
		Gateway:     "manual",
		Description: "Monthly rent",
	}
}

// ─── Create ──────────────────────────────────────────────────────────────────

func TestPaymentService_Create(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		input   inbound.CreatePaymentInput
		setup   func(m paymentMocks)
		wantErr error
	}{
		{
			name:  "success",
			ctx:   context.Background(),
			input: validCreatePaymentInput(),
			setup: func(m paymentMocks) {
				m.ids.EXPECT().New().Return("pay-1")
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			input:   validCreatePaymentInput(),
			setup:   func(_ paymentMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "zero_amount_validation_error",
			ctx:  context.Background(),
			input: func() inbound.CreatePaymentInput {
				in := validCreatePaymentInput()
				in.Amount = 0
				return in
			}(),
			setup: func(m paymentMocks) {
				m.ids.EXPECT().New().Return("pay-1")
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: errors.New("amount must be positive"),
		},
		{
			name:  "repo_insert_error",
			ctx:   context.Background(),
			input: validCreatePaymentInput(),
			setup: func(m paymentMocks) {
				m.ids.EXPECT().New().Return("pay-1")
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Insert(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestPaymentService(t)
			tc.setup(m)

			got, err := svc.Create(tc.ctx, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, "pay-1", got.ID)
			}
			m.repo.AssertExpectations(t)
			m.ids.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── GetByID ─────────────────────────────────────────────────────────────────

func TestPaymentService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		setup   func(m paymentMocks)
		wantErr error
	}{
		{
			name: "success",
			ctx:  context.Background(),
			id:   "pay-1",
			setup: func(m paymentMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "pay-1").Return(validPayment(t), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "pay-1",
			setup:   func(_ paymentMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "not_found",
			ctx:  context.Background(),
			id:   "missing",
			setup: func(m paymentMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrPaymentNotFound)
			},
			wantErr: domain.ErrPaymentNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestPaymentService(t)
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

// ─── HandleWebhook ───────────────────────────────────────────────────────────

func TestPaymentService_HandleWebhook(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		input   inbound.WebhookInput
		setup   func(m paymentMocks)
		wantErr error
	}{
		{
			name: "success_completed",
			ctx:  context.Background(),
			input: inbound.WebhookInput{
				PaymentID:  "pay-1",
				Status:     "completed",
				GatewayTxID: "tx-123",
				Gateway:    "manual",
			},
			setup: func(m paymentMocks) {
				p := validPayment(t)
				m.repo.EXPECT().FindByID(mock.Anything, "pay-1").Return(p, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.invoiceRepo.EXPECT().FindByID(mock.Anything, p.InvoiceID).Return(validInvoice(t), nil)
				m.invoiceRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name: "success_failed",
			ctx:  context.Background(),
			input: inbound.WebhookInput{
				PaymentID: "pay-1",
				Status:    "failed",
				Gateway:   "manual",
			},
			setup: func(m paymentMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "pay-1").Return(validPayment(t), nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			input:   inbound.WebhookInput{PaymentID: "pay-1", Status: "completed"},
			setup:   func(_ paymentMocks) {},
			wantErr: context.Canceled,
		},
		{
			name: "payment_not_found",
			ctx:  context.Background(),
			input: inbound.WebhookInput{
				PaymentID: "missing",
				Status:    "completed",
			},
			setup: func(m paymentMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "missing").Return(nil, domain.ErrPaymentNotFound)
			},
			wantErr: domain.ErrPaymentNotFound,
		},
		{
			name: "already_completed_is_idempotent",
			ctx:  context.Background(),
			input: inbound.WebhookInput{
				PaymentID:  "pay-1",
				Status:     "completed",
				GatewayTxID: "tx-456",
				Gateway:    "manual",
			},
			setup: func(m paymentMocks) {
				p := validPayment(t)
				p.Status = domain.PaymentStatusCompleted
				m.repo.EXPECT().FindByID(mock.Anything, "pay-1").Return(p, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				// no Update call — idempotent early return
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestPaymentService(t)
			tc.setup(m)

			err := svc.HandleWebhook(tc.ctx, tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
			} else {
				require.NoError(t, err)
			}
			m.repo.AssertExpectations(t)
			m.invoiceRepo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── Refund ──────────────────────────────────────────────────────────────────

func TestPaymentService_Refund(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		ownerID string
		setup   func(m paymentMocks)
		wantErr error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			id:      "pay-1",
			ownerID: "owner-1",
			setup: func(m paymentMocks) {
				p := validPayment(t)
				p.Status = domain.PaymentStatusCompleted
				m.repo.EXPECT().FindByID(mock.Anything, "pay-1").Return(p, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
				m.repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			id:      "pay-1",
			ownerID: "owner-1",
			setup:   func(_ paymentMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:    "wrong_owner",
			ctx:     context.Background(),
			id:      "pay-1",
			ownerID: "other",
			setup: func(m paymentMocks) {
				m.repo.EXPECT().FindByID(mock.Anything, "pay-1").Return(validPayment(t), nil)
			},
			wantErr: domain.ErrUnauthorizedAccess,
		},
		{
			name:    "not_completed_cannot_refund",
			ctx:     context.Background(),
			id:      "pay-1",
			ownerID: "owner-1",
			setup: func(m paymentMocks) {
				p := validPayment(t) // status = pending
				m.repo.EXPECT().FindByID(mock.Anything, "pay-1").Return(p, nil)
				m.clock.EXPECT().Now().Return(fixedNow)
			},
			wantErr: errors.New("only completed payments can be refunded"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestPaymentService(t)
			tc.setup(m)

			got, err := svc.Refund(tc.ctx, tc.id, tc.ownerID)

			if tc.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tc.wantErr, domain.ErrUnauthorizedAccess) {
					assert.True(t, errors.Is(err, domain.ErrUnauthorizedAccess))
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, domain.PaymentStatusRefunded, got.Status)
			}
			m.repo.AssertExpectations(t)
			m.clock.AssertExpectations(t)
		})
	}
}

// ─── List ─────────────────────────────────────────────────────────────────────

func TestPaymentService_List(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		filter  inbound.PaymentFilter
		setup   func(m paymentMocks)
		wantErr error
	}{
		{
			name:   "success",
			ctx:    context.Background(),
			filter: inbound.PaymentFilter{Page: 1, Limit: 10},
			setup: func(m paymentMocks) {
				m.repo.EXPECT().FindAll(mock.Anything, mock.Anything).Return([]*domain.Payment{validPayment(t)}, int64(1), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			filter:  inbound.PaymentFilter{},
			setup:   func(_ paymentMocks) {},
			wantErr: context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestPaymentService(t)
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
