// Package clock provides a real-time Clock implementation.
package clock

import "time"

// SystemClock implements outbound.Clock using the system clock.
type SystemClock struct{}

// NewSystemClock creates a new SystemClock.
func NewSystemClock() *SystemClock {
	return &SystemClock{}
}

// Now returns the current UTC time.
func (c *SystemClock) Now() time.Time {
	return time.Now().UTC()
}
