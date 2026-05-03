package outbound

//go:generate go run github.com/vektra/mockery/v2 --name=IDGenerator

// IDGenerator abstracts unique ID generation (UUID v7).
type IDGenerator interface {
	New() string
}
