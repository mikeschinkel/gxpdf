package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// BufferedLogHandler implements slog.Handler and captures log records in memory.
// This is useful for testing and debugging PDF extraction to inspect what
// debug messages were generated without writing to stderr.
//
// Example usage:
//
//	handler := logging.NewBufferedLogHandler(nil)
//	logging.SetLogger(slog.New(handler))
//
//	// ... perform PDF extraction ...
//
//	// Inspect captured logs
//	fmt.Println(handler.String())
//
//	// Or check for specific content
//	if handler.Contains("parseDifferencesArray") {
//	    fmt.Println("Differences array was parsed")
//	}
//
// To filter by level:
//
//	handler := logging.NewBufferedLogHandler(&slog.HandlerOptions{
//	    Level: slog.LevelDebug,
//	})
type BufferedLogHandler struct {
	level      slog.Leveler
	buffer     *bytes.Buffer
	mu         sync.Mutex
	preAttrs   []slog.Attr
	groupNames []string
}

// NewBufferedLogHandler creates a new BufferedLogHandler with an empty buffer.
// Pass nil for opts to capture all log levels, or provide HandlerOptions
// to filter by level.
func NewBufferedLogHandler(opts *slog.HandlerOptions) *BufferedLogHandler {
	h := &BufferedLogHandler{
		buffer: &bytes.Buffer{},
	}
	if opts != nil && opts.Level != nil {
		h.level = opts.Level
	}
	return h
}

// Enabled implements slog.Handler. Returns true if the given level is at or
// above the configured minimum level. If no level was configured, returns
// true for all levels.
func (h *BufferedLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	if h.level == nil {
		return true
	}
	return level >= h.level.Level()
}

// Handle implements slog.Handler. Writes log records as JSON lines to the buffer.
func (h *BufferedLogHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	entry := logEntry{
		Level:    r.Level.String(),
		Message:  r.Message,
		DateTime: r.Time.Format(time.DateTime),
	}

	// Include pre-set attributes from WithAttrs, applying group prefixes
	for _, attr := range h.preAttrs {
		entry.Attrs = append(entry.Attrs, h.prefixedAttr(attr))
	}

	r.Attrs(func(attr slog.Attr) bool {
		entry.Attrs = append(entry.Attrs, h.prefixedAttr(attr))
		return true
	})

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	h.buffer.Write(data)
	h.buffer.WriteByte('\n')

	return nil
}

// prefixedAttr returns the string representation of an attribute with group
// name prefixes applied.
func (h *BufferedLogHandler) prefixedAttr(attr slog.Attr) string {
	if len(h.groupNames) == 0 {
		return attr.String()
	}
	prefix := strings.Join(h.groupNames, ".")
	return prefix + "." + attr.String()
}

// WithAttrs implements slog.Handler. Returns a new handler that includes the
// given attributes in all subsequent log records.
func (h *BufferedLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()

	newAttrs := make([]slog.Attr, len(h.preAttrs), len(h.preAttrs)+len(attrs))
	copy(newAttrs, h.preAttrs)
	newAttrs = append(newAttrs, attrs...)

	return &BufferedLogHandler{
		level:      h.level,
		buffer:     h.buffer,
		preAttrs:   newAttrs,
		groupNames: h.groupNames,
	}
}

// WithGroup implements slog.Handler. Returns a new handler that prefixes all
// subsequent attributes with the given group name.
func (h *BufferedLogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	newGroups := make([]string, len(h.groupNames), len(h.groupNames)+1)
	copy(newGroups, h.groupNames)
	newGroups = append(newGroups, name)

	return &BufferedLogHandler{
		level:      h.level,
		buffer:     h.buffer,
		preAttrs:   h.preAttrs,
		groupNames: newGroups,
	}
}

// String returns all captured log output as a string.
func (h *BufferedLogHandler) String() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.buffer.String()
}

// Reset clears all captured log output.
func (h *BufferedLogHandler) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.buffer.Reset()
}

// Contains returns true if the captured output contains the given substring.
func (h *BufferedLogHandler) Contains(s string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return bytes.Contains(h.buffer.Bytes(), []byte(s))
}

// Len returns the number of bytes in the buffer.
func (h *BufferedLogHandler) Len() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.buffer.Len()
}

// logEntry represents a single log record for JSON serialization.
type logEntry struct {
	Level    string   `json:"level"`
	Message  string   `json:"message"`
	DateTime string   `json:"datetime"`
	Attrs    []string `json:"attrs,omitempty"`
}
