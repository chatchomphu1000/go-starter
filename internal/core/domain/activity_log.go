package domain

import "time"

// ActivityLog is an immutable audit record of a user action.
type ActivityLog struct {
	ID         string
	UserID     string
	Action     string // e.g. "room.create", "booking.approve"
	Resource   string // e.g. "room", "booking"
	ResourceID string
	IPAddress  string
	UserAgent  string
	CreatedAt  time.Time
}

// NewActivityLog creates an ActivityLog entry.
func NewActivityLog(id, userID, action, resource, resourceID, ip, ua string, now time.Time) *ActivityLog {
	return &ActivityLog{
		ID:         id,
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ip,
		UserAgent:  ua,
		CreatedAt:  now,
	}
}
