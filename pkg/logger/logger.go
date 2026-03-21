package logger

import (
	"log/slog"
	"os"
)

// L is the package-level structured logger.
var L *slog.Logger

func init() {
	env := os.Getenv("APP_ENV")
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	if env == "production" || env == "prod" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	L = slog.New(handler)
}

// New returns a logger with the given component name attached.
func New(component string) *slog.Logger {
	return L.With("component", component)
}
