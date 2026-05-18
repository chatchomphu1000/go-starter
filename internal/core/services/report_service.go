package services

import (
	"context"
	"fmt"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/outbound"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

type reportService struct {
	roomRepo        outbound.RoomRepository
	bookingRepo     outbound.BookingRepository
	paymentRepo     outbound.PaymentRepository
	invoiceRepo     outbound.InvoiceRepository
	maintenanceRepo outbound.MaintenanceRepository
	messageRepo     outbound.MessageRepository
	log             logger.Logger
}

// NewReportService creates a new ReportService.
func NewReportService(
	roomRepo outbound.RoomRepository,
	bookingRepo outbound.BookingRepository,
	paymentRepo outbound.PaymentRepository,
	invoiceRepo outbound.InvoiceRepository,
	maintenanceRepo outbound.MaintenanceRepository,
	messageRepo outbound.MessageRepository,
	log logger.Logger,
) inbound.ReportService {
	return &reportService{
		roomRepo:        roomRepo,
		bookingRepo:     bookingRepo,
		paymentRepo:     paymentRepo,
		invoiceRepo:     invoiceRepo,
		maintenanceRepo: maintenanceRepo,
		messageRepo:     messageRepo,
		log:             log,
	}
}

// IncomeExpense returns aggregated income and invoice data for a given month.
func (s *reportService) IncomeExpense(ctx context.Context, ownerID string, month, year int) (*inbound.IncomeExpenseReport, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("reportService.IncomeExpense: %w", err)
	}

	income, payCount, err := s.paymentRepo.SumByOwnerAndMonth(ctx, ownerID, month, year)
	if err != nil {
		return nil, fmt.Errorf("reportService.IncomeExpense: payments: %w", err)
	}

	invoiced, invCount, err := s.invoiceRepo.SumByOwnerAndMonth(ctx, ownerID, month, year)
	if err != nil {
		return nil, fmt.Errorf("reportService.IncomeExpense: invoices: %w", err)
	}

	return &inbound.IncomeExpenseReport{
		OwnerID:       ownerID,
		Month:         month,
		Year:          year,
		TotalIncome:   income,
		TotalInvoiced: invoiced,
		Outstanding:   invoiced - income,
		PaymentCount:  payCount,
		InvoiceCount:  invCount,
	}, nil
}

// OwnerDashboard aggregates summary metrics for a property owner.
func (s *reportService) OwnerDashboard(ctx context.Context, ownerID string) (*inbound.OwnerDashboard, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("reportService.OwnerDashboard: %w", err)
	}

	available := domain.RoomStatusAvailable
	occupied := domain.RoomStatusOccupied
	maint := domain.RoomStatusMaintenance

	_, totalRooms, _ := s.roomRepo.FindAll(ctx, inbound.RoomFilter{OwnerID: ownerID, Page: 1, Limit: 1})
	_, availableRooms, _ := s.roomRepo.FindAll(ctx, inbound.RoomFilter{OwnerID: ownerID, Status: &available, Page: 1, Limit: 1})
	_, occupiedRooms, _ := s.roomRepo.FindAll(ctx, inbound.RoomFilter{OwnerID: ownerID, Status: &occupied, Page: 1, Limit: 1})
	_, maintRooms, _ := s.roomRepo.FindAll(ctx, inbound.RoomFilter{OwnerID: ownerID, Status: &maint, Page: 1, Limit: 1})

	activeStatus := domain.BookingStatusActive
	pendingStatus := domain.BookingStatusPending
	activeBookings, _ := s.bookingRepo.CountByStatus(ctx, ownerID, activeStatus)
	pendingBookings, _ := s.bookingRepo.CountByStatus(ctx, ownerID, pendingStatus)

	openMaintenance, _ := s.maintenanceRepo.CountOpenByOwner(ctx, ownerID)
	overdueInvoices, _ := s.invoiceRepo.CountByOwnerAndStatus(ctx, ownerID, domain.InvoiceStatusOverdue)
	unreadMessages, _ := s.messageRepo.CountUnreadByUser(ctx, ownerID)

	return &inbound.OwnerDashboard{
		OwnerID:            ownerID,
		TotalRooms:         totalRooms,
		AvailableRooms:     availableRooms,
		OccupiedRooms:      occupiedRooms,
		MaintenanceRooms:   maintRooms,
		ActiveBookings:     activeBookings,
		PendingBookings:    pendingBookings,
		PendingMaintenance: openMaintenance,
		OverdueInvoices:    overdueInvoices,
		UnreadMessages:     unreadMessages,
	}, nil
}

// TenantStatement returns a tenant's payment summary for a given month.
func (s *reportService) TenantStatement(ctx context.Context, tenantID string, month, year int) (*inbound.IncomeExpenseReport, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("reportService.TenantStatement: %w", err)
	}

	f := inbound.PaymentFilter{TenantID: tenantID, Page: 1, Limit: 100}
	payments, _, err := s.paymentRepo.FindAll(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("reportService.TenantStatement: %w", err)
	}

	var totalPaid float64
	var payCount int64
	for _, p := range payments {
		if p.PaidAt == nil {
			continue
		}
		pMonth := int(p.PaidAt.Month())
		pYear := p.PaidAt.Year()
		if pMonth == month && pYear == year && p.Status == domain.PaymentStatusCompleted {
			totalPaid += p.Amount
			payCount++
		}
	}

	invF := inbound.InvoiceFilter{TenantID: tenantID, Month: &month, Year: &year, Page: 1, Limit: 100}
	invoices, invTotal, err := s.invoiceRepo.FindAll(ctx, invF)
	if err != nil {
		return nil, fmt.Errorf("reportService.TenantStatement: %w", err)
	}

	var totalInvoiced float64
	for _, inv := range invoices {
		totalInvoiced += inv.TotalAmount
	}

	return &inbound.IncomeExpenseReport{
		OwnerID:       tenantID,
		Month:         month,
		Year:          year,
		TotalIncome:   totalPaid,
		TotalInvoiced: totalInvoiced,
		Outstanding:   totalInvoiced - totalPaid,
		PaymentCount:  payCount,
		InvoiceCount:  invTotal,
	}, nil
}
