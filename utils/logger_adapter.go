package utils

import (
	"github.com/Kisanlink/aaa-service/v2/internal/interfaces"
	"go.uber.org/zap"
)

// LoggerAdapter adapts *zap.Logger to interfaces.Logger
type LoggerAdapter struct {
	logger *zap.Logger
}

// NewLoggerAdapter creates a new LoggerAdapter
func NewLoggerAdapter(logger *zap.Logger) interfaces.Logger {
	return &LoggerAdapter{
		logger: logger,
	}
}

// Debug logs a debug message
func (l *LoggerAdapter) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Info logs an info message
func (l *LoggerAdapter) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Warn logs a warning message
func (l *LoggerAdapter) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Error logs an error message
func (l *LoggerAdapter) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal logs a fatal message
func (l *LoggerAdapter) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// With returns a new logger with the given fields
func (l *LoggerAdapter) With(fields ...zap.Field) interfaces.Logger {
	return &LoggerAdapter{
		logger: l.logger.With(fields...),
	}
}

// Named returns a new logger with the given name
func (l *LoggerAdapter) Named(name string) interfaces.Logger {
	return &LoggerAdapter{
		logger: l.logger.Named(name),
	}
}

// Sync flushes any buffered log entries
func (l *LoggerAdapter) Sync() error {
	return l.logger.Sync()
}
