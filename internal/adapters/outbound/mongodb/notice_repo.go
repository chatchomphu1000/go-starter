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

const collectionNotices = "notices"

// NoticeRepo implements outbound.NoticeRepository for MongoDB.
type NoticeRepo struct {
	col *mongo.Collection
}

// NewNoticeRepo creates a new MongoDB NoticeRepository.
func NewNoticeRepo(db *mongo.Database) *NoticeRepo {
	return &NoticeRepo{col: db.Collection(collectionNotices)}
}

// Insert persists a new notice.
func (r *NoticeRepo) Insert(ctx context.Context, n *domain.Notice) error {
	if _, err := r.col.InsertOne(ctx, noticeFromDomain(n)); err != nil {
		return fmt.Errorf("mongodb.NoticeRepo.Insert: %w", err)
	}
	return nil
}

// FindByID retrieves a notice by ID.
func (r *NoticeRepo) FindByID(ctx context.Context, id string) (*domain.Notice, error) {
	var doc noticeDoc
	if err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("mongodb.NoticeRepo.FindByID: %w", domain.ErrNoticeNotFound)
		}
		return nil, fmt.Errorf("mongodb.NoticeRepo.FindByID: %w", err)
	}
	return noticeToDomain(doc), nil
}

// FindAll returns filtered notices.
func (r *NoticeRepo) FindAll(ctx context.Context, f inbound.NoticeFilter) ([]*domain.Notice, int64, error) {
	filter := bson.M{}
	if f.OwnerID != "" {
		filter["owner_id"] = f.OwnerID
	}
	if f.Type != nil {
		filter["type"] = string(*f.Type)
	}
	if f.PinnedOnly {
		filter["pinned"] = true
	}
	if f.ActiveOnly {
		now := time.Now().UTC()
		filter["$or"] = bson.A{
			bson.M{"expires_at": nil},
			bson.M{"expires_at": bson.M{"$gt": now}},
		}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.NoticeRepo.FindAll: count: %w", err)
	}

	sortOrder := -1
	if !f.SortDesc {
		sortOrder = 1
	}
	skip := int64((f.Page - 1) * f.Limit)

	// Pinned notices first, then by date.
	opts := options.Find().
		SetSort(bson.D{
			{Key: "pinned", Value: -1},
			{Key: "created_at", Value: sortOrder},
		}).
		SetSkip(skip).
		SetLimit(int64(f.Limit))

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.NoticeRepo.FindAll: find: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []noticeDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("mongodb.NoticeRepo.FindAll: decode: %w", err)
	}

	notices := make([]*domain.Notice, 0, len(docs))
	for _, d := range docs {
		notices = append(notices, noticeToDomain(d))
	}
	return notices, total, nil
}

// Update persists notice changes.
func (r *NoticeRepo) Update(ctx context.Context, n *domain.Notice) error {
	doc := noticeFromDomain(n)
	result, err := r.col.UpdateByID(ctx, n.ID, bson.M{"$set": bson.M{
		"title":      doc.Title,
		"content":    doc.Content,
		"type":       doc.Type,
		"pinned":     doc.Pinned,
		"expires_at": doc.ExpiresAt,
		"updated_at": doc.UpdatedAt,
	}})
	if err != nil {
		return fmt.Errorf("mongodb.NoticeRepo.Update: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("mongodb.NoticeRepo.Update: %w", domain.ErrNoticeNotFound)
	}
	return nil
}

// Delete removes a notice by ID.
func (r *NoticeRepo) Delete(ctx context.Context, id string) error {
	result, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("mongodb.NoticeRepo.Delete: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("mongodb.NoticeRepo.Delete: %w", domain.ErrNoticeNotFound)
	}
	return nil
}
