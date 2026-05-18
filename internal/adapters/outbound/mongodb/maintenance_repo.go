package mongodb

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
)

const collectionMaintenance = "maintenance_tickets"

// MaintenanceRepo implements outbound.MaintenanceRepository for MongoDB.
type MaintenanceRepo struct {
	col *mongo.Collection
}

// NewMaintenanceRepo creates a new MongoDB MaintenanceRepository.
func NewMaintenanceRepo(db *mongo.Database) *MaintenanceRepo {
	return &MaintenanceRepo{col: db.Collection(collectionMaintenance)}
}

// Insert persists a new maintenance ticket.
func (r *MaintenanceRepo) Insert(ctx context.Context, t *domain.MaintenanceTicket) error {
	if _, err := r.col.InsertOne(ctx, maintenanceFromDomain(t)); err != nil {
		return fmt.Errorf("mongodb.MaintenanceRepo.Insert: %w", err)
	}
	return nil
}

// FindByID retrieves a ticket by ID.
func (r *MaintenanceRepo) FindByID(ctx context.Context, id string) (*domain.MaintenanceTicket, error) {
	var doc maintenanceDoc
	if err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("mongodb.MaintenanceRepo.FindByID: %w", domain.ErrTicketNotFound)
		}
		return nil, fmt.Errorf("mongodb.MaintenanceRepo.FindByID: %w", err)
	}
	return maintenanceToDomain(doc), nil
}

// FindAll returns filtered maintenance tickets.
func (r *MaintenanceRepo) FindAll(ctx context.Context, f inbound.TicketFilter) ([]*domain.MaintenanceTicket, int64, error) {
	filter := bson.M{}
	if f.RoomID != "" {
		filter["room_id"] = f.RoomID
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
	if f.Priority != nil {
		filter["priority"] = string(*f.Priority)
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.MaintenanceRepo.FindAll: count: %w", err)
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
		return nil, 0, fmt.Errorf("mongodb.MaintenanceRepo.FindAll: find: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []maintenanceDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("mongodb.MaintenanceRepo.FindAll: decode: %w", err)
	}

	tickets := make([]*domain.MaintenanceTicket, 0, len(docs))
	for _, d := range docs {
		tickets = append(tickets, maintenanceToDomain(d))
	}
	return tickets, total, nil
}

// Update persists ticket state changes.
func (r *MaintenanceRepo) Update(ctx context.Context, t *domain.MaintenanceTicket) error {
	doc := maintenanceFromDomain(t)
	result, err := r.col.UpdateByID(ctx, t.ID, bson.M{"$set": bson.M{
		"status":      doc.Status,
		"notes":       doc.Notes,
		"resolved_at": doc.ResolvedAt,
		"updated_at":  doc.UpdatedAt,
	}})
	if err != nil {
		return fmt.Errorf("mongodb.MaintenanceRepo.Update: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("mongodb.MaintenanceRepo.Update: %w", domain.ErrTicketNotFound)
	}
	return nil
}

// CountOpenByOwner counts open/in-progress tickets for an owner.
func (r *MaintenanceRepo) CountOpenByOwner(ctx context.Context, ownerID string) (int64, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{
		"owner_id": ownerID,
		"status":   bson.M{"$in": bson.A{"open", "in_progress"}},
	})
	if err != nil {
		return 0, fmt.Errorf("mongodb.MaintenanceRepo.CountOpenByOwner: %w", err)
	}
	return count, nil
}
