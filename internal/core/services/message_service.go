package services

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/outbound"
	"github.com/chatchomphu1000/go-starter/pkg/apperrors"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

type messageService struct {
	repo  outbound.MessageRepository
	clock outbound.Clock
	ids   outbound.IDGenerator
	log   logger.Logger
}

// NewMessageService creates a new MessageService.
func NewMessageService(
	repo outbound.MessageRepository,
	clock outbound.Clock,
	ids outbound.IDGenerator,
	log logger.Logger,
) inbound.MessageService {
	return &messageService{repo: repo, clock: clock, ids: ids, log: log}
}

// Send delivers a message and creates/updates the conversation thread.
func (s *messageService) Send(ctx context.Context, in inbound.SendMessageInput) (*domain.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("messageService.Send: %w", err)
	}

	now := s.clock.Now()

	// Ensure thread exists (upsert).
	thread, err := s.repo.FindThread(ctx, in.SenderID, in.ReceiverID)
	if err != nil {
		// Create a new thread.
		thread, err = domain.NewThread(s.ids.New(), []string{in.SenderID, in.ReceiverID}, now)
		if err != nil {
			return nil, fmt.Errorf("messageService.Send: %w", apperrors.BadRequest(err.Error(), err))
		}
	}

	msg, err := domain.NewMessage(s.ids.New(), thread.ID, in.SenderID, in.ReceiverID, in.Content, now)
	if err != nil {
		return nil, fmt.Errorf("messageService.Send: %w", apperrors.BadRequest(err.Error(), err))
	}

	if err := s.repo.InsertMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("messageService.Send: %w", err)
	}

	thread.UpdateLastMessage(in.Content, now)
	if err := s.repo.UpsertThread(ctx, thread); err != nil {
		s.log.Error("failed to upsert thread", zap.String("thread_id", thread.ID), zap.Error(err))
	}

	return msg, nil
}

// GetThread retrieves or creates a thread between two users.
func (s *messageService) GetThread(ctx context.Context, userID1, userID2 string) (*domain.Thread, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("messageService.GetThread: %w", err)
	}
	thread, err := s.repo.FindThread(ctx, userID1, userID2)
	if err != nil {
		return nil, fmt.Errorf("messageService.GetThread: %w", err)
	}
	return thread, nil
}

// ListThreads returns all conversation threads for a user.
func (s *messageService) ListThreads(ctx context.Context, userID string, page, limit int) ([]*domain.Thread, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, fmt.Errorf("messageService.ListThreads: %w", err)
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.FindThreadsByUser(ctx, userID, page, limit)
}

// ListMessages returns messages within a thread.
func (s *messageService) ListMessages(ctx context.Context, f inbound.MessageFilter) ([]*domain.Message, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, fmt.Errorf("messageService.ListMessages: %w", err)
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 50
	}
	return s.repo.FindMessagesByThread(ctx, f)
}

// MarkRead marks all messages in a thread as read for the receiver.
func (s *messageService) MarkRead(ctx context.Context, threadID, receiverID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("messageService.MarkRead: %w", err)
	}
	if err := s.repo.MarkThreadRead(ctx, threadID, receiverID); err != nil {
		return fmt.Errorf("messageService.MarkRead: %w", err)
	}
	return nil
}
