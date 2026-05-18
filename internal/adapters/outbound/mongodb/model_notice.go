package mongodb

import (
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

type noticeDoc struct {
	ID        string     `bson:"_id"`
	OwnerID   string     `bson:"owner_id"`
	Title     string     `bson:"title"`
	Content   string     `bson:"content"`
	Type      string     `bson:"type"`
	Pinned    bool       `bson:"pinned"`
	ExpiresAt *time.Time `bson:"expires_at"`
	CreatedAt time.Time  `bson:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at"`
}

func noticeFromDomain(n *domain.Notice) noticeDoc {
	return noticeDoc{
		ID:        n.ID,
		OwnerID:   n.OwnerID,
		Title:     n.Title,
		Content:   n.Content,
		Type:      string(n.Type),
		Pinned:    n.Pinned,
		ExpiresAt: n.ExpiresAt,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

func noticeToDomain(d noticeDoc) *domain.Notice {
	return &domain.Notice{
		ID:        d.ID,
		OwnerID:   d.OwnerID,
		Title:     d.Title,
		Content:   d.Content,
		Type:      domain.NoticeType(d.Type),
		Pinned:    d.Pinned,
		ExpiresAt: d.ExpiresAt,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}
