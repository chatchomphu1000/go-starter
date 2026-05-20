package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validEmail(t *testing.T) Email {
	t.Helper()
	e, err := NewEmail("alice@example.com")
	require.NoError(t, err)
	return e
}

func TestNewUser(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		username       string
		email          Email
		hashedPassword string
		role           Role
		wantErr        bool
	}{
		{
			name:           "success",
			id:             "user-1",
			username:       "Alice",
			email:          "alice@example.com",
			hashedPassword: "hashed",
			role:           RoleUser,
		},
		{
			name:    "empty_id",
			id:      "",
			username: "Alice",
			email:   "alice@example.com",
			role:    RoleUser,
			wantErr: true,
		},
		{
			name:     "empty_name",
			id:       "user-1",
			username: "",
			email:    "alice@example.com",
			role:     RoleUser,
			wantErr:  true,
		},
		{
			name:     "empty_email",
			id:       "user-1",
			username: "Alice",
			email:    "",
			role:     RoleUser,
			wantErr:  true,
		},
		{
			name:     "invalid_role",
			id:       "user-1",
			username: "Alice",
			email:    "alice@example.com",
			role:     Role("superuser"),
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewUser(tc.id, tc.username, tc.email, tc.hashedPassword, tc.role, fixedNow)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.id, got.ID)
				assert.Equal(t, tc.username, got.Name)
				assert.True(t, got.Active)
				assert.Equal(t, fixedNow, got.CreatedAt)
				assert.Equal(t, fixedNow, got.UpdatedAt)
			}
		})
	}
}

func TestUser_Rename(t *testing.T) {
	u, _ := NewUser("u-1", "Alice", validEmail(t), "h", RoleUser, fixedNow)

	assert.NoError(t, u.Rename("Bob"))
	assert.Equal(t, "Bob", u.Name)

	assert.Error(t, u.Rename(""))
	assert.Equal(t, "Bob", u.Name) // unchanged on error
}

func TestUser_ActivateDeactivate(t *testing.T) {
	u, _ := NewUser("u-1", "Alice", validEmail(t), "h", RoleUser, fixedNow)

	assert.True(t, u.Active)
	u.Deactivate()
	assert.False(t, u.Active)
	u.Activate()
	assert.True(t, u.Active)
}

func TestUser_Touch(t *testing.T) {
	u, _ := NewUser("u-1", "Alice", validEmail(t), "h", RoleUser, fixedNow)
	later := fixedNow.AddDate(0, 0, 1)
	u.Touch(later)
	assert.Equal(t, later, u.UpdatedAt)
	assert.Equal(t, fixedNow, u.CreatedAt) // CreatedAt unchanged
}
