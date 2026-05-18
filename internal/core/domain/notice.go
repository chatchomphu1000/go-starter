package domain

import (
	"fmt"
	"time"
)

// NoticeType categorises the announcement.
type NoticeType string

// Notice type constants.
const (
	NoticeTypeGeneral     NoticeType = "general"
	NoticeTypePayment     NoticeType = "payment"
	NoticeTypeMaintenance NoticeType = "maintenance"
	NoticeTypeEmergency   NoticeType = "emergency"
	NoticeTypeEvent       NoticeType = "event"
)

// Notice is a bulletin-board announcement posted by an owner or admin.
type Notice struct {
	ID        string
	OwnerID   string
	Title     string
	Content   string
	Type      NoticeType
	Pinned    bool
	ExpiresAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewNotice creates a validated Notice entity.
func NewNotice(
	id, ownerID, title, content string,
	noticeType NoticeType,
	pinned bool,
	expiresAt *time.Time,
	now time.Time,
) (*Notice, error) {
	n := &Notice{
		ID:        id,
		OwnerID:   ownerID,
		Title:     title,
		Content:   content,
		Type:      noticeType,
		Pinned:    pinned,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := n.Validate(); err != nil {
		return nil, err
	}
	return n, nil
}

// Validate enforces notice invariants.
func (n *Notice) Validate() error {
	if n.ID == "" {
		return fmt.Errorf("notice: id is required")
	}
	if n.Title == "" {
		return fmt.Errorf("notice: title is required")
	}
	if n.Content == "" {
		return fmt.Errorf("notice: content is required")
	}
	return nil
}

// IsExpired returns true if the notice has passed its expiry date.
func (n *Notice) IsExpired(now time.Time) bool {
	if n.ExpiresAt == nil {
		return false
	}
	return now.After(*n.ExpiresAt)
}

// Touch updates the UpdatedAt timestamp.
func (n *Notice) Touch(now time.Time) { n.UpdatedAt = now }
