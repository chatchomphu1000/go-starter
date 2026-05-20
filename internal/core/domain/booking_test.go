package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validBooking(t *testing.T) *Booking {
	t.Helper()
	b, err := NewBooking(
		"booking-1", "room-1", "tenant-1", "owner-1",
		fixedNow, fixedNow.AddDate(0, 1, 0),
		5000.0, 10000.0, "notes", fixedNow,
	)
	require.NoError(t, err)
	return b
}

func TestNewBooking(t *testing.T) {
	later := fixedNow.AddDate(0, 1, 0)

	tests := []struct {
		name        string
		id          string
		roomID      string
		tenantID    string
		ownerID     string
		start, end  time.Time
		monthlyRent float64
		wantErr     bool
	}{
		{
			name: "success", id: "b-1", roomID: "r-1", tenantID: "t-1", ownerID: "o-1",
			start: fixedNow, end: later, monthlyRent: 5000,
		},
		{name: "empty_id", id: "", roomID: "r-1", tenantID: "t-1", ownerID: "o-1", start: fixedNow, end: later, wantErr: true},
		{name: "empty_room_id", id: "b-1", roomID: "", tenantID: "t-1", ownerID: "o-1", start: fixedNow, end: later, wantErr: true},
		{name: "empty_tenant_id", id: "b-1", roomID: "r-1", tenantID: "", ownerID: "o-1", start: fixedNow, end: later, wantErr: true},
		{name: "empty_owner_id", id: "b-1", roomID: "r-1", tenantID: "t-1", ownerID: "", start: fixedNow, end: later, wantErr: true},
		{
			name: "end_before_start", id: "b-1", roomID: "r-1", tenantID: "t-1", ownerID: "o-1",
			start: later, end: fixedNow, wantErr: true,
		},
		{
			name: "end_equals_start", id: "b-1", roomID: "r-1", tenantID: "t-1", ownerID: "o-1",
			start: fixedNow, end: fixedNow, wantErr: true,
		},
		{
			name: "negative_rent", id: "b-1", roomID: "r-1", tenantID: "t-1", ownerID: "o-1",
			start: fixedNow, end: later, monthlyRent: -1, wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewBooking(tc.id, tc.roomID, tc.tenantID, tc.ownerID,
				tc.start, tc.end, tc.monthlyRent, 0, "", fixedNow)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, BookingStatusPending, got.Status)
				assert.Equal(t, fixedNow, got.CreatedAt)
			}
		})
	}
}

func TestBooking_Approve(t *testing.T) {
	t.Run("success_from_pending", func(t *testing.T) {
		b := validBooking(t)
		require.NoError(t, b.Approve(fixedNow))
		assert.Equal(t, BookingStatusApproved, b.Status)
	})

	for _, status := range []BookingStatus{BookingStatusApproved, BookingStatusRejected, BookingStatusActive, BookingStatusCancelled} {
		status := status
		t.Run("fail_from_"+string(status), func(t *testing.T) {
			b := validBooking(t)
			b.Status = status
			err := b.Approve(fixedNow)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidBookingTransition)
		})
	}
}

func TestBooking_Reject(t *testing.T) {
	t.Run("success_from_pending", func(t *testing.T) {
		b := validBooking(t)
		require.NoError(t, b.Reject("not suitable", fixedNow))
		assert.Equal(t, BookingStatusRejected, b.Status)
		assert.Equal(t, "not suitable", b.RejectionReason)
	})

	t.Run("fail_from_approved", func(t *testing.T) {
		b := validBooking(t)
		b.Status = BookingStatusApproved
		err := b.Reject("reason", fixedNow)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidBookingTransition)
	})
}

func TestBooking_Activate(t *testing.T) {
	t.Run("success_from_approved", func(t *testing.T) {
		b := validBooking(t)
		b.Status = BookingStatusApproved
		require.NoError(t, b.Activate(fixedNow))
		assert.Equal(t, BookingStatusActive, b.Status)
	})

	t.Run("fail_from_pending", func(t *testing.T) {
		b := validBooking(t)
		err := b.Activate(fixedNow)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidBookingTransition)
	})
}

func TestBooking_Complete(t *testing.T) {
	t.Run("success_from_active", func(t *testing.T) {
		b := validBooking(t)
		b.Status = BookingStatusActive
		require.NoError(t, b.Complete(fixedNow))
		assert.Equal(t, BookingStatusCompleted, b.Status)
	})

	t.Run("fail_from_approved", func(t *testing.T) {
		b := validBooking(t)
		b.Status = BookingStatusApproved
		err := b.Complete(fixedNow)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidBookingTransition)
	})
}

func TestBooking_Cancel(t *testing.T) {
	t.Run("success_from_pending", func(t *testing.T) {
		b := validBooking(t)
		require.NoError(t, b.Cancel(fixedNow))
		assert.Equal(t, BookingStatusCancelled, b.Status)
	})

	t.Run("success_from_approved", func(t *testing.T) {
		b := validBooking(t)
		b.Status = BookingStatusApproved
		require.NoError(t, b.Cancel(fixedNow))
		assert.Equal(t, BookingStatusCancelled, b.Status)
	})

	for _, status := range []BookingStatus{BookingStatusActive, BookingStatusCompleted, BookingStatusCancelled} {
		status := status
		t.Run("fail_from_"+string(status), func(t *testing.T) {
			b := validBooking(t)
			b.Status = status
			err := b.Cancel(fixedNow)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidBookingTransition)
		})
	}
}

func TestBooking_Touch(t *testing.T) {
	b := validBooking(t)
	later := fixedNow.AddDate(0, 0, 1)
	b.Touch(later)
	assert.Equal(t, later, b.UpdatedAt)
}
