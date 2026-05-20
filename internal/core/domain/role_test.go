package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRole(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Role
		wantErr bool
	}{
		{name: "admin", input: "admin", want: RoleAdmin},
		{name: "owner", input: "owner", want: RoleOwner},
		{name: "tenant", input: "tenant", want: RoleTenant},
		{name: "user", input: "user", want: RoleUser},
		{name: "invalid", input: "superuser", wantErr: true},
		{name: "empty", input: "", wantErr: true},
		{name: "case_sensitive", input: "Admin", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseRole(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidRole)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestRole_IsValid(t *testing.T) {
	assert.True(t, RoleAdmin.IsValid())
	assert.True(t, RoleOwner.IsValid())
	assert.True(t, RoleTenant.IsValid())
	assert.True(t, RoleUser.IsValid())
	assert.False(t, Role("superuser").IsValid())
	assert.False(t, Role("").IsValid())
}

func TestRole_String(t *testing.T) {
	assert.Equal(t, "admin", RoleAdmin.String())
	assert.Equal(t, "owner", RoleOwner.String())
}
