// Package logging provides *slog.Logger functionality to gxpdf
package logging

import (
	"log/slog"
	"sync/atomic"
)

// logger holds the package-level logger instance for debug output.
// Defaults to nil, which causes Logger() to return a discard logger.
var logger atomic.Pointer[slog.Logger]

// newDiscardLogger creates a logger that discards all output.
func newDiscardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
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
//
// Example capturing logs in tests:
//
//	handler := logging.NewBufferedLogHandler(nil)
//	logging.SetLogger(slog.New(handler))
//	// ... run extraction ...
//	fmt.Println(handler.String()) // inspect captured logs
func SetLogger(sl *slog.Logger) {
	if sl == nil {
		logger.Store(newDiscardLogger())
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
	l := logger.Load()
	if l == nil {
		l = newDiscardLogger()
		logger.Store(l)
	}
	return l
}
