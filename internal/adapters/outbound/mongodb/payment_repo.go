package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
)

const collectionPayments = "payments"

// PaymentRepo implements outbound.PaymentRepository for MongoDB.
type PaymentRepo struct {
	col *mongo.Collection
}

// NewPaymentRepo creates a new MongoDB PaymentRepository.
func NewPaymentRepo(db *mongo.Database) *PaymentRepo {
	return &PaymentRepo{col: db.Collection(collectionPayments)}
}

// Insert persists a new payment.
func (r *PaymentRepo) Insert(ctx context.Context, p *domain.Payment) error {
	if _, err := r.col.InsertOne(ctx, paymentFromDomain(p)); err != nil {
		return fmt.Errorf("mongodb.PaymentRepo.Insert: %w", err)
	}
	return nil
}

// FindByID retrieves a payment by ID.
func (r *PaymentRepo) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
	var doc paymentDoc
	if err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("mongodb.PaymentRepo.FindByID: %w", domain.ErrPaymentNotFound)
		}
		return nil, fmt.Errorf("mongodb.PaymentRepo.FindByID: %w", err)
	}
	return paymentToDomain(doc), nil
}

// FindAll returns filtered payments.
func (r *PaymentRepo) FindAll(ctx context.Context, f inbound.PaymentFilter) ([]*domain.Payment, int64, error) {
	filter := bson.M{}
	if f.BookingID != "" {
		filter["booking_id"] = f.BookingID
	}
	if f.InvoiceID != "" {
		filter["invoice_id"] = f.InvoiceID
	}
	if f.TenantID != "" {
		filter["tenant_id"] = f.TenantID
	}
	if f.OwnerID != "" {
		filter["owner_id"] = f.OwnerID
	}
	if f.Status != nil {
		filter["status"] = string(*f.Status)
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.PaymentRepo.FindAll: count: %w", err)
	}

	sortOrder := -1
	if !f.SortDesc {
		sortOrder = 1
	}
	skip := int64((f.Page - 1) * f.Limit)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: sortOrder}}).
		SetSkip(skip).
		SetLimit(int64(f.Limit))

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.PaymentRepo.FindAll: find: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []paymentDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("mongodb.PaymentRepo.FindAll: decode: %w", err)
	}

	payments := make([]*domain.Payment, 0, len(docs))
	for _, d := range docs {
		payments = append(payments, paymentToDomain(d))
	}
	return payments, total, nil
}

// Update persists payment state changes.
func (r *PaymentRepo) Update(ctx context.Context, p *domain.Payment) error {
	doc := paymentFromDomain(p)
	result, err := r.col.UpdateByID(ctx, p.ID, bson.M{"$set": bson.M{
		"status":          doc.Status,
		"gateway_tx_id":   doc.GatewayTxID,
		"webhook_payload": doc.WebhookPayload,
		"paid_at":         doc.PaidAt,
		"updated_at":      doc.UpdatedAt,
	}})
	if err != nil {
		return fmt.Errorf("mongodb.PaymentRepo.Update: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("mongodb.PaymentRepo.Update: %w", domain.ErrPaymentNotFound)
	}
	return nil
}

// SumByOwnerAndMonth aggregates total income for a given month.
func (r *PaymentRepo) SumByOwnerAndMonth(ctx context.Context, ownerID string, month, year int) (float64, int64, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	pipeline := bson.A{
		bson.M{"$match": bson.M{
			"owner_id": ownerID,
			"status":   "completed",
			"paid_at":  bson.M{"$gte": start, "$lt": end},
		}},
		bson.M{"$group": bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$amount"},
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, fmt.Errorf("mongodb.PaymentRepo.SumByOwnerAndMonth: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		Total float64 `bson:"total"`
		Count int64   `bson:"count"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, 0, fmt.Errorf("mongodb.PaymentRepo.SumByOwnerAndMonth: decode: %w", err)
		}
	}
	return result.Total, result.Count, nil
}
