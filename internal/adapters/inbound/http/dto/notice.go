package dto

import (
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

// CreateNoticeRequest is the request body for creating a notice.
type CreateNoticeRequest struct {
	Title     string  `json:"title" validate:"required,min=3,max=200"`
	Content   string  `json:"content" validate:"required,min=5,max=5000"`
	Type      string  `json:"type" validate:"required,oneof=general payment maintenance emergency event"`
	Pinned    bool    `json:"pinned"`
	ExpiresAt *string `json:"expires_at"` // YYYY-MM-DD optional
}

// UpdateNoticeRequest is the request body for updating a notice.
type UpdateNoticeRequest struct {
	Title     *string `json:"title" validate:"omitempty,min=3,max=200"`
	Content   *string `json:"content" validate:"omitempty,min=5,max=5000"`
	Type      *string `json:"type" validate:"omitempty,oneof=general payment maintenance emergency event"`
	Pinned    *bool   `json:"pinned"`
	ExpiresAt *string `json:"expires_at"` // YYYY-MM-DD or null to clear
}

// NoticeResponse is the API representation of a notice.
type NoticeResponse struct {
	ID        string  `json:"id"`
	OwnerID   string  `json:"owner_id"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	Type      string  `json:"type"`
	Pinned    bool    `json:"pinned"`
	ExpiresAt *string `json:"expires_at,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// NoticeListResponse wraps a paginated list of notices.
type NoticeListResponse struct {
	Data  []NoticeResponse `json:"data"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

// ToNoticeResponse maps a domain Notice to NoticeResponse.
func ToNoticeResponse(n *domain.Notice) NoticeResponse {
	r := NoticeResponse{
		ID:        n.ID,
		OwnerID:   n.OwnerID,
		Title:     n.Title,
		Content:   n.Content,
		Type:      string(n.Type),
		Pinned:    n.Pinned,
		CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: n.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if n.ExpiresAt != nil {
		s := n.ExpiresAt.Format("2006-01-02")
		r.ExpiresAt = &s
	}
	return r
}

// ToNoticeListResponse builds a paginated notice list response.
func ToNoticeListResponse(notices []*domain.Notice, total int64, page, limit int) NoticeListResponse {
	data := make([]NoticeResponse, 0, len(notices))
	for _, n := range notices {
		data = append(data, ToNoticeResponse(n))
	}
	return NoticeListResponse{Data: data, Total: total, Page: page, Limit: limit}
}

// ParseExpiresAt parses an optional date string to a time pointer.
func ParseExpiresAt(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}
