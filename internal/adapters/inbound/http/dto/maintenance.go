package dto

import "github.com/chatchomphu1000/go-starter/internal/core/domain"

// CreateTicketRequest is the request body for opening a maintenance ticket.
type CreateTicketRequest struct {
	RoomID      string   `json:"room_id" validate:"required"`
	OwnerID     string   `json:"owner_id" validate:"required"`
	Title       string   `json:"title" validate:"required,min=3,max=200"`
	Description string   `json:"description" validate:"required,min=5,max=2000"`
	Category    string   `json:"category" validate:"required,oneof=plumbing electrical ac furniture other"`
	Priority    string   `json:"priority" validate:"required,oneof=low medium high urgent"`
	Photos      []string `json:"photos"`
}

// ResolveTicketRequest carries resolution notes from the owner.
type ResolveTicketRequest struct {
	Notes string `json:"notes" validate:"max=2000"`
}

// TicketResponse is the API representation of a maintenance ticket.
type TicketResponse struct {
	ID          string   `json:"id"`
	RoomID      string   `json:"room_id"`
	TenantID    string   `json:"tenant_id"`
	OwnerID     string   `json:"owner_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Priority    string   `json:"priority"`
	Status      string   `json:"status"`
	Photos      []string `json:"photos"`
	Notes       string   `json:"notes,omitempty"`
	ResolvedAt  *string  `json:"resolved_at,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// TicketListResponse wraps a paginated list of tickets.
type TicketListResponse struct {
	Data  []TicketResponse `json:"data"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

// ToTicketResponse maps a domain MaintenanceTicket to TicketResponse.
func ToTicketResponse(t *domain.MaintenanceTicket) TicketResponse {
	photos := t.Photos
	if photos == nil {
		photos = []string{}
	}
	r := TicketResponse{
		ID:          t.ID,
		RoomID:      t.RoomID,
		TenantID:    t.TenantID,
		OwnerID:     t.OwnerID,
		Title:       t.Title,
		Description: t.Description,
		Category:    t.Category,
		Priority:    string(t.Priority),
		Status:      string(t.Status),
		Photos:      photos,
		Notes:       t.Notes,
		CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if t.ResolvedAt != nil {
		s := t.ResolvedAt.Format("2006-01-02T15:04:05Z")
		r.ResolvedAt = &s
	}
	return r
}

// ToTicketListResponse builds a paginated ticket list response.
func ToTicketListResponse(tickets []*domain.MaintenanceTicket, total int64, page, limit int) TicketListResponse {
	data := make([]TicketResponse, 0, len(tickets))
	for _, t := range tickets {
		data = append(data, ToTicketResponse(t))
	}
	return TicketListResponse{Data: data, Total: total, Page: page, Limit: limit}
}
