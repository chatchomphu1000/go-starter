package mongodb

import (
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

type messageDoc struct {
	ID         string    `bson:"_id"`
	ThreadID   string    `bson:"thread_id"`
	SenderID   string    `bson:"sender_id"`
	ReceiverID string    `bson:"receiver_id"`
	Content    string    `bson:"content"`
	Read       bool      `bson:"read"`
	CreatedAt  time.Time `bson:"created_at"`
}

type threadDoc struct {
	ID           string    `bson:"_id"`
	Participants []string  `bson:"participants"`
	LastMessage  string    `bson:"last_message"`
	LastAt       time.Time `bson:"last_at"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

func messageFromDomain(m *domain.Message) messageDoc {
	return messageDoc{
		ID:         m.ID,
		ThreadID:   m.ThreadID,
		SenderID:   m.SenderID,
		ReceiverID: m.ReceiverID,
		Content:    m.Content,
		Read:       m.Read,
		CreatedAt:  m.CreatedAt,
	}
}

func messageToDomain(d messageDoc) *domain.Message {
	return &domain.Message{
		ID:         d.ID,
		ThreadID:   d.ThreadID,
		SenderID:   d.SenderID,
		ReceiverID: d.ReceiverID,
		Content:    d.Content,
		Read:       d.Read,
		CreatedAt:  d.CreatedAt,
	}
}

func threadFromDomain(t *domain.Thread) threadDoc {
	return threadDoc{
		ID:           t.ID,
		Participants: t.Participants,
		LastMessage:  t.LastMessage,
		LastAt:       t.LastAt,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}

func threadToDomain(d threadDoc) *domain.Thread {
	return &domain.Thread{
		ID:           d.ID,
		Participants: d.Participants,
		LastMessage:  d.LastMessage,
		LastAt:       d.LastAt,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}
