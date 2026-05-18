package dto

import "github.com/chatchomphu1000/go-starter/internal/core/domain"

// SendMessageRequest is the request body for sending a message.
type SendMessageRequest struct {
	ReceiverID string `json:"receiver_id" validate:"required"`
	Content    string `json:"content" validate:"required,min=1,max=2000"`
}

// MessageResponse is the API representation of a message.
type MessageResponse struct {
	ID         string `json:"id"`
	ThreadID   string `json:"thread_id"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content"`
	Read       bool   `json:"read"`
	CreatedAt  string `json:"created_at"`
}

// ThreadResponse is the API representation of a conversation thread.
type ThreadResponse struct {
	ID           string   `json:"id"`
	Participants []string `json:"participants"`
	LastMessage  string   `json:"last_message"`
	LastAt       string   `json:"last_at"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// MessageListResponse wraps paginated messages.
type MessageListResponse struct {
	Data  []MessageResponse `json:"data"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
}

// ThreadListResponse wraps paginated threads.
type ThreadListResponse struct {
	Data  []ThreadResponse `json:"data"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

// ToMessageResponse maps a domain Message to MessageResponse.
func ToMessageResponse(m *domain.Message) MessageResponse {
	return MessageResponse{
		ID:         m.ID,
		ThreadID:   m.ThreadID,
		SenderID:   m.SenderID,
		ReceiverID: m.ReceiverID,
		Content:    m.Content,
		Read:       m.Read,
		CreatedAt:  m.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToThreadResponse maps a domain Thread to ThreadResponse.
func ToThreadResponse(t *domain.Thread) ThreadResponse {
	return ThreadResponse{
		ID:           t.ID,
		Participants: t.Participants,
		LastMessage:  t.LastMessage,
		LastAt:       t.LastAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:    t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToMessageListResponse builds a paginated message list response.
func ToMessageListResponse(msgs []*domain.Message, total int64, page, limit int) MessageListResponse {
	data := make([]MessageResponse, 0, len(msgs))
	for _, m := range msgs {
		data = append(data, ToMessageResponse(m))
	}
	return MessageListResponse{Data: data, Total: total, Page: page, Limit: limit}
}

// ToThreadListResponse builds a paginated thread list response.
func ToThreadListResponse(threads []*domain.Thread, total int64, page, limit int) ThreadListResponse {
	data := make([]ThreadResponse, 0, len(threads))
	for _, t := range threads {
		data = append(data, ToThreadResponse(t))
	}
	return ThreadListResponse{Data: data, Total: total, Page: page, Limit: limit}
}
