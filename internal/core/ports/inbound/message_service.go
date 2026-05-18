package inbound

import (
	"context"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2 --name=MessageService

// SendMessageInput holds data to send a message.
type SendMessageInput struct {
	SenderID   string
	ReceiverID string
	Content    string
}

// MessageFilter defines query parameters for listing messages in a thread.
type MessageFilter struct {
	ThreadID string
	Page     int
	Limit    int
}

// MessageService defines the inbound port for simple owner-tenant messaging.
type MessageService interface {
	Send(ctx context.Context, in SendMessageInput) (*domain.Message, error)
	GetThread(ctx context.Context, userID1, userID2 string) (*domain.Thread, error)
	ListThreads(ctx context.Context, userID string, page, limit int) ([]*domain.Thread, int64, error)
	ListMessages(ctx context.Context, f MessageFilter) ([]*domain.Message, int64, error)
	MarkRead(ctx context.Context, threadID string, receiverID string) error
}
