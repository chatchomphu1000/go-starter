package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validPayment(t *testing.T) *Payment {
	t.Helper()
	p, err := NewPayment("pay-1", "booking-1", "inv-1", "tenant-1", "owner-1",
		5000.0, "THB", PaymentMethodBankTransfer, "manual", "rent", fixedNow)
	require.NoError(t, err)
	return p
}

func TestNewPayment(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		amount  float64
		wantErr bool
	}{
		{name: "success", id: "pay-1", amount: 5000},
		{name: "empty_id", id: "", amount: 5000, wantErr: true},
		{name: "zero_amount", id: "pay-1", amount: 0, wantErr: true},
		{name: "negative_amount", id: "pay-1", amount: -100, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewPayment(tc.id, "b-1", "inv-1", "t-1", "o-1",
				tc.amount, "THB", PaymentMethodCash, "manual", "", fixedNow)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, PaymentStatusPending, got.Status)
				assert.Nil(t, got.PaidAt)
			}
		})
	}
}

func TestPayment_Complete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := validPayment(t)
		paidAt := fixedNow.AddDate(0, 0, 1)
		require.NoError(t, p.Complete("tx-123", paidAt))
		assert.Equal(t, PaymentStatusCompleted, p.Status)
		assert.Equal(t, "tx-123", p.GatewayTxID)
		assert.NotNil(t, p.PaidAt)
		assert.Equal(t, paidAt, *p.PaidAt)
	})

	t.Run("already_completed", func(t *testing.T) {
		p := validPayment(t)
		p.Status = PaymentStatusCompleted
		err := p.Complete("tx-456", fixedNow)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrPaymentAlreadyPaid)
	})
}

func TestPayment_Fail(t *testing.T) {
	p := validPayment(t)
	later := fixedNow.AddDate(0, 0, 1)
	p.Fail(later)
	assert.Equal(t, PaymentStatusFailed, p.Status)
	assert.Equal(t, later, p.UpdatedAt)
}

func TestPayment_Refund(t *testing.T) {
	t.Run("success_from_completed", func(t *testing.T) {
		p := validPayment(t)
		p.Status = PaymentStatusCompleted
		require.NoError(t, p.Refund(fixedNow))
		assert.Equal(t, PaymentStatusRefunded, p.Status)
	})

	t.Run("fail_from_pending", func(t *testing.T) {
		p := validPayment(t)
		assert.Error(t, p.Refund(fixedNow))
		assert.Equal(t, PaymentStatusPending, p.Status)
	})

	t.Run("fail_from_failed", func(t *testing.T) {
		p := validPayment(t)
		p.Status = PaymentStatusFailed
		assert.Error(t, p.Refund(fixedNow))
	})
}

func TestPayment_Touch(t *testing.T) {
	p := validPayment(t)
	later := fixedNow.AddDate(0, 0, 1)
	p.Touch(later)
	assert.Equal(t, later, p.UpdatedAt)
	assert.Equal(t, fixedNow, p.CreatedAt)
}
