package logging

import (
	"log/slog"
	"os"
)

// InitLogger configures the global slog logger to use JSON format.
// This should be called exactly once at the beginning of main().
func InitLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	// Set this JSON logger as the default for the entire application.
	// Any calls to slog.Info(), slog.Error(), etc. will now use this.
	slog.SetDefault(logger)
}
