package dto

import (
	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
)

// InvoiceItemRequest is a single line item for invoice creation.
type InvoiceItemRequest struct {
	Description string  `json:"description" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" validate:"required,min=0"`
}

// CreateInvoiceRequest is the request body for creating an invoice.
type CreateInvoiceRequest struct {
	BookingID string               `json:"booking_id" validate:"required"`
	TenantID  string               `json:"tenant_id" validate:"required"`
	Items     []InvoiceItemRequest `json:"items" validate:"required,min=1,dive"`
	DueDate   string               `json:"due_date" validate:"required"` // YYYY-MM-DD
	Month     int                  `json:"month" validate:"required,min=1,max=12"`
	Year      int                  `json:"year" validate:"required,min=2020"`
	Notes     string               `json:"notes" validate:"max=500"`
}

// InvoiceItemResponse is the API representation of an invoice line item.
type InvoiceItemResponse struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}

// InvoiceResponse is the API representation of an invoice.
type InvoiceResponse struct {
	ID          string                `json:"id"`
	BookingID   string                `json:"booking_id"`
	TenantID    string                `json:"tenant_id"`
	OwnerID     string                `json:"owner_id"`
	Items       []InvoiceItemResponse `json:"items"`
	TotalAmount float64               `json:"total_amount"`
	Status      string                `json:"status"`
	DueDate     string                `json:"due_date"`
	Month       int                   `json:"month"`
	Year        int                   `json:"year"`
	Notes       string                `json:"notes"`
	PaidAt      *string               `json:"paid_at,omitempty"`
	CreatedAt   string                `json:"created_at"`
	UpdatedAt   string                `json:"updated_at"`
}

// InvoiceListResponse wraps a paginated list of invoices.
type InvoiceListResponse struct {
	Data  []InvoiceResponse `json:"data"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
}

// ToInvoiceItemInput converts DTO items to service input items.
func ToInvoiceItemInput(items []InvoiceItemRequest) []inbound.InvoiceItemInput {
	out := make([]inbound.InvoiceItemInput, 0, len(items))
	for _, it := range items {
		out = append(out, inbound.InvoiceItemInput{
			Description: it.Description,
			Quantity:    it.Quantity,
			UnitPrice:   it.UnitPrice,
		})
	}
	return out
}

// ToInvoiceResponse maps a domain Invoice to InvoiceResponse.
func ToInvoiceResponse(inv *domain.Invoice) InvoiceResponse {
	items := make([]InvoiceItemResponse, 0, len(inv.Items))
	for _, it := range inv.Items {
		items = append(items, InvoiceItemResponse{
			Description: it.Description,
			Quantity:    it.Quantity,
			UnitPrice:   it.UnitPrice,
			Total:       it.Total,
		})
	}
	r := InvoiceResponse{
		ID:          inv.ID,
		BookingID:   inv.BookingID,
		TenantID:    inv.TenantID,
		OwnerID:     inv.OwnerID,
		Items:       items,
		TotalAmount: inv.TotalAmount,
		Status:      string(inv.Status),
		DueDate:     inv.DueDate.Format("2006-01-02"),
		Month:       inv.Month,
		Year:        inv.Year,
		Notes:       inv.Notes,
		CreatedAt:   inv.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   inv.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if inv.PaidAt != nil {
		s := inv.PaidAt.Format("2006-01-02T15:04:05Z")
		r.PaidAt = &s
	}
	return r
}

// ToInvoiceListResponse builds a paginated invoice list response.
func ToInvoiceListResponse(invoices []*domain.Invoice, total int64, page, limit int) InvoiceListResponse {
	data := make([]InvoiceResponse, 0, len(invoices))
	for _, inv := range invoices {
		data = append(data, ToInvoiceResponse(inv))
	}
	return InvoiceListResponse{Data: data, Total: total, Page: page, Limit: limit}
}
