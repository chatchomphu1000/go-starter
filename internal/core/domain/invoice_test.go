package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validInvoiceItems() []InvoiceItem {
	return []InvoiceItem{
		{Description: "Rent", Quantity: 1, UnitPrice: 5000, Total: 5000},
		{Description: "Water", Quantity: 1, UnitPrice: 400, Total: 400},
	}
}

func validInvoice(t *testing.T) *Invoice {
	t.Helper()
	inv, err := NewInvoice("inv-1", "booking-1", "tenant-1", "owner-1",
		validInvoiceItems(), fixedNow.AddDate(0, 1, 0), 1, 2025, "", fixedNow)
	require.NoError(t, err)
	return inv
}

func TestNewInvoice(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		bookingID string
		items     []InvoiceItem
		month     int
		year      int
		wantErr   bool
		wantTotal float64
	}{
		{
			name: "success_sums_items", id: "inv-1", bookingID: "b-1",
			items:     validInvoiceItems(),
			month:     1, year: 2025,
			wantTotal: 5400.0,
		},
		{
			name: "success_empty_items", id: "inv-1", bookingID: "b-1",
			items: []InvoiceItem{}, month: 6, year: 2025, wantTotal: 0,
		},
		{name: "empty_id", id: "", bookingID: "b-1", items: validInvoiceItems(), month: 1, year: 2025, wantErr: true},
		{name: "empty_booking_id", id: "inv-1", bookingID: "", items: validInvoiceItems(), month: 1, year: 2025, wantErr: true},
		{name: "month_zero", id: "inv-1", bookingID: "b-1", items: validInvoiceItems(), month: 0, year: 2025, wantErr: true},
		{name: "month_13", id: "inv-1", bookingID: "b-1", items: validInvoiceItems(), month: 13, year: 2025, wantErr: true},
		{name: "year_invalid", id: "inv-1", bookingID: "b-1", items: validInvoiceItems(), month: 1, year: 2019, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewInvoice(tc.id, tc.bookingID, "t-1", "o-1",
				tc.items, fixedNow.AddDate(0, 1, 0), tc.month, tc.year, "", fixedNow)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, InvoiceStatusDraft, got.Status)
				assert.Equal(t, tc.wantTotal, got.TotalAmount)
			}
		})
	}
}

func TestInvoice_Send(t *testing.T) {
	t.Run("success_from_draft", func(t *testing.T) {
		inv := validInvoice(t)
		require.NoError(t, inv.Send(fixedNow))
		assert.Equal(t, InvoiceStatusSent, inv.Status)
	})

	t.Run("fail_from_sent", func(t *testing.T) {
		inv := validInvoice(t)
		inv.Status = InvoiceStatusSent
		assert.Error(t, inv.Send(fixedNow))
	})

	t.Run("fail_from_paid", func(t *testing.T) {
		inv := validInvoice(t)
		inv.Status = InvoiceStatusPaid
		assert.Error(t, inv.Send(fixedNow))
	})
}

func TestInvoice_MarkPaid(t *testing.T) {
	paidAt := fixedNow.AddDate(0, 0, 5)

	t.Run("success", func(t *testing.T) {
		inv := validInvoice(t)
		require.NoError(t, inv.MarkPaid(paidAt))
		assert.Equal(t, InvoiceStatusPaid, inv.Status)
		assert.NotNil(t, inv.PaidAt)
		assert.Equal(t, paidAt, *inv.PaidAt)
	})

	t.Run("already_paid", func(t *testing.T) {
		inv := validInvoice(t)
		inv.Status = InvoiceStatusPaid
		err := inv.MarkPaid(paidAt)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrPaymentAlreadyPaid)
	})
}

func TestInvoice_MarkOverdue(t *testing.T) {
	t.Run("sent_becomes_overdue", func(t *testing.T) {
		inv := validInvoice(t)
		inv.Status = InvoiceStatusSent
		inv.MarkOverdue(fixedNow)
		assert.Equal(t, InvoiceStatusOverdue, inv.Status)
	})

	t.Run("paid_unchanged", func(t *testing.T) {
		inv := validInvoice(t)
		inv.Status = InvoiceStatusPaid
		inv.MarkOverdue(fixedNow)
		assert.Equal(t, InvoiceStatusPaid, inv.Status)
	})

	t.Run("draft_unchanged", func(t *testing.T) {
		inv := validInvoice(t)
		inv.MarkOverdue(fixedNow)
		assert.Equal(t, InvoiceStatusDraft, inv.Status)
	})
}

func TestInvoice_Cancel(t *testing.T) {
	inv := validInvoice(t)
	later := fixedNow.AddDate(0, 0, 1)
	inv.Cancel(later)
	assert.Equal(t, InvoiceStatusCancelled, inv.Status)
	assert.Equal(t, later, inv.UpdatedAt)
}

func TestInvoice_Touch(t *testing.T) {
	inv := validInvoice(t)
	later := fixedNow.AddDate(0, 0, 1)
	inv.Touch(later)
	assert.Equal(t, later, inv.UpdatedAt)
}

// validInvoice is also used in report_service_test.go (same package) — keep exported helper here.
var _ = time.Now // prevent unused import if tests are removed
