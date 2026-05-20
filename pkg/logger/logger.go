// Package logger provides structured logging built on top of Uber Zap.
package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// contextKey is the type for context value keys in this package.
type contextKey string

// RequestIDKey is the context key for the request ID.
const RequestIDKey contextKey = "request_id"

// LoggerConfig holds configuration for the logger.
//
//nolint:revive
type LoggerConfig struct {
	Level       string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
	Format      string `mapstructure:"format" validate:"required,oneof=json console"`
	Development bool   `mapstructure:"development"`
}

// Logger is the application logging interface.
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	WithContext(ctx context.Context) Logger
	Sync() error
}

// zapLogger wraps zap.Logger to implement Logger.
type zapLogger struct {
	zap *zap.Logger
}

// NewLogger creates a new Logger from the given configuration.
func NewLogger(cfg LoggerConfig) (Logger, error) {
	var zapCfg zap.Config

	if cfg.Development {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapCfg = zap.NewProductionConfig()
		zapCfg.EncoderConfig.TimeKey = "timestamp"
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	if cfg.Format == "console" {
		zapCfg.Encoding = "console"
	} else {
		zapCfg.Encoding = "json"
	}

	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	zapCfg.Level = zap.NewAtomicLevelAt(level)

	z, err := zapCfg.Build(zap.AddCallerSkip(0))
	if err != nil {
		return nil, err
	}

	return &zapLogger{zap: z}, nil
}

// Debug logs a message at debug level.
func (l *zapLogger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

// Info logs a message at info level.
func (l *zapLogger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Warn logs a message at warn level.
func (l *zapLogger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Error logs a message at error level.
func (l *zapLogger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// With returns a new Logger with the given fields attached.
func (l *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{zap: l.zap.With(fields...)}
}

// WithContext returns a new Logger with the request_id extracted from the context.
func (l *zapLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok && reqID != "" {
		return l.With(zap.String("request_id", reqID))
	}
	return l
}

// Sync flushes any buffered log entries.
func (l *zapLogger) Sync() error {
	return l.zap.Sync()
}
