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

type reportMocks struct {
	roomRepo        *mocks.MockRoomRepository
	bookingRepo     *mocks.MockBookingRepository
	paymentRepo     *mocks.MockPaymentRepository
	invoiceRepo     *mocks.MockInvoiceRepository
	maintenanceRepo *mocks.MockMaintenanceRepository
	messageRepo     *mocks.MockMessageRepository
}

func newTestReportService(t *testing.T) (*reportService, reportMocks) {
	t.Helper()
	m := reportMocks{
		roomRepo:        new(mocks.MockRoomRepository),
		bookingRepo:     new(mocks.MockBookingRepository),
		paymentRepo:     new(mocks.MockPaymentRepository),
		invoiceRepo:     new(mocks.MockInvoiceRepository),
		maintenanceRepo: new(mocks.MockMaintenanceRepository),
		messageRepo:     new(mocks.MockMessageRepository),
	}
	svc := NewReportService(m.roomRepo, m.bookingRepo, m.paymentRepo, m.invoiceRepo, m.maintenanceRepo, m.messageRepo, nopLogger()).(*reportService)
	return svc, m
}

// ─── IncomeExpense ────────────────────────────────────────────────────────────

func TestReportService_IncomeExpense(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		ownerID string
		month   int
		year    int
		setup   func(m reportMocks)
		wantErr error
		check   func(t *testing.T, r *inbound.IncomeExpenseReport)
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			ownerID: "owner-1",
			month:   1,
			year:    2025,
			setup: func(m reportMocks) {
				m.paymentRepo.EXPECT().SumByOwnerAndMonth(mock.Anything, "owner-1", 1, 2025).Return(5000.0, int64(1), nil)
				m.invoiceRepo.EXPECT().SumByOwnerAndMonth(mock.Anything, "owner-1", 1, 2025).Return(6000.0, int64(2), nil)
			},
			check: func(t *testing.T, r *inbound.IncomeExpenseReport) {
				assert.Equal(t, 5000.0, r.TotalIncome)
				assert.Equal(t, 6000.0, r.TotalInvoiced)
				assert.Equal(t, 1000.0, r.Outstanding) // 6000 - 5000
				assert.Equal(t, int64(1), r.PaymentCount)
				assert.Equal(t, int64(2), r.InvoiceCount)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			ownerID: "owner-1",
			month:   1,
			year:    2025,
			setup:   func(m reportMocks) {},
			wantErr: context.Canceled,
		},
		{
			name:    "payment_repo_error",
			ctx:     context.Background(),
			ownerID: "owner-1",
			month:   1,
			year:    2025,
			setup: func(m reportMocks) {
				m.paymentRepo.EXPECT().SumByOwnerAndMonth(mock.Anything, "owner-1", 1, 2025).Return(0.0, int64(0), errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
		{
			name:    "invoice_repo_error",
			ctx:     context.Background(),
			ownerID: "owner-1",
			month:   1,
			year:    2025,
			setup: func(m reportMocks) {
				m.paymentRepo.EXPECT().SumByOwnerAndMonth(mock.Anything, "owner-1", 1, 2025).Return(5000.0, int64(1), nil)
				m.invoiceRepo.EXPECT().SumByOwnerAndMonth(mock.Anything, "owner-1", 1, 2025).Return(0.0, int64(0), errors.New("inv error"))
			},
			wantErr: errors.New("inv error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestReportService(t)
			tc.setup(m)

			got, err := svc.IncomeExpense(tc.ctx, tc.ownerID, tc.month, tc.year)

			if tc.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tc.ownerID, got.OwnerID)
				assert.Equal(t, tc.month, got.Month)
				assert.Equal(t, tc.year, got.Year)
				if tc.check != nil {
					tc.check(t, got)
				}
			}
			m.paymentRepo.AssertExpectations(t)
			m.invoiceRepo.AssertExpectations(t)
		})
	}
}

// ─── OwnerDashboard ───────────────────────────────────────────────────────────

func TestReportService_OwnerDashboard(t *testing.T) {
	available := domain.RoomStatusAvailable
	occupied := domain.RoomStatusOccupied
	maint := domain.RoomStatusMaintenance

	tests := []struct {
		name    string
		ctx     context.Context
		ownerID string
		setup   func(m reportMocks)
		wantErr error
	}{
		{
			name:    "success",
			ctx:     context.Background(),
			ownerID: "owner-1",
			setup: func(m reportMocks) {
				m.roomRepo.EXPECT().FindAll(mock.Anything, mock.MatchedBy(func(f inbound.RoomFilter) bool {
					return f.OwnerID == "owner-1" && f.Status == nil
				})).Return([]*domain.Room{}, int64(10), nil)
				m.roomRepo.EXPECT().FindAll(mock.Anything, mock.MatchedBy(func(f inbound.RoomFilter) bool {
					return f.Status != nil && *f.Status == available
				})).Return([]*domain.Room{}, int64(5), nil)
				m.roomRepo.EXPECT().FindAll(mock.Anything, mock.MatchedBy(func(f inbound.RoomFilter) bool {
					return f.Status != nil && *f.Status == occupied
				})).Return([]*domain.Room{}, int64(4), nil)
				m.roomRepo.EXPECT().FindAll(mock.Anything, mock.MatchedBy(func(f inbound.RoomFilter) bool {
					return f.Status != nil && *f.Status == maint
				})).Return([]*domain.Room{}, int64(1), nil)
				m.bookingRepo.EXPECT().CountByStatus(mock.Anything, "owner-1", domain.BookingStatusActive).Return(int64(4), nil)
				m.bookingRepo.EXPECT().CountByStatus(mock.Anything, "owner-1", domain.BookingStatusPending).Return(int64(2), nil)
				m.maintenanceRepo.EXPECT().CountOpenByOwner(mock.Anything, "owner-1").Return(int64(3), nil)
				m.invoiceRepo.EXPECT().CountByOwnerAndStatus(mock.Anything, "owner-1", domain.InvoiceStatusOverdue).Return(int64(1), nil)
				m.messageRepo.EXPECT().CountUnreadByUser(mock.Anything, "owner-1").Return(int64(5), nil)
			},
		},
		{
			name:    "context_already_cancelled",
			ctx:     cancelledCtx(),
			ownerID: "owner-1",
			setup:   func(m reportMocks) {},
			wantErr: context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestReportService(t)
			tc.setup(m)

			got, err := svc.OwnerDashboard(tc.ctx, tc.ownerID)

			if tc.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tc.ownerID, got.OwnerID)
				assert.Equal(t, int64(10), got.TotalRooms)
				assert.Equal(t, int64(5), got.AvailableRooms)
				assert.Equal(t, int64(4), got.OccupiedRooms)
				assert.Equal(t, int64(4), got.ActiveBookings)
				assert.Equal(t, int64(2), got.PendingBookings)
				assert.Equal(t, int64(3), got.PendingMaintenance)
				assert.Equal(t, int64(1), got.OverdueInvoices)
				assert.Equal(t, int64(5), got.UnreadMessages)
			}
			m.roomRepo.AssertExpectations(t)
			m.bookingRepo.AssertExpectations(t)
			m.maintenanceRepo.AssertExpectations(t)
			m.invoiceRepo.AssertExpectations(t)
			m.messageRepo.AssertExpectations(t)
		})
	}
}

// ─── TenantStatement ─────────────────────────────────────────────────────────

func TestReportService_TenantStatement(t *testing.T) {
	paidAt := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)

	completedPayment := func() *domain.Payment {
		p, _ := domain.NewPayment("pay-1", "booking-1", "inv-1", "tenant-1", "owner-1",
			5000.0, "THB", domain.PaymentMethodBankTransfer, "manual", "rent", fixedNow)
		p.Status = domain.PaymentStatusCompleted
		p.PaidAt = &paidAt
		return p
	}

	tests := []struct {
		name     string
		ctx      context.Context
		tenantID string
		month    int
		year     int
		setup    func(m reportMocks)
		wantErr  error
		check    func(t *testing.T, r *inbound.IncomeExpenseReport)
	}{
		{
			name:     "success_with_matching_payment",
			ctx:      context.Background(),
			tenantID: "tenant-1",
			month:    1,
			year:     2025,
			setup: func(m reportMocks) {
				m.paymentRepo.EXPECT().FindAll(mock.Anything, mock.Anything).Return([]*domain.Payment{completedPayment()}, int64(1), nil)
				m.invoiceRepo.EXPECT().FindAll(mock.Anything, mock.Anything).Return([]*domain.Invoice{validInvoice(t)}, int64(1), nil)
			},
			check: func(t *testing.T, r *inbound.IncomeExpenseReport) {
				assert.Equal(t, 5000.0, r.TotalIncome)
				assert.Equal(t, 5000.0, r.TotalInvoiced)
				assert.Equal(t, 0.0, r.Outstanding)
				assert.Equal(t, int64(1), r.PaymentCount)
			},
		},
		{
			name:     "context_already_cancelled",
			ctx:      cancelledCtx(),
			tenantID: "tenant-1",
			month:    1,
			year:     2025,
			setup:    func(m reportMocks) {},
			wantErr:  context.Canceled,
		},
		{
			name:     "payment_repo_error",
			ctx:      context.Background(),
			tenantID: "tenant-1",
			month:    1,
			year:     2025,
			setup: func(m reportMocks) {
				m.paymentRepo.EXPECT().FindAll(mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
		{
			name:     "invoice_repo_error",
			ctx:      context.Background(),
			tenantID: "tenant-1",
			month:    1,
			year:     2025,
			setup: func(m reportMocks) {
				m.paymentRepo.EXPECT().FindAll(mock.Anything, mock.Anything).Return([]*domain.Payment{}, int64(0), nil)
				m.invoiceRepo.EXPECT().FindAll(mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("inv error"))
			},
			wantErr: errors.New("inv error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, m := newTestReportService(t)
			tc.setup(m)

			got, err := svc.TenantStatement(tc.ctx, tc.tenantID, tc.month, tc.year)

			if tc.wantErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tc.tenantID, got.OwnerID)
				if tc.check != nil {
					tc.check(t, got)
				}
			}
			m.paymentRepo.AssertExpectations(t)
			m.invoiceRepo.AssertExpectations(t)
		})
	}
}
