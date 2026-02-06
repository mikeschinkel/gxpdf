package logging_test

import (
	"log/slog"
	"testing"

	"github.com/coregx/gxpdf/logging"
)

func TestSetLogger(t *testing.T) {
	oldLogger := logging.Logger()
	defer func() { logging.SetLogger(oldLogger) }()

	handler := logging.NewBufferedLogHandler(nil)
	logging.SetLogger(slog.New(handler))

	log := logging.Logger()
	log.Debug("test message", slog.String("key", "value"))

	if !handler.Contains("test message") {
		t.Error("expected SetLogger to configure the package logger")
	}
}

func TestSetLogger_Nil(t *testing.T) {
	oldLogger := logging.Logger()
	defer func() { logging.SetLogger(oldLogger) }()

	logging.SetLogger(nil)

	log := logging.Logger()
	if log == nil {
		t.Fatal("expected Logger() to return non-nil after SetLogger(nil)")
	}

	if log.Handler() != slog.DiscardHandler {
		t.Error("expected Logger() to use slog.DiscardHandler after SetLogger(nil)")
	}
}

func TestLogger_ReturnsDiscardLoggerByDefault(t *testing.T) {
	oldLogger := logging.Logger()
	logging.SetLogger(nil)
	defer func() { logging.SetLogger(oldLogger) }()

	log := logging.Logger()
	if log == nil {
		t.Fatal("expected non-nil logger")
	}

	if log.Handler() != slog.DiscardHandler {
		t.Error("expected default logger to use slog.DiscardHandler")
	}
}

func TestLogger_ReturnsSameInstance(t *testing.T) {
	oldLogger := logging.Logger()
	defer func() { logging.SetLogger(oldLogger) }()

	handler := logging.NewBufferedLogHandler(nil)
	logging.SetLogger(slog.New(handler))

	log1 := logging.Logger()
	log2 := logging.Logger()

	if log1 != log2 {
		t.Error("expected Logger() to return same instance")
	}
}
