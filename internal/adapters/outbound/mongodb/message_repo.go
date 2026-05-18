package mongodb

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
)

const (
	collectionMessages = "messages"
	collectionThreads  = "message_threads"
)

// MessageRepo implements outbound.MessageRepository for MongoDB.
type MessageRepo struct {
	msgCol    *mongo.Collection
	threadCol *mongo.Collection
}

// NewMessageRepo creates a new MongoDB MessageRepository.
func NewMessageRepo(db *mongo.Database) *MessageRepo {
	return &MessageRepo{
		msgCol:    db.Collection(collectionMessages),
		threadCol: db.Collection(collectionThreads),
	}
}

// InsertMessage persists a new message.
func (r *MessageRepo) InsertMessage(ctx context.Context, m *domain.Message) error {
	if _, err := r.msgCol.InsertOne(ctx, messageFromDomain(m)); err != nil {
		return fmt.Errorf("mongodb.MessageRepo.InsertMessage: %w", err)
	}
	return nil
}

// FindMessagesByThread returns paginated messages for a thread.
func (r *MessageRepo) FindMessagesByThread(ctx context.Context, f inbound.MessageFilter) ([]*domain.Message, int64, error) {
	filter := bson.M{"thread_id": f.ThreadID}

	total, err := r.msgCol.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.MessageRepo.FindMessagesByThread: count: %w", err)
	}

	skip := int64((f.Page - 1) * f.Limit)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(f.Limit))

	cursor, err := r.msgCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.MessageRepo.FindMessagesByThread: find: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []messageDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("mongodb.MessageRepo.FindMessagesByThread: decode: %w", err)
	}

	msgs := make([]*domain.Message, 0, len(docs))
	for _, d := range docs {
		msgs = append(msgs, messageToDomain(d))
	}
	return msgs, total, nil
}

// MarkThreadRead marks all unread messages in a thread as read for a receiver.
func (r *MessageRepo) MarkThreadRead(ctx context.Context, threadID, receiverID string) error {
	_, err := r.msgCol.UpdateMany(ctx,
		bson.M{"thread_id": threadID, "receiver_id": receiverID, "read": false},
		bson.M{"$set": bson.M{"read": true}},
	)
	if err != nil {
		return fmt.Errorf("mongodb.MessageRepo.MarkThreadRead: %w", err)
	}
	return nil
}

// UpsertThread creates or updates a message thread.
func (r *MessageRepo) UpsertThread(ctx context.Context, t *domain.Thread) error {
	doc := threadFromDomain(t)

	// Sort participants for consistent lookup.
	participants := make([]string, len(t.Participants))
	copy(participants, t.Participants)
	sort.Strings(participants)
	doc.Participants = participants

	filter := bson.M{"_id": t.ID}
	update := bson.M{"$set": bson.M{
		"participants": doc.Participants,
		"last_message": doc.LastMessage,
		"last_at":      doc.LastAt,
		"updated_at":   doc.UpdatedAt,
	}, "$setOnInsert": bson.M{
		"created_at": doc.CreatedAt,
	}}

	opts := options.UpdateOne().SetUpsert(true)
	if _, err := r.threadCol.UpdateOne(ctx, filter, update, opts); err != nil {
		return fmt.Errorf("mongodb.MessageRepo.UpsertThread: %w", err)
	}
	return nil
}

// FindThread finds an existing thread between two users.
func (r *MessageRepo) FindThread(ctx context.Context, userID1, userID2 string) (*domain.Thread, error) {
	// Sort participants for consistent lookup.
	p := []string{userID1, userID2}
	sort.Strings(p)

	var doc threadDoc
	err := r.threadCol.FindOne(ctx, bson.M{"participants": bson.M{"$all": p}}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("mongodb.MessageRepo.FindThread: %w", domain.ErrMessageNotFound)
		}
		return nil, fmt.Errorf("mongodb.MessageRepo.FindThread: %w", err)
	}
	return threadToDomain(doc), nil
}

// FindThreadByID retrieves a thread by its ID.
func (r *MessageRepo) FindThreadByID(ctx context.Context, threadID string) (*domain.Thread, error) {
	var doc threadDoc
	if err := r.threadCol.FindOne(ctx, bson.M{"_id": threadID}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("mongodb.MessageRepo.FindThreadByID: %w", domain.ErrMessageNotFound)
		}
		return nil, fmt.Errorf("mongodb.MessageRepo.FindThreadByID: %w", err)
	}
	return threadToDomain(doc), nil
}

// FindThreadsByUser returns all threads involving a user, ordered by last message time.
func (r *MessageRepo) FindThreadsByUser(ctx context.Context, userID string, page, limit int) ([]*domain.Thread, int64, error) {
	filter := bson.M{"participants": userID}

	total, err := r.threadCol.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.MessageRepo.FindThreadsByUser: count: %w", err)
	}

	skip := int64((page - 1) * limit)
	opts := options.Find().
		SetSort(bson.D{{Key: "last_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(int64(limit))

	cursor, err := r.threadCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.MessageRepo.FindThreadsByUser: find: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []threadDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("mongodb.MessageRepo.FindThreadsByUser: decode: %w", err)
	}

	threads := make([]*domain.Thread, 0, len(docs))
	for _, d := range docs {
		threads = append(threads, threadToDomain(d))
	}
	return threads, total, nil
}

// CountUnreadByUser counts all unread messages for a user across all threads.
func (r *MessageRepo) CountUnreadByUser(ctx context.Context, userID string) (int64, error) {
	count, err := r.msgCol.CountDocuments(ctx, bson.M{
		"receiver_id": userID,
		"read":        false,
	})
	if err != nil {
		return 0, fmt.Errorf("mongodb.MessageRepo.CountUnreadByUser: %w", err)
	}
	return count, nil
}
