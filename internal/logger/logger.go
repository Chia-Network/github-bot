package logger

import (
	"log/slog"
	"os"
)

// Logger is a custom logger from the stdlib slog package
var Logger *slog.Logger

func init() {
	Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}
