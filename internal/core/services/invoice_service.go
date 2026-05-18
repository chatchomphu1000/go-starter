package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/outbound"
	"github.com/chatchomphu1000/go-starter/pkg/apperrors"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

type invoiceService struct {
	repo    outbound.InvoiceRepository
	userRepo outbound.UserRepository
	clock   outbound.Clock
	ids     outbound.IDGenerator
	log     logger.Logger
}

// NewInvoiceService creates a new InvoiceService.
func NewInvoiceService(
	repo outbound.InvoiceRepository,
	userRepo outbound.UserRepository,
	clock outbound.Clock,
	ids outbound.IDGenerator,
	log logger.Logger,
) inbound.InvoiceService {
	return &invoiceService{repo: repo, userRepo: userRepo, clock: clock, ids: ids, log: log}
}

// Create generates a new invoice draft.
func (s *invoiceService) Create(ctx context.Context, in inbound.CreateInvoiceInput) (*domain.Invoice, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("invoiceService.Create: %w", err)
	}

	items := make([]domain.InvoiceItem, 0, len(in.Items))
	for _, i := range in.Items {
		items = append(items, domain.InvoiceItem{
			Description: i.Description,
			Quantity:    i.Quantity,
			UnitPrice:   i.UnitPrice,
			Total:       float64(i.Quantity) * i.UnitPrice,
		})
	}

	inv, err := domain.NewInvoice(
		s.ids.New(), in.BookingID, in.TenantID, in.OwnerID,
		items, in.DueDate, in.Month, in.Year, in.Notes, s.clock.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("invoiceService.Create: %w", apperrors.BadRequest(err.Error(), err))
	}

	if err := s.repo.Insert(ctx, inv); err != nil {
		return nil, fmt.Errorf("invoiceService.Create: %w", err)
	}

	s.log.Info("invoice created", zap.String("invoice_id", inv.ID), zap.Float64("total", inv.TotalAmount))
	return inv, nil
}

// GetByID retrieves an invoice by ID.
func (s *invoiceService) GetByID(ctx context.Context, id string) (*domain.Invoice, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("invoiceService.GetByID: %w", err)
	}
	inv, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("invoiceService.GetByID: %w", err)
	}
	return inv, nil
}

// List returns filtered invoices.
func (s *invoiceService) List(ctx context.Context, f inbound.InvoiceFilter) ([]*domain.Invoice, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, fmt.Errorf("invoiceService.List: %w", err)
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	return s.repo.FindAll(ctx, f)
}

// Send marks the invoice as sent to tenant.
func (s *invoiceService) Send(ctx context.Context, id, ownerID string) (*domain.Invoice, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("invoiceService.Send: %w", err)
	}
	inv, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("invoiceService.Send: %w", err)
	}
	if inv.OwnerID != ownerID {
		return nil, fmt.Errorf("invoiceService.Send: %w", domain.ErrUnauthorizedAccess)
	}
	if err := inv.Send(s.clock.Now()); err != nil {
		return nil, fmt.Errorf("invoiceService.Send: %w", apperrors.BadRequest(err.Error(), err))
	}
	if err := s.repo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("invoiceService.Send: %w", err)
	}
	s.log.Info("invoice sent", zap.String("invoice_id", id))
	return inv, nil
}

// MarkPaid marks the invoice as paid by the tenant.
func (s *invoiceService) MarkPaid(ctx context.Context, id, ownerID string) (*domain.Invoice, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("invoiceService.MarkPaid: %w", err)
	}
	inv, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("invoiceService.MarkPaid: %w", err)
	}
	if inv.OwnerID != ownerID {
		return nil, fmt.Errorf("invoiceService.MarkPaid: %w", domain.ErrUnauthorizedAccess)
	}
	if err := inv.MarkPaid(s.clock.Now()); err != nil {
		return nil, fmt.Errorf("invoiceService.MarkPaid: %w", apperrors.BadRequest(err.Error(), err))
	}
	if err := s.repo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("invoiceService.MarkPaid: %w", err)
	}
	s.log.Info("invoice marked paid", zap.String("invoice_id", id))
	return inv, nil
}

// Cancel voids the invoice.
func (s *invoiceService) Cancel(ctx context.Context, id, ownerID string) (*domain.Invoice, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("invoiceService.Cancel: %w", err)
	}
	inv, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("invoiceService.Cancel: %w", err)
	}
	if inv.OwnerID != ownerID {
		return nil, fmt.Errorf("invoiceService.Cancel: %w", domain.ErrUnauthorizedAccess)
	}
	inv.Cancel(s.clock.Now())
	if err := s.repo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("invoiceService.Cancel: %w", err)
	}
	s.log.Info("invoice cancelled", zap.String("invoice_id", id))
	return inv, nil
}

// GeneratePDF renders the invoice as print-ready HTML (text/html).
func (s *invoiceService) GeneratePDF(ctx context.Context, id string) ([]byte, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", fmt.Errorf("invoiceService.GeneratePDF: %w", err)
	}

	inv, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, "", fmt.Errorf("invoiceService.GeneratePDF: %w", err)
	}

	data := invoicePDFData{
		Invoice:   inv,
		Generated: time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
	}

	var buf bytes.Buffer
	if err := invoiceTmpl.Execute(&buf, data); err != nil {
		return nil, "", fmt.Errorf("invoiceService.GeneratePDF: template: %w", err)
	}

	return buf.Bytes(), "text/html; charset=utf-8", nil
}

type invoicePDFData struct {
	Invoice   *domain.Invoice
	Generated string
}

var invoiceTmpl = template.Must(template.New("invoice").Funcs(template.FuncMap{
	"monthName": func(m int) string {
		return time.Month(m).String()
	},
}).Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Invoice #{{.Invoice.ID}}</title>
<style>
  body { font-family: Arial, sans-serif; margin: 40px; color: #333; }
  h1 { color: #2c3e50; }
  table { width: 100%; border-collapse: collapse; margin-top: 20px; }
  th, td { padding: 10px; border: 1px solid #ddd; text-align: left; }
  th { background: #f4f4f4; }
  .total { font-weight: bold; font-size: 1.1em; }
  .badge { padding: 4px 10px; border-radius: 12px; font-size: .85em; }
  .paid { background: #d4edda; color: #155724; }
  .overdue { background: #f8d7da; color: #721c24; }
  .sent { background: #d1ecf1; color: #0c5460; }
  .draft { background: #e2e3e5; color: #383d41; }
  @media print { button { display: none; } }
</style>
</head>
<body>
<button onclick="window.print()">Print / Save as PDF</button>
<h1>Invoice</h1>
<p><strong>Invoice ID:</strong> {{.Invoice.ID}}</p>
<p><strong>Billing Period:</strong> {{monthName .Invoice.Month}} {{.Invoice.Year}}</p>
<p><strong>Status:</strong> <span class="badge {{.Invoice.Status}}">{{.Invoice.Status}}</span></p>
<p><strong>Due Date:</strong> {{.Invoice.DueDate.Format "2006-01-02"}}</p>
{{if .Invoice.PaidAt}}<p><strong>Paid At:</strong> {{.Invoice.PaidAt.Format "2006-01-02 15:04"}}</p>{{end}}
{{if .Invoice.Notes}}<p><strong>Notes:</strong> {{.Invoice.Notes}}</p>{{end}}
<table>
  <thead>
    <tr><th>Description</th><th>Qty</th><th>Unit Price</th><th>Total</th></tr>
  </thead>
  <tbody>
    {{range .Invoice.Items}}
    <tr>
      <td>{{.Description}}</td>
      <td>{{.Quantity}}</td>
      <td>{{printf "%.2f" .UnitPrice}}</td>
      <td>{{printf "%.2f" .Total}}</td>
    </tr>
    {{end}}
    <tr class="total">
      <td colspan="3">Total Amount</td>
      <td>{{printf "%.2f" .Invoice.TotalAmount}}</td>
    </tr>
  </tbody>
</table>
<p style="margin-top:40px;font-size:.8em;color:#999">Generated: {{.Generated}}</p>
</body>
</html>`))
