// Package httpclient provides outbound HTTP client adapters.
package httpclient

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/pkg/apperrors"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// NotifierConfig holds configuration for the outbound notifier HTTP client.
type NotifierConfig struct {
	BaseURL string
	Timeout time.Duration
	Retry   int
}

// RestyNotifier implements outbound.Notifier using go-resty.
type RestyNotifier struct {
	client *resty.Client
	log    logger.Logger
}

// NewNotifier creates a new RestyNotifier.
func NewNotifier(cfg NotifierConfig, log logger.Logger) *RestyNotifier {
	client := resty.New().
		SetBaseURL(cfg.BaseURL).
		SetTimeout(cfg.Timeout).
		SetRetryCount(cfg.Retry).
		SetRetryWaitTime(500 * time.Millisecond).
		SetRetryMaxWaitTime(5 * time.Second).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			if err != nil {
				return true
			}
			return r.StatusCode() >= 500
		})

	client.OnBeforeRequest(func(_ *resty.Client, req *resty.Request) error {
		log.Debug("outbound request",
			zap.String("method", req.Method),
			zap.String("url", req.URL),
		)
		return nil
	})

	client.OnAfterResponse(func(_ *resty.Client, resp *resty.Response) error {
		log.Debug("outbound response",
			zap.Int("status", resp.StatusCode()),
			zap.Int64("latency_ms", resp.Time().Milliseconds()),
		)
		return nil
	})

	return &RestyNotifier{client: client, log: log}
}

// SendWelcomeEmail sends a welcome email notification via the external notification service.
func (n *RestyNotifier) SendWelcomeEmail(ctx context.Context, to domain.Email, name string) error {
	reqID := ""
	if rid, ok := ctx.Value(logger.RequestIDKey).(string); ok {
		reqID = rid
	}

	resp, err := n.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Request-ID", reqID).
		SetHeader("Idempotency-Key", uuid.New().String()).
		SetBody(map[string]string{
			"to":   to.String(),
			"name": name,
			"type": "welcome",
		}).
		Post("/api/v1/notifications/email")

	if err != nil {
		return fmt.Errorf("notifier.SendWelcomeEmail: %w", apperrors.Internal("notification service unreachable", err))
	}

	if resp.IsError() {
		return fmt.Errorf("notifier.SendWelcomeEmail: %w",
			apperrors.Wrap(apperrors.CodeNotifierFailed,
				fmt.Sprintf("notification service returned status %d", resp.StatusCode()), nil))
	}

	return nil
}

// CloseIdleConnections closes idle connections in the underlying HTTP transport.
func (n *RestyNotifier) CloseIdleConnections() {
	transport := n.client.GetClient().Transport
	if t, ok := transport.(interface{ CloseIdleConnections() }); ok {
		t.CloseIdleConnections()
	}
}
