package mongodb

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
)

const collectionRooms = "rooms"

// RoomRepo implements outbound.RoomRepository for MongoDB.
type RoomRepo struct {
	col *mongo.Collection
}

// NewRoomRepo creates a new MongoDB RoomRepository.
func NewRoomRepo(db *mongo.Database) *RoomRepo {
	return &RoomRepo{col: db.Collection(collectionRooms)}
}

// Insert persists a new room document.
func (r *RoomRepo) Insert(ctx context.Context, room *domain.Room) error {
	doc := roomFromDomain(room)
	if _, err := r.col.InsertOne(ctx, doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("mongodb.RoomRepo.Insert: %w", domain.ErrRoomNumberExists)
		}
		return fmt.Errorf("mongodb.RoomRepo.Insert: %w", err)
	}
	return nil
}

// FindByID retrieves a room by its string ID.
func (r *RoomRepo) FindByID(ctx context.Context, id string) (*domain.Room, error) {
	var doc roomDoc
	if err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("mongodb.RoomRepo.FindByID: %w", domain.ErrRoomNotFound)
		}
		return nil, fmt.Errorf("mongodb.RoomRepo.FindByID: %w", err)
	}
	return roomToDomain(doc), nil
}

// FindAll returns filtered and paginated rooms.
func (r *RoomRepo) FindAll(ctx context.Context, f inbound.RoomFilter) ([]*domain.Room, int64, error) {
	filter := bson.M{}
	if f.OwnerID != "" {
		filter["owner_id"] = f.OwnerID
	}
	if f.Status != nil {
		filter["status"] = string(*f.Status)
	}
	if f.Type != nil {
		filter["type"] = string(*f.Type)
	}
	if f.MinPrice != nil {
		filter["rent_price"] = bson.M{"$gte": *f.MinPrice}
	}
	if f.MaxPrice != nil {
		if _, ok := filter["rent_price"]; ok {
			filter["rent_price"].(bson.M)["$lte"] = *f.MaxPrice
		} else {
			filter["rent_price"] = bson.M{"$lte": *f.MaxPrice}
		}
	}
	if f.Search != "" {
		escaped := regexp.QuoteMeta(f.Search)
		regex := bson.M{"$regex": escaped, "$options": "i"}
		filter["$or"] = bson.A{
			bson.M{"name": regex},
			bson.M{"number": regex},
			bson.M{"description": regex},
		}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.RoomRepo.FindAll: count: %w", err)
	}

	sortField := "created_at"
	if f.SortBy == "name" || f.SortBy == "rent_price" || f.SortBy == "floor" {
		sortField = f.SortBy
	}
	sortOrder := 1
	if f.SortDesc {
		sortOrder = -1
	}

	skip := int64((f.Page - 1) * f.Limit)
	opts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: sortOrder}}).
		SetSkip(skip).
		SetLimit(int64(f.Limit))

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.RoomRepo.FindAll: find: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []roomDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("mongodb.RoomRepo.FindAll: decode: %w", err)
	}

	rooms := make([]*domain.Room, 0, len(docs))
	for _, d := range docs {
		rooms = append(rooms, roomToDomain(d))
	}
	return rooms, total, nil
}

// Update persists changes to an existing room.
func (r *RoomRepo) Update(ctx context.Context, room *domain.Room) error {
	doc := roomFromDomain(room)
	result, err := r.col.UpdateByID(ctx, room.ID, bson.M{"$set": bson.M{
		"owner_id":    doc.OwnerID,
		"number":      doc.Number,
		"name":        doc.Name,
		"type":        doc.Type,
		"floor":       doc.Floor,
		"size_sqm":    doc.SizeSqm,
		"rent_price":  doc.RentPrice,
		"deposit":     doc.Deposit,
		"status":      doc.Status,
		"amenities":   doc.Amenities,
		"photos":      doc.Photos,
		"description": doc.Description,
		"updated_at":  doc.UpdatedAt,
	}})
	if err != nil {
		return fmt.Errorf("mongodb.RoomRepo.Update: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("mongodb.RoomRepo.Update: %w", domain.ErrRoomNotFound)
	}
	return nil
}

// Delete removes a room by ID.
func (r *RoomRepo) Delete(ctx context.Context, id string) error {
	result, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("mongodb.RoomRepo.Delete: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("mongodb.RoomRepo.Delete: %w", domain.ErrRoomNotFound)
	}
	return nil
}

// ExistsByNumber checks if a room number already exists for an owner.
func (r *RoomRepo) ExistsByNumber(ctx context.Context, ownerID, number string) (bool, error) {
	count, err := r.col.CountDocuments(ctx,
		bson.M{"owner_id": ownerID, "number": number},
		options.Count().SetLimit(1),
	)
	if err != nil {
		return false, fmt.Errorf("mongodb.RoomRepo.ExistsByNumber: %w", err)
	}
	return count > 0, nil
}
