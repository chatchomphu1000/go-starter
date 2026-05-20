package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validTicketDomain(t *testing.T) *MaintenanceTicket {
	t.Helper()
	ticket, err := NewMaintenanceTicket(
		"ticket-1", "room-1", "tenant-1", "owner-1",
		"Broken AC", "AC not cooling", "ac",
		PriorityHigh, []string{}, fixedNow,
	)
	require.NoError(t, err)
	return ticket
}

func TestNewMaintenanceTicket(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		roomID  string
		title   string
		wantErr bool
	}{
		{name: "success", id: "t-1", roomID: "r-1", title: "Broken AC"},
		{name: "empty_id", id: "", roomID: "r-1", title: "Broken AC", wantErr: true},
		{name: "empty_room_id", id: "t-1", roomID: "", title: "Broken AC", wantErr: true},
		{name: "empty_title", id: "t-1", roomID: "r-1", title: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewMaintenanceTicket(tc.id, tc.roomID, "tenant-1", "owner-1",
				tc.title, "", "ac", PriorityMedium, nil, fixedNow)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, TicketStatusOpen, got.Status)
			}
		})
	}
}

func TestMaintenanceTicket_StartWork(t *testing.T) {
	t.Run("success_from_open", func(t *testing.T) {
		ticket := validTicketDomain(t)
		require.NoError(t, ticket.StartWork(fixedNow))
		assert.Equal(t, TicketStatusInProgress, ticket.Status)
	})

	for _, status := range []TicketStatus{TicketStatusInProgress, TicketStatusResolved, TicketStatusClosed} {
		status := status
		t.Run("fail_from_"+string(status), func(t *testing.T) {
			ticket := validTicketDomain(t)
			ticket.Status = status
			assert.Error(t, ticket.StartWork(fixedNow))
			assert.Equal(t, status, ticket.Status) // unchanged
		})
	}
}

func TestMaintenanceTicket_Resolve(t *testing.T) {
	t.Run("success_from_open", func(t *testing.T) {
		ticket := validTicketDomain(t)
		require.NoError(t, ticket.Resolve("Fixed", fixedNow))
		assert.Equal(t, TicketStatusResolved, ticket.Status)
		assert.Equal(t, "Fixed", ticket.Notes)
		assert.NotNil(t, ticket.ResolvedAt)
	})

	t.Run("success_from_in_progress", func(t *testing.T) {
		ticket := validTicketDomain(t)
		ticket.Status = TicketStatusInProgress
		require.NoError(t, ticket.Resolve("Fixed", fixedNow))
		assert.Equal(t, TicketStatusResolved, ticket.Status)
	})

	t.Run("fail_from_closed", func(t *testing.T) {
		ticket := validTicketDomain(t)
		ticket.Status = TicketStatusClosed
		assert.Error(t, ticket.Resolve("notes", fixedNow))
		assert.Equal(t, TicketStatusClosed, ticket.Status)
	})

	t.Run("fail_from_resolved", func(t *testing.T) {
		ticket := validTicketDomain(t)
		ticket.Status = TicketStatusResolved
		assert.Error(t, ticket.Resolve("notes", fixedNow))
	})
}

func TestMaintenanceTicket_Close(t *testing.T) {
	t.Run("success_from_resolved", func(t *testing.T) {
		ticket := validTicketDomain(t)
		ticket.Status = TicketStatusResolved
		require.NoError(t, ticket.Close(fixedNow))
		assert.Equal(t, TicketStatusClosed, ticket.Status)
	})

	for _, status := range []TicketStatus{TicketStatusOpen, TicketStatusInProgress, TicketStatusClosed} {
		status := status
		t.Run("fail_from_"+string(status), func(t *testing.T) {
			ticket := validTicketDomain(t)
			ticket.Status = status
			assert.Error(t, ticket.Close(fixedNow))
		})
	}
}

func TestMaintenanceTicket_Touch(t *testing.T) {
	ticket := validTicketDomain(t)
	later := fixedNow.AddDate(0, 0, 1)
	ticket.Touch(later)
	assert.Equal(t, later, ticket.UpdatedAt)
}
