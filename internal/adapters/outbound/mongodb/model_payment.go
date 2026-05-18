package mongodb

import (
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

type paymentDoc struct {
	ID             string     `bson:"_id"`
	BookingID      string     `bson:"booking_id"`
	InvoiceID      string     `bson:"invoice_id"`
	TenantID       string     `bson:"tenant_id"`
	OwnerID        string     `bson:"owner_id"`
	Amount         float64    `bson:"amount"`
	Currency       string     `bson:"currency"`
	Status         string     `bson:"status"`
	Method         string     `bson:"method"`
	Gateway        string     `bson:"gateway"`
	GatewayTxID    string     `bson:"gateway_tx_id"`
	WebhookPayload string     `bson:"webhook_payload"`
	PaidAt         *time.Time `bson:"paid_at"`
	Description    string     `bson:"description"`
	CreatedAt      time.Time  `bson:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at"`
}

func paymentFromDomain(p *domain.Payment) paymentDoc {
	return paymentDoc{
		ID:             p.ID,
		BookingID:      p.BookingID,
		InvoiceID:      p.InvoiceID,
		TenantID:       p.TenantID,
		OwnerID:        p.OwnerID,
		Amount:         p.Amount,
		Currency:       p.Currency,
		Status:         string(p.Status),
		Method:         string(p.Method),
		Gateway:        p.Gateway,
		GatewayTxID:    p.GatewayTxID,
		WebhookPayload: p.WebhookPayload,
		PaidAt:         p.PaidAt,
		Description:    p.Description,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func paymentToDomain(d paymentDoc) *domain.Payment {
	return &domain.Payment{
		ID:             d.ID,
		BookingID:      d.BookingID,
		InvoiceID:      d.InvoiceID,
		TenantID:       d.TenantID,
		OwnerID:        d.OwnerID,
		Amount:         d.Amount,
		Currency:       d.Currency,
		Status:         domain.PaymentStatus(d.Status),
		Method:         domain.PaymentMethod(d.Method),
		Gateway:        d.Gateway,
		GatewayTxID:    d.GatewayTxID,
		WebhookPayload: d.WebhookPayload,
		PaidAt:         d.PaidAt,
		Description:    d.Description,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}
