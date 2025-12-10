package domain

import (
	"log/slog"
	"os"
)

type Logger interface {
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Info(msg string, args ...any)
}

var StdoutLogger = slog.New(slog.NewTextHandler(os.Stdout, nil))
