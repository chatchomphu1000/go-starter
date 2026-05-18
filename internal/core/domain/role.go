package domain

import "fmt"

// Role represents a user's role in the system.
type Role string

// Role constants.
const (
	RoleAdmin  Role = "admin"
	RoleOwner  Role = "owner"
	RoleTenant Role = "tenant"
	RoleUser   Role = "user"
)

// validRoles is the set of valid roles.
var validRoles = map[Role]struct{}{
	RoleAdmin:  {},
	RoleOwner:  {},
	RoleTenant: {},
	RoleUser:   {},
}

// ParseRole parses a string into a Role, returning ErrInvalidRole if the value is not recognized.
func ParseRole(s string) (Role, error) {
	r := Role(s)
	if !r.IsValid() {
		return "", fmt.Errorf("%w: %q", ErrInvalidRole, s)
	}
	return r, nil
}

// IsValid returns true if the role is a known valid role.
func (r Role) IsValid() bool {
	_, ok := validRoles[r]
	return ok
}

// String returns the string representation of the role.
func (r Role) String() string {
	return string(r)
}
