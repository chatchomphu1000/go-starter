package outbound

import (
	"context"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
)

//go:generate go run github.com/vektra/mockery/v2 --name=NoticeRepository

// NoticeRepository defines the outbound port for notice persistence.
type NoticeRepository interface {
	Insert(ctx context.Context, n *domain.Notice) error
	FindByID(ctx context.Context, id string) (*domain.Notice, error)
	FindAll(ctx context.Context, f inbound.NoticeFilter) ([]*domain.Notice, int64, error)
	Update(ctx context.Context, n *domain.Notice) error
	Delete(ctx context.Context, id string) error
}
