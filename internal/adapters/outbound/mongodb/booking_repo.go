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

const collectionBookings = "bookings"

// BookingRepo implements outbound.BookingRepository for MongoDB.
type BookingRepo struct {
	col *mongo.Collection
}

// NewBookingRepo creates a new MongoDB BookingRepository.
func NewBookingRepo(db *mongo.Database) *BookingRepo {
	return &BookingRepo{col: db.Collection(collectionBookings)}
}

// Insert persists a new booking document.
func (r *BookingRepo) Insert(ctx context.Context, b *domain.Booking) error {
	if _, err := r.col.InsertOne(ctx, bookingFromDomain(b)); err != nil {
		return fmt.Errorf("mongodb.BookingRepo.Insert: %w", err)
	}
	return nil
}

// FindByID retrieves a booking by ID.
func (r *BookingRepo) FindByID(ctx context.Context, id string) (*domain.Booking, error) {
	var doc bookingDoc
	if err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("mongodb.BookingRepo.FindByID: %w", domain.ErrBookingNotFound)
		}
		return nil, fmt.Errorf("mongodb.BookingRepo.FindByID: %w", err)
	}
	return bookingToDomain(doc), nil
}

// FindAll returns filtered, paginated bookings.
func (r *BookingRepo) FindAll(ctx context.Context, f inbound.BookingFilter) ([]*domain.Booking, int64, error) {
	filter := bson.M{}
	if f.OwnerID != "" {
		filter["owner_id"] = f.OwnerID
	}
	if f.TenantID != "" {
		filter["tenant_id"] = f.TenantID
	}
	if f.RoomID != "" {
		filter["room_id"] = f.RoomID
	}
	if f.Status != nil {
		filter["status"] = string(*f.Status)
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.BookingRepo.FindAll: count: %w", err)
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
		return nil, 0, fmt.Errorf("mongodb.BookingRepo.FindAll: find: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []bookingDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("mongodb.BookingRepo.FindAll: decode: %w", err)
	}

	bookings := make([]*domain.Booking, 0, len(docs))
	for _, d := range docs {
		bookings = append(bookings, bookingToDomain(d))
	}
	return bookings, total, nil
}

// Update persists changes to an existing booking.
func (r *BookingRepo) Update(ctx context.Context, b *domain.Booking) error {
	doc := bookingFromDomain(b)
	result, err := r.col.UpdateByID(ctx, b.ID, bson.M{"$set": bson.M{
		"status":           doc.Status,
		"rejection_reason": doc.RejectionReason,
		"updated_at":       doc.UpdatedAt,
	}})
	if err != nil {
		return fmt.Errorf("mongodb.BookingRepo.Update: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("mongodb.BookingRepo.Update: %w", domain.ErrBookingNotFound)
	}
	return nil
}

// HasActiveBooking checks for an overlapping active booking for a room.
func (r *BookingRepo) HasActiveBooking(ctx context.Context, roomID string, start, end time.Time) (bool, error) {
	filter := bson.M{
		"room_id": roomID,
		"status":  bson.M{"$in": bson.A{"pending", "approved", "active"}},
		"start_date": bson.M{"$lt": end},
		"end_date":   bson.M{"$gt": start},
	}
	count, err := r.col.CountDocuments(ctx, filter, options.Count().SetLimit(1))
	if err != nil {
		return false, fmt.Errorf("mongodb.BookingRepo.HasActiveBooking: %w", err)
	}
	return count > 0, nil
}

// CountByStatus counts bookings for an owner filtered by status.
func (r *BookingRepo) CountByStatus(ctx context.Context, ownerID string, status domain.BookingStatus) (int64, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{"owner_id": ownerID, "status": string(status)})
	if err != nil {
		return 0, fmt.Errorf("mongodb.BookingRepo.CountByStatus: %w", err)
	}
	return count, nil
}
