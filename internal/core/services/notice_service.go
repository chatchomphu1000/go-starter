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

type noticeService struct {
	repo  outbound.NoticeRepository
	clock outbound.Clock
	ids   outbound.IDGenerator
	log   logger.Logger
}

// NewNoticeService creates a new NoticeService.
func NewNoticeService(
	repo outbound.NoticeRepository,
	clock outbound.Clock,
	ids outbound.IDGenerator,
	log logger.Logger,
) inbound.NoticeService {
	return &noticeService{repo: repo, clock: clock, ids: ids, log: log}
}

// Create posts a new notice.
func (s *noticeService) Create(ctx context.Context, in inbound.CreateNoticeInput) (*domain.Notice, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("noticeService.Create: %w", err)
	}

	n, err := domain.NewNotice(
		s.ids.New(), in.OwnerID, in.Title, in.Content,
		domain.NoticeType(in.Type), in.Pinned, in.ExpiresAt, s.clock.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("noticeService.Create: %w", apperrors.BadRequest(err.Error(), err))
	}

	if err := s.repo.Insert(ctx, n); err != nil {
		return nil, fmt.Errorf("noticeService.Create: %w", err)
	}

	s.log.Info("notice created", zap.String("notice_id", n.ID))
	return n, nil
}

// GetByID retrieves a notice by ID.
func (s *noticeService) GetByID(ctx context.Context, id string) (*domain.Notice, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("noticeService.GetByID: %w", err)
	}
	n, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("noticeService.GetByID: %w", err)
	}
	return n, nil
}

// List returns filtered notices.
func (s *noticeService) List(ctx context.Context, f inbound.NoticeFilter) ([]*domain.Notice, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, fmt.Errorf("noticeService.List: %w", err)
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	return s.repo.FindAll(ctx, f)
}

// Update modifies a notice — only the owner may update their own notices.
func (s *noticeService) Update(ctx context.Context, id, ownerID string, in inbound.UpdateNoticeInput) (*domain.Notice, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("noticeService.Update: %w", err)
	}

	n, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("noticeService.Update: %w", err)
	}
	if n.OwnerID != ownerID {
		return nil, fmt.Errorf("noticeService.Update: %w", domain.ErrUnauthorizedAccess)
	}

	if in.Title != nil {
		n.Title = *in.Title
	}
	if in.Content != nil {
		n.Content = *in.Content
	}
	if in.Type != nil {
		n.Type = domain.NoticeType(*in.Type)
	}
	if in.Pinned != nil {
		n.Pinned = *in.Pinned
	}
	if in.ExpiresAt != nil {
		n.ExpiresAt = in.ExpiresAt
	}
	n.Touch(s.clock.Now())

	if err := s.repo.Update(ctx, n); err != nil {
		return nil, fmt.Errorf("noticeService.Update: %w", err)
	}
	s.log.Info("notice updated", zap.String("notice_id", id))
	return n, nil
}

// Delete removes a notice — only the owner may delete their own notices.
func (s *noticeService) Delete(ctx context.Context, id, ownerID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("noticeService.Delete: %w", err)
	}
	n, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("noticeService.Delete: %w", err)
	}
	if n.OwnerID != ownerID {
		return fmt.Errorf("noticeService.Delete: %w", domain.ErrUnauthorizedAccess)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("noticeService.Delete: %w", err)
	}
	s.log.Info("notice deleted", zap.String("notice_id", id))
	return nil
}
