// Package logging provides *slog.Logger functionality to gxpdf
package logging

import (
	"log/slog"
	"sync/atomic"
)

// logger holds the package-level logger instance for debug output.
// Initialized to a discard logger in init() to avoid nil checks.
var logger atomic.Pointer[slog.Logger]

func init() {
	logger.Store(slog.New(slog.DiscardHandler))
}

// SetLogger configures the package-level logger for debug output.
// Pass nil to disable logging (will use slog.DiscardHandler).
// Pass a configured *slog.Logger to capture debug output.
//
// SetLogger is safe for concurrent use.
//
// Example enabling debug output to stderr:
//
//	logging.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, nil)))
func SetLogger(sl *slog.Logger) {
	if sl == nil {
		logger.Store(slog.New(slog.DiscardHandler))
	} else {
		logger.Store(sl)
	}
}

// Logger returns the package-level logger.
// If no logger has been set via SetLogger, returns a discard logger
// that discards all output.
//
// Logger is safe for concurrent use.
func Logger() *slog.Logger {
	return logger.Load()
}
