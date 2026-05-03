package outbound

import (
	"context"
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2 --name=TokenIssuer

// TokenIssuer abstracts JWT token issuance and verification.
type TokenIssuer interface {
	Issue(ctx context.Context, userID string, role domain.Role) (token string, expiresAt time.Time, err error)
	Verify(ctx context.Context, token string) (userID string, role domain.Role, err error)
}
