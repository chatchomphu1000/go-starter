package mongodb

import (
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

// userDoc is the internal BSON model for MongoDB user documents.
// Never exported outside this package.
type userDoc struct {
	ID             string    `bson:"_id"`
	Name           string    `bson:"name"`
	Email          string    `bson:"email"`
	HashedPassword string    `bson:"hashed_password"`
	Role           string    `bson:"role"`
	Active         bool      `bson:"active"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
}

// toDomain converts a userDoc to a domain.User.
func toDomain(d userDoc) (*domain.User, error) {
	email, err := domain.NewEmail(d.Email)
	if err != nil {
		return nil, err
	}

	role, err := domain.ParseRole(d.Role)
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:             d.ID,
		Name:           d.Name,
		Email:          email,
		HashedPassword: d.HashedPassword,
		Role:           role,
		Active:         d.Active,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}, nil
}

// fromDomain converts a domain.User to a userDoc.
func fromDomain(u *domain.User) userDoc {
	return userDoc{
		ID:             u.ID,
		Name:           u.Name,
		Email:          u.Email.String(),
		HashedPassword: u.HashedPassword,
		Role:           u.Role.String(),
		Active:         u.Active,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}
