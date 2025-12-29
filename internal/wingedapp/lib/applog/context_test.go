package applog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// ERROR CATEGORY TESTS
// ============================================================================

func TestErrorCategories_AllDefined(t *testing.T) {
	t.Parallel()

	// All error categories should have corresponding ErrorInfo entries
	categories := []string{
		ErrCategoryValidation,
		ErrCategoryAuthorization,
		ErrCategoryStateGuard,
		ErrCategoryNotFound,
		ErrCategoryInternal,
		ErrCategoryExternal,
	}

	for _, cat := range categories {
		info, exists := ErrorCategoryInfo[cat]
		assert.True(t, exists, "ErrorCategoryInfo should have entry for %s", cat)
		assert.Equal(t, cat, info.Category, "Category field should match key for %s", cat)
		assert.NotEmpty(t, info.Title, "Title should not be empty for %s", cat)
		assert.NotEmpty(t, info.Description, "Description should not be empty for %s", cat)
		assert.NotEmpty(t, info.Suggestion, "Suggestion should not be empty for %s", cat)
	}
}

func TestGetErrorInfo_KnownCategory(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		category      string
		expectedTitle string
	}{
		{ErrCategoryValidation, "Validation Error"},
		{ErrCategoryAuthorization, "Authorization Error"},
		{ErrCategoryStateGuard, "Invalid State Transition"},
		{ErrCategoryNotFound, "Resource Not Found"},
		{ErrCategoryInternal, "Internal Server Error"},
		{ErrCategoryExternal, "External Service Error"},
	}

	for _, tc := range testCases {
		t.Run(tc.category, func(t *testing.T) {
			info := GetErrorInfo(tc.category)
			assert.Equal(t, tc.category, info.Category)
			assert.Equal(t, tc.expectedTitle, info.Title)
			assert.NotEmpty(t, info.Description)
			assert.NotEmpty(t, info.Suggestion)
		})
	}
}

func TestGetErrorInfo_UnknownCategory(t *testing.T) {
	t.Parallel()

	info := GetErrorInfo("unknown_category")

	assert.Equal(t, "unknown_category", info.Category)
	assert.Equal(t, "Unknown Error", info.Title)
	assert.Contains(t, info.Description, "unrecognized")
	assert.Contains(t, info.Suggestion, "Report")
}

func TestErrorInfo_ValidationCategory_HasHelpfulSuggestion(t *testing.T) {
	t.Parallel()

	info := GetErrorInfo(ErrCategoryValidation)

	// Validation errors should mention checking API docs
	assert.Contains(t, info.Suggestion, "API docs")
	assert.Contains(t, info.Suggestion, "required")
}

func TestErrorInfo_StateGuardCategory_HasHelpfulSuggestion(t *testing.T) {
	t.Parallel()

	info := GetErrorInfo(ErrCategoryStateGuard)

	// State guard errors should mention checking available_actions
	assert.Contains(t, info.Suggestion, "available_actions")
	assert.Contains(t, info.Suggestion, "ui-state")
}

func TestErrorInfo_AuthorizationCategory_HasHelpfulSuggestion(t *testing.T) {
	t.Parallel()

	info := GetErrorInfo(ErrCategoryAuthorization)

	// Authorization errors should mention JWT/user verification
	assert.Contains(t, info.Suggestion, "JWT")
	assert.Contains(t, info.Description, "participant")
}

func TestErrorInfo_InternalCategory_HasCorrelationIDMention(t *testing.T) {
	t.Parallel()

	info := GetErrorInfo(ErrCategoryInternal)

	// Internal errors should tell dev to report with correlation_id
	assert.Contains(t, info.Suggestion, "correlation_id")
	assert.Contains(t, info.Suggestion, "backend team")
}

// ============================================================================
// CONTEXT FUNCTIONS TESTS
// ============================================================================

func TestWithCorrelationID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	corrID := "abc123"

	ctx = WithCorrelationID(ctx, corrID)

	assert.Equal(t, corrID, CorrelationIDFromContext(ctx))
}

func TestCorrelationIDFromContext_Empty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	assert.Empty(t, CorrelationIDFromContext(ctx))
}

func TestWithRequestID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	reqID := "req-456"

	ctx = WithRequestID(ctx, reqID)

	assert.Equal(t, reqID, RequestIDFromContext(ctx))
}

func TestRequestIDFromContext_Empty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	assert.Empty(t, RequestIDFromContext(ctx))
}

func TestWithUserID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	userID := "user-789"

	ctx = WithUserID(ctx, userID)

	assert.Equal(t, userID, UserIDFromContext(ctx))
}

func TestUserIDFromContext_Empty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	assert.Empty(t, UserIDFromContext(ctx))
}

func TestWithLogger(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := NewTestLogger()

	ctx = WithLogger(ctx, logger)

	retrieved := FromContext(ctx)
	require.NotNil(t, retrieved)

	// Verify it's the same logger by logging and checking entries
	retrieved.Info(ctx, "test message")
	entries := logger.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, "test message", entries[0].Message)
}

func TestFromContext_NilWhenNotSet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	assert.Nil(t, FromContext(ctx))
}

func TestContextChaining_MultipleValues(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-001")
	ctx = WithCorrelationID(ctx, "corr-002")
	ctx = WithUserID(ctx, "user-003")

	// All values should be retrievable
	assert.Equal(t, "req-001", RequestIDFromContext(ctx))
	assert.Equal(t, "corr-002", CorrelationIDFromContext(ctx))
	assert.Equal(t, "user-003", UserIDFromContext(ctx))
}

// ============================================================================
// TEST LOGGER - NEW ASSERTION METHODS
// ============================================================================

func TestTestLogger_WarnEntries(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Info(ctx, "info")
	logger.Warn(ctx, "warn 1")
	logger.Error(ctx, "error", nil)
	logger.Warn(ctx, "warn 2")

	warnEntries := logger.WarnEntries()
	require.Len(t, warnEntries, 2)
	assert.Equal(t, "warn 1", warnEntries[0].Message)
	assert.Equal(t, "warn 2", warnEntries[1].Message)
}

func TestTestLogger_InfoEntries(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Info(ctx, "info 1")
	logger.Warn(ctx, "warn")
	logger.Info(ctx, "info 2")

	infoEntries := logger.InfoEntries()
	require.Len(t, infoEntries, 2)
	assert.Equal(t, "info 1", infoEntries[0].Message)
	assert.Equal(t, "info 2", infoEntries[1].Message)
}

func TestTestLogger_EntriesWithField(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Info(ctx, "with user", UserID("u1"))
	logger.Info(ctx, "without user")
	logger.Info(ctx, "with user again", UserID("u2"))

	entries := logger.EntriesWithField("user_id")
	require.Len(t, entries, 2)
	assert.Equal(t, "with user", entries[0].Message)
	assert.Equal(t, "with user again", entries[1].Message)
}

func TestTestLogger_EntriesWithFieldValue(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Info(ctx, "action A", SchedulingAction("suggest_times"))
	logger.Info(ctx, "action B", SchedulingAction("confirm_time"))
	logger.Info(ctx, "action A again", SchedulingAction("suggest_times"))

	entries := logger.EntriesWithFieldValue("scheduling_action", "suggest_times")
	require.Len(t, entries, 2)
	assert.Equal(t, "action A", entries[0].Message)
	assert.Equal(t, "action A again", entries[1].Message)
}

func TestTestLogger_EntriesWithAction(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Info(ctx, "entry 1", SchedulingAction("confirm_time"))
	logger.Info(ctx, "entry 2", SchedulingAction("select_venue"))
	logger.Info(ctx, "entry 3", SchedulingAction("confirm_time"))

	entries := logger.EntriesWithAction("confirm_time")
	require.Len(t, entries, 2)
}

func TestTestLogger_EntriesWithErrorCategory(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Warn(ctx, "validation fail", ErrorCategory(ErrCategoryValidation))
	logger.Warn(ctx, "state guard fail", ErrorCategory(ErrCategoryStateGuard))
	logger.Warn(ctx, "another validation", ErrorCategory(ErrCategoryValidation))

	entries := logger.EntriesWithErrorCategory(ErrCategoryValidation)
	require.Len(t, entries, 2)
	assert.Equal(t, "validation fail", entries[0].Message)
	assert.Equal(t, "another validation", entries[1].Message)
}

func TestTestLogger_EntriesWithCorrelationID(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx1 := WithCorrelationID(context.Background(), "corr-AAA")
	ctx2 := WithCorrelationID(context.Background(), "corr-BBB")

	logger.Info(ctx1, "request A")
	logger.Info(ctx2, "request B")
	logger.Info(ctx1, "request A continued")

	entries := logger.EntriesWithCorrelationID("corr-AAA")
	require.Len(t, entries, 2)
	assert.Equal(t, "request A", entries[0].Message)
	assert.Equal(t, "request A continued", entries[1].Message)
}

func TestTestLogger_HasWarn(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Warn(ctx, "action validation rejected")
	logger.Info(ctx, "some info")

	assert.True(t, logger.HasWarn("validation"))
	assert.True(t, logger.HasWarn("rejected"))
	assert.False(t, logger.HasWarn("error"))
}

func TestTestLogger_HasInfo(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Info(ctx, "scheduling action received")
	logger.Warn(ctx, "some warning")

	assert.True(t, logger.HasInfo("scheduling"))
	assert.True(t, logger.HasInfo("received"))
	assert.False(t, logger.HasInfo("warning"))
}

func TestTestLogger_HasFieldValue(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Info(ctx, "test", ActionSuccess(true), SchedulingAction("confirm_time"))

	assert.True(t, logger.HasFieldValue("action_success", true))
	assert.True(t, logger.HasFieldValue("scheduling_action", "confirm_time"))
	assert.False(t, logger.HasFieldValue("action_success", false))
	assert.False(t, logger.HasFieldValue("nonexistent", "value"))
}

func TestTestLogger_LastWarn(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	assert.Nil(t, logger.LastWarn())

	logger.Warn(ctx, "first warn")
	logger.Info(ctx, "info")
	logger.Warn(ctx, "second warn")
	logger.Info(ctx, "more info")

	lastWarn := logger.LastWarn()
	require.NotNil(t, lastWarn)
	assert.Equal(t, "second warn", lastWarn.Message)
}

func TestTestLogger_ActionSuccessEntries(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Info(ctx, "success 1", ActionSuccess(true))
	logger.Info(ctx, "failure", ActionSuccess(false))
	logger.Info(ctx, "success 2", ActionSuccess(true))
	logger.Info(ctx, "no action_success field")

	successEntries := logger.ActionSuccessEntries()
	require.Len(t, successEntries, 2)
	assert.Equal(t, "success 1", successEntries[0].Message)
	assert.Equal(t, "success 2", successEntries[1].Message)
}

func TestTestLogger_ActionFailureEntries(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := context.Background()

	logger.Info(ctx, "success", ActionSuccess(true))
	logger.Warn(ctx, "failure 1", ActionSuccess(false))
	logger.Warn(ctx, "failure 2", ActionSuccess(false))

	failureEntries := logger.ActionFailureEntries()
	require.Len(t, failureEntries, 2)
	assert.Equal(t, "failure 1", failureEntries[0].Message)
	assert.Equal(t, "failure 2", failureEntries[1].Message)
}

func TestTestLogger_CapturesCorrelationID(t *testing.T) {
	t.Parallel()

	logger := NewTestLogger()
	ctx := WithCorrelationID(context.Background(), "test-corr-id")

	logger.Info(ctx, "test message")

	entries := logger.Entries()
	require.Len(t, entries, 1)
	assert.Equal(t, "test-corr-id", entries[0].CorrelationID)
}
