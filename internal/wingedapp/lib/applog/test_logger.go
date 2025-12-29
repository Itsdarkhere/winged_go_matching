package applog

import (
	"context"
	"sync"
)

// LogLevel represents the severity of a log entry.
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
	LevelFatal LogLevel = "fatal"
)

// LogEntry represents a captured log entry for testing.
type LogEntry struct {
	Level         LogLevel
	Message       string
	Error         error
	Fields        map[string]any
	RequestID     string
	CorrelationID string
}

// TestLogger captures log entries for testing assertions.
type TestLogger struct {
	mu      sync.RWMutex
	entries []LogEntry
	fields  map[string]any
	parent  *TestLogger // parent logger for shared entries
}

// NewTestLogger creates a logger for testing.
func NewTestLogger() *TestLogger {
	return &TestLogger{
		entries: make([]LogEntry, 0),
		fields:  make(map[string]any),
	}
}

func (l *TestLogger) capture(ctx context.Context, level LogLevel, msg string, err error, fields []Field) {
	// If we have a parent, delegate to it so entries are shared
	if l.parent != nil {
		// Merge our preset fields with the provided fields
		allFields := make([]Field, 0, len(l.fields)+len(fields))
		for k, v := range l.fields {
			allFields = append(allFields, Field{Key: k, Value: v})
		}
		allFields = append(allFields, fields...)
		l.parent.capture(ctx, level, msg, err, allFields)
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		Level:         level,
		Message:       msg,
		Error:         err,
		Fields:        make(map[string]any),
		RequestID:     RequestIDFromContext(ctx),
		CorrelationID: CorrelationIDFromContext(ctx),
	}

	// Copy preset fields
	for k, v := range l.fields {
		entry.Fields[k] = v
	}

	// Add provided fields
	for _, f := range fields {
		entry.Fields[f.Key] = f.Value
	}

	l.entries = append(l.entries, entry)
}

func (l *TestLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.capture(ctx, LevelDebug, msg, nil, fields)
}

func (l *TestLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.capture(ctx, LevelInfo, msg, nil, fields)
}

func (l *TestLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.capture(ctx, LevelWarn, msg, nil, fields)
}

func (l *TestLogger) Error(ctx context.Context, msg string, err error, fields ...Field) {
	l.capture(ctx, LevelError, msg, err, fields)
}

func (l *TestLogger) Fatal(ctx context.Context, msg string, err error, fields ...Field) {
	l.capture(ctx, LevelFatal, msg, err, fields)
}

func (l *TestLogger) WithFields(fields ...Field) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newLogger := &TestLogger{
		entries: nil, // Will share parent's entries via pointer reference
		fields:  make(map[string]any),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for _, f := range fields {
		newLogger.fields[f.Key] = f.Value
	}

	// Store reference to parent for shared entries
	newLogger.parent = l

	return newLogger
}

// Entries returns all captured log entries.
func (l *TestLogger) Entries() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]LogEntry, len(l.entries))
	copy(result, l.entries)
	return result
}

// Clear removes all captured entries.
func (l *TestLogger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = make([]LogEntry, 0)
}

// ErrorEntries returns only error-level entries.
func (l *TestLogger) ErrorEntries() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []LogEntry
	for _, e := range l.entries {
		if e.Level == LevelError {
			result = append(result, e)
		}
	}
	return result
}

// HasError checks if any error was logged with the given message substring.
func (l *TestLogger) HasError(msgSubstr string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, e := range l.entries {
		if e.Level == LevelError && contains(e.Message, msgSubstr) {
			return true
		}
	}
	return false
}

// HasErrorCode checks if any error was logged with the given error code.
func (l *TestLogger) HasErrorCode(code string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, e := range l.entries {
		if e.Level == LevelError {
			if c, ok := e.Fields["error_code"].(string); ok && c == code {
				return true
			}
		}
	}
	return false
}

// LastEntry returns the most recent log entry.
func (l *TestLogger) LastEntry() *LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.entries) == 0 {
		return nil
	}
	entry := l.entries[len(l.entries)-1]
	return &entry
}

// LastError returns the most recent error-level entry.
func (l *TestLogger) LastError() *LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for i := len(l.entries) - 1; i >= 0; i-- {
		if l.entries[i].Level == LevelError {
			entry := l.entries[i]
			return &entry
		}
	}
	return nil
}

// WarnEntries returns only warn-level entries.
func (l *TestLogger) WarnEntries() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []LogEntry
	for _, e := range l.entries {
		if e.Level == LevelWarn {
			result = append(result, e)
		}
	}
	return result
}

// InfoEntries returns only info-level entries.
func (l *TestLogger) InfoEntries() []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []LogEntry
	for _, e := range l.entries {
		if e.Level == LevelInfo {
			result = append(result, e)
		}
	}
	return result
}

// EntriesWithField returns entries that have a specific field key.
func (l *TestLogger) EntriesWithField(key string) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []LogEntry
	for _, e := range l.entries {
		if _, ok := e.Fields[key]; ok {
			result = append(result, e)
		}
	}
	return result
}

// EntriesWithFieldValue returns entries that have a specific field key-value pair.
func (l *TestLogger) EntriesWithFieldValue(key string, value any) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []LogEntry
	for _, e := range l.entries {
		if v, ok := e.Fields[key]; ok && v == value {
			result = append(result, e)
		}
	}
	return result
}

// EntriesWithAction returns entries for a specific scheduling action.
func (l *TestLogger) EntriesWithAction(action string) []LogEntry {
	return l.EntriesWithFieldValue("scheduling_action", action)
}

// EntriesWithErrorCategory returns entries with a specific error category.
func (l *TestLogger) EntriesWithErrorCategory(category string) []LogEntry {
	return l.EntriesWithFieldValue("error_category", category)
}

// EntriesWithCorrelationID returns entries with a specific correlation ID.
func (l *TestLogger) EntriesWithCorrelationID(correlationID string) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []LogEntry
	for _, e := range l.entries {
		if e.CorrelationID == correlationID {
			result = append(result, e)
		}
	}
	return result
}

// HasWarn checks if any warning was logged with the given message substring.
func (l *TestLogger) HasWarn(msgSubstr string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, e := range l.entries {
		if e.Level == LevelWarn && contains(e.Message, msgSubstr) {
			return true
		}
	}
	return false
}

// HasInfo checks if any info was logged with the given message substring.
func (l *TestLogger) HasInfo(msgSubstr string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, e := range l.entries {
		if e.Level == LevelInfo && contains(e.Message, msgSubstr) {
			return true
		}
	}
	return false
}

// HasFieldValue checks if any entry has a specific field key-value pair.
func (l *TestLogger) HasFieldValue(key string, value any) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, e := range l.entries {
		if v, ok := e.Fields[key]; ok && v == value {
			return true
		}
	}
	return false
}

// LastWarn returns the most recent warn-level entry.
func (l *TestLogger) LastWarn() *LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for i := len(l.entries) - 1; i >= 0; i-- {
		if l.entries[i].Level == LevelWarn {
			entry := l.entries[i]
			return &entry
		}
	}
	return nil
}

// ActionSuccessEntries returns entries where action_success field is true.
func (l *TestLogger) ActionSuccessEntries() []LogEntry {
	return l.EntriesWithFieldValue("action_success", true)
}

// ActionFailureEntries returns entries where action_success field is false.
func (l *TestLogger) ActionFailureEntries() []LogEntry {
	return l.EntriesWithFieldValue("action_success", false)
}

func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && searchSubstr(s, substr))
}

func searchSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
