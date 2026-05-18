package domain

import (
	"fmt"
	"time"
)

// Message is a single chat message between two users.
type Message struct {
	ID         string
	ThreadID   string
	SenderID   string
	ReceiverID string
	Content    string
	Read       bool
	CreatedAt  time.Time
}

// NewMessage creates a validated Message entity.
func NewMessage(id, threadID, senderID, receiverID, content string, now time.Time) (*Message, error) {
	m := &Message{
		ID:         id,
		ThreadID:   threadID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		Read:       false,
		CreatedAt:  now,
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// Validate enforces message invariants.
func (m *Message) Validate() error {
	if m.ID == "" {
		return fmt.Errorf("message: id is required")
	}
	if m.SenderID == "" {
		return fmt.Errorf("message: sender_id is required")
	}
	if m.ReceiverID == "" {
		return fmt.Errorf("message: receiver_id is required")
	}
	if m.Content == "" {
		return fmt.Errorf("message: content is required")
	}
	return nil
}

// MarkRead marks the message as read by the receiver.
func (m *Message) MarkRead() { m.Read = true }

// Thread groups messages between two participants.
type Thread struct {
	ID           string
	Participants []string // [userID1, userID2]
	LastMessage  string
	LastAt       time.Time
	UnreadCount  int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewThread creates a new conversation thread.
func NewThread(id string, participants []string, now time.Time) (*Thread, error) {
	if len(participants) != 2 {
		return nil, fmt.Errorf("thread: exactly 2 participants required")
	}
	if participants[0] == participants[1] {
		return nil, fmt.Errorf("thread: participants must be different users")
	}
	return &Thread{
		ID:           id,
		Participants: participants,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// UpdateLastMessage refreshes the thread preview after a new message.
func (t *Thread) UpdateLastMessage(content string, now time.Time) {
	t.LastMessage = content
	t.LastAt = now
	t.UpdatedAt = now
}
