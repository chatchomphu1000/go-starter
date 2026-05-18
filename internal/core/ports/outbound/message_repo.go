package outbound

import (
	"context"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
)

//go:generate go run github.com/vektra/mockery/v2 --name=MessageRepository

// MessageRepository defines the outbound port for message and thread persistence.
type MessageRepository interface {
	InsertMessage(ctx context.Context, m *domain.Message) error
	FindMessagesByThread(ctx context.Context, f inbound.MessageFilter) ([]*domain.Message, int64, error)
	MarkThreadRead(ctx context.Context, threadID, receiverID string) error

	UpsertThread(ctx context.Context, t *domain.Thread) error
	FindThread(ctx context.Context, userID1, userID2 string) (*domain.Thread, error)
	FindThreadByID(ctx context.Context, threadID string) (*domain.Thread, error)
	FindThreadsByUser(ctx context.Context, userID string, page, limit int) ([]*domain.Thread, int64, error)
	CountUnreadByUser(ctx context.Context, userID string) (int64, error)
}
