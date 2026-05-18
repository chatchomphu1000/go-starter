package domain

import (
	"fmt"
	"time"
)

// TicketPriority reflects urgency of a maintenance request.
type TicketPriority string

// Ticket priority constants.
const (
	PriorityLow    TicketPriority = "low"
	PriorityMedium TicketPriority = "medium"
	PriorityHigh   TicketPriority = "high"
	PriorityUrgent TicketPriority = "urgent"
)

// TicketStatus represents the lifecycle state of a maintenance ticket.
type TicketStatus string

// Ticket status constants.
const (
	TicketStatusOpen       TicketStatus = "open"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusResolved   TicketStatus = "resolved"
	TicketStatusClosed     TicketStatus = "closed"
)

// MaintenanceTicket tracks a repair or service request from a tenant.
type MaintenanceTicket struct {
	ID          string
	RoomID      string
	TenantID    string
	OwnerID     string
	Title       string
	Description string
	Category    string // plumbing | electrical | ac | furniture | other
	Priority    TicketPriority
	Status      TicketStatus
	Photos      []string
	Notes       string // owner notes / resolution description
	ResolvedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewMaintenanceTicket creates a validated MaintenanceTicket entity.
func NewMaintenanceTicket(
	id, roomID, tenantID, ownerID string,
	title, description, category string,
	priority TicketPriority,
	photos []string,
	now time.Time,
) (*MaintenanceTicket, error) {
	t := &MaintenanceTicket{
		ID:          id,
		RoomID:      roomID,
		TenantID:    tenantID,
		OwnerID:     ownerID,
		Title:       title,
		Description: description,
		Category:    category,
		Priority:    priority,
		Status:      TicketStatusOpen,
		Photos:      photos,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return t, nil
}

// Validate enforces ticket invariants.
func (t *MaintenanceTicket) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("ticket: id is required")
	}
	if t.Title == "" {
		return fmt.Errorf("ticket: title is required")
	}
	if t.RoomID == "" {
		return fmt.Errorf("ticket: room_id is required")
	}
	return nil
}

// StartWork transitions the ticket to in-progress.
func (t *MaintenanceTicket) StartWork(now time.Time) error {
	if t.Status != TicketStatusOpen {
		return fmt.Errorf("ticket: only open tickets can be started")
	}
	t.Status = TicketStatusInProgress
	t.UpdatedAt = now
	return nil
}

// Resolve marks the ticket as resolved with optional notes.
func (t *MaintenanceTicket) Resolve(notes string, now time.Time) error {
	if t.Status != TicketStatusInProgress && t.Status != TicketStatusOpen {
		return fmt.Errorf("ticket: cannot resolve a %s ticket", t.Status)
	}
	t.Status = TicketStatusResolved
	t.Notes = notes
	t.ResolvedAt = &now
	t.UpdatedAt = now
	return nil
}

// Close finalises the ticket (tenant confirmation).
func (t *MaintenanceTicket) Close(now time.Time) error {
	if t.Status != TicketStatusResolved {
		return fmt.Errorf("ticket: only resolved tickets can be closed")
	}
	t.Status = TicketStatusClosed
	t.UpdatedAt = now
	return nil
}

// Touch updates the UpdatedAt timestamp.
func (t *MaintenanceTicket) Touch(now time.Time) { t.UpdatedAt = now }
