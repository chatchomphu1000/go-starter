package inbound

import (
	"context"
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2 --name=NoticeService

// CreateNoticeInput holds data to post an announcement.
type CreateNoticeInput struct {
	OwnerID   string
	Title     string
	Content   string
	Type      string
	Pinned    bool
	ExpiresAt *time.Time
}

// UpdateNoticeInput holds updatable notice fields.
type UpdateNoticeInput struct {
	Title     *string
	Content   *string
	Type      *string
	Pinned    *bool
	ExpiresAt *time.Time
}

// NoticeFilter defines query parameters for listing notices.
type NoticeFilter struct {
	OwnerID    string
	Type       *domain.NoticeType
	PinnedOnly bool
	ActiveOnly bool // not expired
	Page       int
	Limit      int
	SortDesc   bool
}

// NoticeService defines the inbound port for notice management.
type NoticeService interface {
	Create(ctx context.Context, in CreateNoticeInput) (*domain.Notice, error)
	GetByID(ctx context.Context, id string) (*domain.Notice, error)
	List(ctx context.Context, f NoticeFilter) ([]*domain.Notice, int64, error)
	Update(ctx context.Context, id string, ownerID string, in UpdateNoticeInput) (*domain.Notice, error)
	Delete(ctx context.Context, id string, ownerID string) error
}
