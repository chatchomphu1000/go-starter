package mongodb

import (
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

type invoiceItemDoc struct {
	Description string  `bson:"description"`
	Quantity    int     `bson:"quantity"`
	UnitPrice   float64 `bson:"unit_price"`
	Total       float64 `bson:"total"`
}

type invoiceDoc struct {
	ID          string           `bson:"_id"`
	BookingID   string           `bson:"booking_id"`
	TenantID    string           `bson:"tenant_id"`
	OwnerID     string           `bson:"owner_id"`
	Items       []invoiceItemDoc `bson:"items"`
	TotalAmount float64          `bson:"total_amount"`
	Status      string           `bson:"status"`
	DueDate     time.Time        `bson:"due_date"`
	PaidAt      *time.Time       `bson:"paid_at"`
	Month       int              `bson:"month"`
	Year        int              `bson:"year"`
	Notes       string           `bson:"notes"`
	CreatedAt   time.Time        `bson:"created_at"`
	UpdatedAt   time.Time        `bson:"updated_at"`
}

func invoiceFromDomain(inv *domain.Invoice) invoiceDoc {
	items := make([]invoiceItemDoc, 0, len(inv.Items))
	for _, it := range inv.Items {
		items = append(items, invoiceItemDoc{
			Description: it.Description,
			Quantity:    it.Quantity,
			UnitPrice:   it.UnitPrice,
			Total:       it.Total,
		})
	}
	return invoiceDoc{
		ID:          inv.ID,
		BookingID:   inv.BookingID,
		TenantID:    inv.TenantID,
		OwnerID:     inv.OwnerID,
		Items:       items,
		TotalAmount: inv.TotalAmount,
		Status:      string(inv.Status),
		DueDate:     inv.DueDate,
		PaidAt:      inv.PaidAt,
		Month:       inv.Month,
		Year:        inv.Year,
		Notes:       inv.Notes,
		CreatedAt:   inv.CreatedAt,
		UpdatedAt:   inv.UpdatedAt,
	}
}

func invoiceToDomain(d invoiceDoc) *domain.Invoice {
	items := make([]domain.InvoiceItem, 0, len(d.Items))
	for _, it := range d.Items {
		items = append(items, domain.InvoiceItem{
			Description: it.Description,
			Quantity:    it.Quantity,
			UnitPrice:   it.UnitPrice,
			Total:       it.Total,
		})
	}
	return &domain.Invoice{
		ID:          d.ID,
		BookingID:   d.BookingID,
		TenantID:    d.TenantID,
		OwnerID:     d.OwnerID,
		Items:       items,
		TotalAmount: d.TotalAmount,
		Status:      domain.InvoiceStatus(d.Status),
		DueDate:     d.DueDate,
		PaidAt:      d.PaidAt,
		Month:       d.Month,
		Year:        d.Year,
		Notes:       d.Notes,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
