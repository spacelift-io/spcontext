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

// NewJSONLoggerWithAtts creates a JSON slog logger that logs to the specified writer with the given attributes.
func NewJSONLoggerWithAtts(w io.Writer, opts *slog.HandlerOptions, atts ...slog.Attr) *slog.Logger {
	handler := slog.NewJSONHandler(w, opts)
	return slog.New(handler.WithAttrs(atts))
}

// NewJSONLogger creates a JSON slog logger that logs to the specified writer.
func NewJSONLogger(w io.Writer, opts *slog.HandlerOptions) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, opts))
}
