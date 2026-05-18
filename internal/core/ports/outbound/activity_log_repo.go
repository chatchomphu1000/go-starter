package outbound

import (
	"context"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2 --name=ActivityLogRepository

// ActivityLogFilter defines query parameters for listing audit logs.
type ActivityLogFilter struct {
	UserID     string
	Resource   string
	ResourceID string
	Action     string
	Page       int
	Limit      int
}

// ActivityLogRepository defines the outbound port for audit-trail persistence.
type ActivityLogRepository interface {
	Insert(ctx context.Context, l *domain.ActivityLog) error
	FindAll(ctx context.Context, f ActivityLogFilter) ([]*domain.ActivityLog, int64, error)
}
