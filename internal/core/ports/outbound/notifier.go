package outbound

import (
	"context"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2 --name=Notifier

// Notifier defines the outbound port for sending notifications.
type Notifier interface {
	SendWelcomeEmail(ctx context.Context, to domain.Email, name string) error
}
