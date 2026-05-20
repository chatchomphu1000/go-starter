package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixedNow is shared across all domain test files in this package.
var fixedNow = time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    Email
		wantErr bool
	}{
		{
			name: "valid_email",
			raw:  "alice@example.com",
			want: "alice@example.com",
		},
		{
			name: "trims_whitespace",
			raw:  "  alice@example.com  ",
			want: "alice@example.com",
		},
		{
			name: "lowercases",
			raw:  "ALICE@EXAMPLE.COM",
			want: "alice@example.com",
		},
		{
			name:    "empty",
			raw:     "",
			wantErr: true,
		},
		{
			name:    "whitespace_only",
			raw:     "   ",
			wantErr: true,
		},
		{
			name:    "invalid_format_no_at",
			raw:     "notanemail",
			wantErr: true,
		},
		{
			name:    "invalid_format_no_domain",
			raw:     "user@",
			wantErr: true,
		},
		{
			name:    "exceeds_254_chars",
			raw:     strings.Repeat("a", 249) + "@b.com", // 249+6=255 > 254
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewEmail(tc.raw)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidEmail)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestEmail_String(t *testing.T) {
	e := Email("user@example.com")
	assert.Equal(t, "user@example.com", e.String())
}
