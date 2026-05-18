package mongodb

import (
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

type maintenanceDoc struct {
	ID          string     `bson:"_id"`
	RoomID      string     `bson:"room_id"`
	TenantID    string     `bson:"tenant_id"`
	OwnerID     string     `bson:"owner_id"`
	Title       string     `bson:"title"`
	Description string     `bson:"description"`
	Category    string     `bson:"category"`
	Priority    string     `bson:"priority"`
	Status      string     `bson:"status"`
	Photos      []string   `bson:"photos"`
	Notes       string     `bson:"notes"`
	ResolvedAt  *time.Time `bson:"resolved_at"`
	CreatedAt   time.Time  `bson:"created_at"`
	UpdatedAt   time.Time  `bson:"updated_at"`
}

func maintenanceFromDomain(t *domain.MaintenanceTicket) maintenanceDoc {
	return maintenanceDoc{
		ID:          t.ID,
		RoomID:      t.RoomID,
		TenantID:    t.TenantID,
		OwnerID:     t.OwnerID,
		Title:       t.Title,
		Description: t.Description,
		Category:    t.Category,
		Priority:    string(t.Priority),
		Status:      string(t.Status),
		Photos:      t.Photos,
		Notes:       t.Notes,
		ResolvedAt:  t.ResolvedAt,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func maintenanceToDomain(d maintenanceDoc) *domain.MaintenanceTicket {
	return &domain.MaintenanceTicket{
		ID:          d.ID,
		RoomID:      d.RoomID,
		TenantID:    d.TenantID,
		OwnerID:     d.OwnerID,
		Title:       d.Title,
		Description: d.Description,
		Category:    d.Category,
		Priority:    domain.TicketPriority(d.Priority),
		Status:      domain.TicketStatus(d.Status),
		Photos:      d.Photos,
		Notes:       d.Notes,
		ResolvedAt:  d.ResolvedAt,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
