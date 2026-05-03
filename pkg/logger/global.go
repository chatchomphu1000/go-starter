package logger

import "sync"

var (
	global Logger
	mu     sync.RWMutex
)

// L returns the global logger. Panics if the global logger has not been set.
func L() Logger {
	mu.RLock()
	defer mu.RUnlock()
	if global == nil {
		panic("logger: global logger not set — call SetGlobal first")
	}
	return global
}

// SetGlobal sets the global logger instance.
func SetGlobal(l Logger) {
	mu.Lock()
	defer mu.Unlock()
	global = l
}
