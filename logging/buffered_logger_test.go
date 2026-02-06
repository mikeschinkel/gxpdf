package logging_test

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/coregx/gxpdf/logging"
)

func TestBufferedLogHandler_CapturesOutput(t *testing.T) {
	handler := logging.NewBufferedLogHandler(nil)
	logger := slog.New(handler)

	logger.Debug("test debug message", slog.String("key", "value"))
	logger.Info("test info message", slog.Int("count", 42))
	logger.Warn("test warning")

	output := handler.String()
	if output == "" {
		t.Error("expected captured output, got empty string")
	}

	if !handler.Contains("test debug message") {
		t.Error("expected output to contain 'test debug message'")
	}
	if !handler.Contains("test info message") {
		t.Error("expected output to contain 'test info message'")
	}
	if !handler.Contains("key=value") {
		t.Error("expected output to contain 'key=value' attribute")
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 log lines, got %d", len(lines))
	}
}

func TestBufferedLogHandler_Reset(t *testing.T) {
	handler := logging.NewBufferedLogHandler(nil)
	logger := slog.New(handler)

	logger.Info("message before reset")
	if handler.Len() == 0 {
		t.Error("expected non-zero length before reset")
	}

	handler.Reset()
	if handler.Len() != 0 {
		t.Error("expected zero length after reset")
	}
	if handler.String() != "" {
		t.Error("expected empty string after reset")
	}
}

func TestBufferedLogHandler_Enabled_NilLevel(t *testing.T) {
	handler := logging.NewBufferedLogHandler(nil)

	levels := []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}
	for _, level := range levels {
		if !handler.Enabled(nil, level) {
			t.Errorf("expected Enabled(%v) to return true with nil level", level)
		}
	}
}

func TestBufferedLogHandler_Enabled_WithLevel(t *testing.T) {
	handler := logging.NewBufferedLogHandler(&slog.HandlerOptions{
		Level: slog.LevelWarn,
	})

	if handler.Enabled(nil, slog.LevelDebug) {
		t.Error("expected DEBUG to be filtered when level is WARN")
	}
	if handler.Enabled(nil, slog.LevelInfo) {
		t.Error("expected INFO to be filtered when level is WARN")
	}
	if !handler.Enabled(nil, slog.LevelWarn) {
		t.Error("expected WARN to be enabled when level is WARN")
	}
	if !handler.Enabled(nil, slog.LevelError) {
		t.Error("expected ERROR to be enabled when level is WARN")
	}
}

func TestBufferedLogHandler_WithAttrs_PreservesAttrs(t *testing.T) {
	handler := logging.NewBufferedLogHandler(nil)
	derived := handler.WithAttrs([]slog.Attr{slog.String("func", "parseDifferencesArray")})

	// WithAttrs should return a new handler, not the same one
	if derived == handler {
		t.Error("expected WithAttrs to return a new handler")
	}

	// Log through the derived handler and verify pre-set attrs appear
	logger := slog.New(derived.(slog.Handler))
	logger.Info("test message")

	if !handler.Contains("func=parseDifferencesArray") {
		t.Errorf("expected output to contain pre-set attr 'func=parseDifferencesArray', got: %s", handler.String())
	}
}

func TestBufferedLogHandler_WithAttrs_SharesBuffer(t *testing.T) {
	handler := logging.NewBufferedLogHandler(nil)
	derived := handler.WithAttrs([]slog.Attr{slog.String("key", "value")})

	// Log through the derived handler
	logger := slog.New(derived.(slog.Handler))
	logger.Info("derived message")

	// Both handlers should see the output (shared buffer)
	if !handler.Contains("derived message") {
		t.Error("expected original handler to see output from derived handler")
	}
}

func TestBufferedLogHandler_WithGroup_PrefixesAttrs(t *testing.T) {
	handler := logging.NewBufferedLogHandler(nil)
	derived := handler.WithGroup("mygroup")

	// WithGroup should return a new handler
	if derived == handler {
		t.Error("expected WithGroup to return a new handler")
	}

	logger := slog.New(derived.(slog.Handler))
	logger.Info("grouped message", slog.String("key", "value"))

	if !handler.Contains("mygroup.key=value") {
		t.Errorf("expected output to contain 'mygroup.key=value', got: %s", handler.String())
	}
}

func TestBufferedLogHandler_WithGroup_EmptyName(t *testing.T) {
	handler := logging.NewBufferedLogHandler(nil)
	derived := handler.WithGroup("")

	if derived != handler {
		t.Error("expected WithGroup('') to return same handler")
	}
}

func TestBufferedLogHandler_Contains(t *testing.T) {
	handler := logging.NewBufferedLogHandler(nil)
	logger := slog.New(handler)

	logger.Info("unique test string xyz123")

	if !handler.Contains("xyz123") {
		t.Error("expected Contains to find 'xyz123'")
	}
	if handler.Contains("not present") {
		t.Error("expected Contains to return false for missing string")
	}
}

func TestBufferedLogHandler_Len(t *testing.T) {
	handler := logging.NewBufferedLogHandler(nil)

	if handler.Len() != 0 {
		t.Error("expected Len() to be 0 for new handler")
	}

	logger := slog.New(handler)
	logger.Info("test")

	if handler.Len() == 0 {
		t.Error("expected Len() to be non-zero after logging")
	}
}

func TestBufferedLogHandler_String(t *testing.T) {
	handler := logging.NewBufferedLogHandler(nil)

	if handler.String() != "" {
		t.Error("expected String() to be empty for new handler")
	}

	logger := slog.New(handler)
	logger.Info("test message")

	output := handler.String()
	if output == "" {
		t.Error("expected String() to be non-empty after logging")
	}
	if !strings.Contains(output, "test message") {
		t.Error("expected String() to contain logged message")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("expected String() to contain log level")
	}
}
