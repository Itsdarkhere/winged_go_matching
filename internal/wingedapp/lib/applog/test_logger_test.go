package applog

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestLogger_CapturesEntries(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Debug(ctx, "debug message", F("key", "value"))
	logger.Info(ctx, "info message")
	logger.Warn(ctx, "warn message")
	logger.Error(ctx, "error message", errors.New("test error"), ErrorCode("ERR_001"))

	entries := logger.Entries()
	require.Len(t, entries, 4)

	assert.Equal(t, LevelDebug, entries[0].Level)
	assert.Equal(t, "debug message", entries[0].Message)
	assert.Equal(t, "value", entries[0].Fields["key"])

	assert.Equal(t, LevelInfo, entries[1].Level)
	assert.Equal(t, LevelWarn, entries[2].Level)

	assert.Equal(t, LevelError, entries[3].Level)
	assert.Equal(t, "error message", entries[3].Message)
	assert.Error(t, entries[3].Error)
	assert.Equal(t, "ERR_001", entries[3].Fields["error_code"])
}

func TestTestLogger_CapturesRequestID(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := WithRequestID(context.Background(), "req-123")

	logger.Info(ctx, "test message")

	entries := logger.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, "req-123", entries[0].RequestID)
}

func TestTestLogger_WithFields(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	childLogger := logger.WithFields(Component("api"), Operation("auth"))
	ctx := context.Background()

	childLogger.Info(ctx, "test")

	entries := logger.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, "api", entries[0].Fields["component"])
	assert.Equal(t, "auth", entries[0].Fields["operation"])
}

func TestTestLogger_ErrorEntries(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Info(ctx, "info")
	logger.Error(ctx, "error 1", errors.New("e1"))
	logger.Warn(ctx, "warn")
	logger.Error(ctx, "error 2", errors.New("e2"))

	errorEntries := logger.ErrorEntries()
	require.Len(t, errorEntries, 2)
	assert.Equal(t, "error 1", errorEntries[0].Message)
	assert.Equal(t, "error 2", errorEntries[1].Message)
}

func TestTestLogger_HasError(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Error(ctx, "database connection failed", errors.New("conn refused"))

	assert.True(t, logger.HasError("database"))
	assert.True(t, logger.HasError("connection"))
	assert.False(t, logger.HasError("authentication"))
}

func TestTestLogger_HasErrorCode(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Error(ctx, "auth failed", errors.New("invalid token"), ErrorCode("AUTH_002"))

	assert.True(t, logger.HasErrorCode("AUTH_002"))
	assert.False(t, logger.HasErrorCode("AUTH_001"))
}

func TestTestLogger_LastEntry(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	assert.Nil(t, logger.LastEntry())

	logger.Info(ctx, "first")
	logger.Info(ctx, "second")
	logger.Info(ctx, "third")

	last := logger.LastEntry()
	require.NotNil(t, last)
	assert.Equal(t, "third", last.Message)
}

func TestTestLogger_LastError(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	assert.Nil(t, logger.LastError())

	logger.Error(ctx, "first error", errors.New("e1"))
	logger.Info(ctx, "info")
	logger.Error(ctx, "second error", errors.New("e2"))
	logger.Info(ctx, "more info")

	lastErr := logger.LastError()
	require.NotNil(t, lastErr)
	assert.Equal(t, "second error", lastErr.Message)
}

func TestTestLogger_Clear(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Info(ctx, "message")
	require.Len(t, logger.Entries(), 1)

	logger.Clear()
	assert.Empty(t, logger.Entries())
}
