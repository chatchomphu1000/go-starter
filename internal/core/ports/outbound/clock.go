package outbound

import "time"

//go:generate go run github.com/vektra/mockery/v2 --name=Clock

// Clock is a time abstraction for testability.
type Clock interface {
	Now() time.Time
}
