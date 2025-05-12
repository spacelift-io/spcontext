package spcontext

import (
	"io"
	"log/slog"
)

// NewNopLogger creates a slog logger that doesn't log anywhere.
func NewNopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// NewTextLogger creates a text slog logger that logs to the specified writer.
func NewTextLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, nil))
}

// NewJSONLogger creates a JSON slog logger that logs to the specified writer.
func NewJSONLogger(w io.Writer, opts *slog.HandlerOptions) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, opts))
}
