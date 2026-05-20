package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validRoom(t *testing.T) *Room {
	t.Helper()
	r, err := NewRoom("room-1", "owner-1", "101", "Room 101",
		RoomTypeStudio, 1, 25.0, 5000.0, 10000.0,
		[]string{}, []string{}, "nice room", fixedNow)
	require.NoError(t, err)
	return r
}

func TestNewRoom(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		ownerID   string
		number    string
		rentPrice float64
		deposit   float64
		wantErr   bool
	}{
		{
			name: "success", id: "room-1", ownerID: "owner-1",
			number: "101", rentPrice: 5000, deposit: 10000,
		},
		{name: "empty_id", id: "", ownerID: "owner-1", number: "101", wantErr: true},
		{name: "empty_owner", id: "room-1", ownerID: "", number: "101", wantErr: true},
		{name: "empty_number", id: "room-1", ownerID: "owner-1", number: "", wantErr: true},
		{
			name: "negative_rent", id: "room-1", ownerID: "owner-1",
			number: "101", rentPrice: -1, wantErr: true,
		},
		{
			name: "negative_deposit", id: "room-1", ownerID: "owner-1",
			number: "101", deposit: -1, wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewRoom(tc.id, tc.ownerID, tc.number, "Room",
				RoomTypeStudio, 1, 25.0, tc.rentPrice, tc.deposit,
				nil, nil, "", fixedNow)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, RoomStatusAvailable, got.Status)
				assert.Equal(t, fixedNow, got.CreatedAt)
			}
		})
	}
}

func TestRoom_IsAvailable(t *testing.T) {
	r := validRoom(t)
	assert.True(t, r.IsAvailable())

	r.SetStatus(RoomStatusOccupied)
	assert.False(t, r.IsAvailable())

	r.SetStatus(RoomStatusMaintenance)
	assert.False(t, r.IsAvailable())

	r.SetStatus(RoomStatusAvailable)
	assert.True(t, r.IsAvailable())
}

func TestRoom_Touch(t *testing.T) {
	r := validRoom(t)
	later := fixedNow.AddDate(0, 0, 1)
	r.Touch(later)
	assert.Equal(t, later, r.UpdatedAt)
	assert.Equal(t, fixedNow, r.CreatedAt)
}
