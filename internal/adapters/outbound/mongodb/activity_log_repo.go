package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	outbound2 "github.com/chatchomphu1000/go-starter/internal/core/ports/outbound"
)

const collectionActivityLogs = "activity_logs"

type activityLogDoc struct {
	ID         string    `bson:"_id"`
	UserID     string    `bson:"user_id"`
	Action     string    `bson:"action"`
	Resource   string    `bson:"resource"`
	ResourceID string    `bson:"resource_id"`
	IPAddress  string    `bson:"ip_address"`
	UserAgent  string    `bson:"user_agent"`
	CreatedAt  time.Time `bson:"created_at"`
}

// ActivityLogRepo implements outbound.ActivityLogRepository for MongoDB.
type ActivityLogRepo struct {
	col *mongo.Collection
}

// NewActivityLogRepo creates a new MongoDB ActivityLogRepository.
func NewActivityLogRepo(db *mongo.Database) *ActivityLogRepo {
	return &ActivityLogRepo{col: db.Collection(collectionActivityLogs)}
}

// Insert appends a new activity log entry.
func (r *ActivityLogRepo) Insert(ctx context.Context, l *domain.ActivityLog) error {
	doc := activityLogDoc{
		ID:         l.ID,
		UserID:     l.UserID,
		Action:     l.Action,
		Resource:   l.Resource,
		ResourceID: l.ResourceID,
		IPAddress:  l.IPAddress,
		UserAgent:  l.UserAgent,
		CreatedAt:  l.CreatedAt,
	}
	if _, err := r.col.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("mongodb.ActivityLogRepo.Insert: %w", err)
	}
	return nil
}

// FindAll returns filtered activity logs.
func (r *ActivityLogRepo) FindAll(ctx context.Context, f outbound2.ActivityLogFilter) ([]*domain.ActivityLog, int64, error) {
	filter := bson.M{}
	if f.UserID != "" {
		filter["user_id"] = f.UserID
	}
	if f.Resource != "" {
		filter["resource"] = f.Resource
	}
	if f.ResourceID != "" {
		filter["resource_id"] = f.ResourceID
	}
	if f.Action != "" {
		filter["action"] = f.Action
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.ActivityLogRepo.FindAll: count: %w", err)
	}

	skip := int64((f.Page - 1) * f.Limit)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(f.Limit))

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.ActivityLogRepo.FindAll: find: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []activityLogDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("mongodb.ActivityLogRepo.FindAll: decode: %w", err)
	}

	logs := make([]*domain.ActivityLog, 0, len(docs))
	for _, d := range docs {
		logs = append(logs, &domain.ActivityLog{
			ID:         d.ID,
			UserID:     d.UserID,
			Action:     d.Action,
			Resource:   d.Resource,
			ResourceID: d.ResourceID,
			IPAddress:  d.IPAddress,
			UserAgent:  d.UserAgent,
			CreatedAt:  d.CreatedAt,
		})
	}
	return logs, total, nil
}
