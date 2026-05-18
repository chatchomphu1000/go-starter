package inbound

import "context"

//go:generate go run github.com/vektra/mockery/v2 --name=ReportService

// IncomeExpenseReport holds aggregated financial data for a period.
type IncomeExpenseReport struct {
	OwnerID       string
	Month         int
	Year          int
	TotalIncome   float64
	TotalInvoiced float64
	Outstanding   float64
	PaymentCount  int64
	InvoiceCount  int64
}

// OwnerDashboard holds summary metrics for an owner.
type OwnerDashboard struct {
	OwnerID             string
	TotalRooms          int64
	AvailableRooms      int64
	OccupiedRooms       int64
	MaintenanceRooms    int64
	ActiveBookings      int64
	PendingBookings     int64
	MonthlyRevenue      float64
	PendingMaintenance  int64
	OverdueInvoices     int64
	UnreadMessages      int64
}

// ReportService defines the inbound port for analytics and reporting.
type ReportService interface {
	IncomeExpense(ctx context.Context, ownerID string, month, year int) (*IncomeExpenseReport, error)
	OwnerDashboard(ctx context.Context, ownerID string) (*OwnerDashboard, error)
	TenantStatement(ctx context.Context, tenantID string, month, year int) (*IncomeExpenseReport, error)
}
