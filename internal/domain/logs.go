//go:generate mockgen -source=logs.go -destination=mocks/logs.go -package=mocks
package domain

import (
	"log/slog"
	"os"
)

// Logger defines the interface for structured logging operations.
// Implementations should provide thread-safe logging capabilities.
type Logger interface {
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Info(msg string, args ...any)
}

// StdoutLogger is the default logger instance that writes structured logs to stdout.
var StdoutLogger = slog.New(slog.NewTextHandler(os.Stdout, nil))
