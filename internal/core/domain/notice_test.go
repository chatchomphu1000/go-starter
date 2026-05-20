package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validNotice(t *testing.T) *Notice {
	t.Helper()
	n, err := NewNotice("notice-1", "owner-1", "Test Title", "Test content",
		NoticeTypeGeneral, false, nil, fixedNow)
	require.NoError(t, err)
	return n
}

func TestNewNotice(t *testing.T) {
	exp := fixedNow.AddDate(0, 1, 0)

	tests := []struct {
		name      string
		id        string
		title     string
		content   string
		expiresAt *time.Time
		wantErr   bool
	}{
		{name: "success_no_expiry", id: "n-1", title: "Title", content: "Content"},
		{name: "success_with_expiry", id: "n-1", title: "Title", content: "Content", expiresAt: &exp},
		{name: "empty_id", id: "", title: "Title", content: "Content", wantErr: true},
		{name: "empty_title", id: "n-1", title: "", content: "Content", wantErr: true},
		{name: "empty_content", id: "n-1", title: "Title", content: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewNotice(tc.id, "owner-1", tc.title, tc.content,
				NoticeTypeGeneral, false, tc.expiresAt, fixedNow)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expiresAt, got.ExpiresAt)
			}
		})
	}
}

func TestNotice_IsExpired(t *testing.T) {
	now := fixedNow

	t.Run("nil_expiry_never_expires", func(t *testing.T) {
		n := validNotice(t)
		assert.False(t, n.IsExpired(now))
	})

	t.Run("future_expiry_not_expired", func(t *testing.T) {
		exp := now.AddDate(0, 1, 0)
		n := validNotice(t)
		n.ExpiresAt = &exp
		assert.False(t, n.IsExpired(now))
	})

	t.Run("past_expiry_is_expired", func(t *testing.T) {
		exp := now.AddDate(0, -1, 0)
		n := validNotice(t)
		n.ExpiresAt = &exp
		assert.True(t, n.IsExpired(now))
	})

	t.Run("exact_expiry_time_not_expired", func(t *testing.T) {
		n := validNotice(t)
		n.ExpiresAt = &now
		// now.After(now) == false, so not expired
		assert.False(t, n.IsExpired(now))
	})
}

func TestNotice_Touch(t *testing.T) {
	n := validNotice(t)
	later := fixedNow.AddDate(0, 0, 1)
	n.Touch(later)
	assert.Equal(t, later, n.UpdatedAt)
	assert.Equal(t, fixedNow, n.CreatedAt)
}
