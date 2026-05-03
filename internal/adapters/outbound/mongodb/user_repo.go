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

const collectionUsers = "users"

// UserRepo implements outbound.UserRepository for MongoDB.
type UserRepo struct {
	col *mongo.Collection
}

// NewUserRepo creates a new MongoDB UserRepository.
func NewUserRepo(db *mongo.Database) *UserRepo {
	return &UserRepo{col: db.Collection(collectionUsers)}
}

// Insert inserts a new user document.
func (r *UserRepo) Insert(ctx context.Context, u *domain.User) error {
	doc := fromDomain(u)
	_, err := r.col.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("mongodb.Insert: %w", domain.ErrEmailAlreadyExists)
		}
		return fmt.Errorf("mongodb.Insert: %w", err)
	}
	return nil
}

// FindByID finds a user by their string ID.
func (r *UserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var doc userDoc
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("mongodb.FindByID: %w", domain.ErrUserNotFound)
		}
		return nil, fmt.Errorf("mongodb.FindByID: %w", err)
	}

	user, err := toDomain(doc)
	if err != nil {
		return nil, fmt.Errorf("mongodb.FindByID: %w", err)
	}

	return user, nil
}

// FindByEmail finds a user by their email.
func (r *UserRepo) FindByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	var doc userDoc
	err := r.col.FindOne(ctx, bson.M{"email": email.String()}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("mongodb.FindByEmail: %w", domain.ErrUserNotFound)
		}
		return nil, fmt.Errorf("mongodb.FindByEmail: %w", err)
	}

	user, err := toDomain(doc)
	if err != nil {
		return nil, fmt.Errorf("mongodb.FindByEmail: %w", err)
	}

	return user, nil
}

// FindAll returns a filtered, paginated, and sorted list of users along with the total count.
func (r *UserRepo) FindAll(ctx context.Context, f inbound.ListFilter) ([]*domain.User, int64, error) {
	filter := bson.M{}

	if f.Role != nil {
		filter["role"] = f.Role.String()
	}
	if f.Active != nil {
		filter["active"] = *f.Active
	}
	if f.Search != "" {
		escaped := regexp.QuoteMeta(f.Search)
		regex := bson.M{"$regex": escaped, "$options": "i"}
		filter["$or"] = bson.A{
			bson.M{"name": regex},
			bson.M{"email": regex},
		}
	}

	total, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.FindAll: count: %w", err)
	}

	sortField := "created_at"
	if f.SortBy == "name" || f.SortBy == "email" {
		sortField = f.SortBy
	}

	sortOrder := 1
	if f.SortDesc {
		sortOrder = -1
	}

	skip := int64((f.Page - 1) * f.Limit)

	findOpts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: sortOrder}}).
		SetSkip(skip).
		SetLimit(int64(f.Limit))

	cursor, err := r.col.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("mongodb.FindAll: find: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []userDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, fmt.Errorf("mongodb.FindAll: decode: %w", err)
	}

	users := make([]*domain.User, 0, len(docs))
	for _, doc := range docs {
		u, err := toDomain(doc)
		if err != nil {
			return nil, 0, fmt.Errorf("mongodb.FindAll: toDomain: %w", err)
		}
		users = append(users, u)
	}

	return users, total, nil
}

// Update updates an existing user document.
func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	doc := fromDomain(u)
	result, err := r.col.UpdateByID(ctx, u.ID, bson.M{
		"$set": bson.M{
			"name":            doc.Name,
			"email":           doc.Email,
			"hashed_password": doc.HashedPassword,
			"role":            doc.Role,
			"active":          doc.Active,
			"updated_at":      doc.UpdatedAt,
		},
	})
	if err != nil {
		return fmt.Errorf("mongodb.Update: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("mongodb.Update: %w", domain.ErrUserNotFound)
	}
	return nil
}

// Delete removes a user document by ID.
func (r *UserRepo) Delete(ctx context.Context, id string) error {
	result, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("mongodb.Delete: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("mongodb.Delete: %w", domain.ErrUserNotFound)
	}
	return nil
}

// ExistsByEmail checks if a user with the given email exists.
func (r *UserRepo) ExistsByEmail(ctx context.Context, email domain.Email) (bool, error) {
	count, err := r.col.CountDocuments(ctx, bson.M{"email": email.String()}, options.Count().SetLimit(1))
	if err != nil {
		return false, fmt.Errorf("mongodb.ExistsByEmail: %w", err)
	}
	return count > 0, nil
}
