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

const collectionInvoices = "invoices"

// InvoiceRepo implements outbound.InvoiceRepository for MongoDB.
type InvoiceRepo struct {
	col *mongo.Collection
}

// NewInvoiceRepo creates a new MongoDB InvoiceRepository.
func NewInvoiceRepo(db *mongo.Database) *InvoiceRepo {
	return &InvoiceRepo{col: db.Collection(collectionInvoices)}
}

// Insert persists a new invoice.
func (r *InvoiceRepo) Insert(ctx context.Context, inv *domain.Invoice) error {
	if _, err := r.col.InsertOne(ctx, invoiceFromDomain(inv)); err != nil {
		return fmt.Errorf("mongodb.InvoiceRepo.Insert: %w", err)
	}
	return nil
}

// FindByID retrieves an invoice by ID.
func (r *InvoiceRepo) FindByID(ctx context.Context, id string) (*domain.Invoice, error) {
	var doc invoiceDoc
	if err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("mongodb.InvoiceRepo.FindByID: %w", domain.ErrInvoiceNotFound)
		}
		return nil, fmt.Errorf("mongodb.InvoiceRepo.FindByID: %w", err)
	}
	return invoiceToDomain(doc), nil
}

// FindAll returns filtered invoices.
func (r *InvoiceRepo) FindAll(ctx context.Context, f inbound.InvoiceFilter) ([]*domain.Invoice, int64, error) {
	filter := bson.M{}
	if f.BookingID != "" {
		filter["booking_id"] = f.BookingID
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
	if f.Month != nil {
		filter["month"] = *f.Month
	}
	if f.Year != nil {
		filter["year"] = *f.Year
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.InvoiceRepo.FindAll: count: %w", err)
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
		return nil, 0, fmt.Errorf("mongodb.InvoiceRepo.FindAll: find: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []invoiceDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("mongodb.InvoiceRepo.FindAll: decode: %w", err)
	}

	invoices := make([]*domain.Invoice, 0, len(docs))
	for _, d := range docs {
		invoices = append(invoices, invoiceToDomain(d))
	}
	return invoices, total, nil
}

// Update persists invoice state changes.
func (r *InvoiceRepo) Update(ctx context.Context, inv *domain.Invoice) error {
	doc := invoiceFromDomain(inv)
	result, err := r.col.UpdateByID(ctx, inv.ID, bson.M{"$set": bson.M{
		"status":     doc.Status,
		"paid_at":    doc.PaidAt,
		"updated_at": doc.UpdatedAt,
	}})
	if err != nil {
		return fmt.Errorf("mongodb.InvoiceRepo.Update: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("mongodb.InvoiceRepo.Update: %w", domain.ErrInvoiceNotFound)
	}
	return nil
}

// FindOverdue returns all sent invoices past their due date.
func (r *InvoiceRepo) FindOverdue(ctx context.Context) ([]*domain.Invoice, error) {
	filter := bson.M{
		"status":   "sent",
		"due_date": bson.M{"$lt": time.Now().UTC()},
	}
	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("mongodb.InvoiceRepo.FindOverdue: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []invoiceDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("mongodb.InvoiceRepo.FindOverdue: decode: %w", err)
	}

	invoices := make([]*domain.Invoice, 0, len(docs))
	for _, d := range docs {
		invoices = append(invoices, invoiceToDomain(d))
	}
	return invoices, nil
}

// CountByOwnerAndStatus counts invoices for an owner by status.
func (r *InvoiceRepo) CountByOwnerAndStatus(ctx context.Context, ownerID string, status domain.InvoiceStatus) (int64, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{"owner_id": ownerID, "status": string(status)})
	if err != nil {
		return 0, fmt.Errorf("mongodb.InvoiceRepo.CountByOwnerAndStatus: %w", err)
	}
	return count, nil
}

// SumByOwnerAndMonth aggregates invoice totals for a given month.
func (r *InvoiceRepo) SumByOwnerAndMonth(ctx context.Context, ownerID string, month, year int) (float64, int64, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{
			"owner_id": ownerID,
			"month":    month,
			"year":     year,
		}},
		bson.M{"$group": bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": "$total_amount"},
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, fmt.Errorf("mongodb.InvoiceRepo.SumByOwnerAndMonth: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		Total float64 `bson:"total"`
		Count int64   `bson:"count"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, 0, fmt.Errorf("mongodb.InvoiceRepo.SumByOwnerAndMonth: decode: %w", err)
		}
	}
	return result.Total, result.Count, nil
}
