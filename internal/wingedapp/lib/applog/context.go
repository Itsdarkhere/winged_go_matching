package applog

import "context"

type ctxKey string

const (
	CtxKeyRequestID     ctxKey = "request_id"
	CtxKeyCorrelationID ctxKey = "correlation_id"
	CtxKeyUserID        ctxKey = "user_id"
	CtxKeyLogger        ctxKey = "logger"
)

// Error categories for frontend observability.
const (
	ErrCategoryValidation    = "validation"
	ErrCategoryAuthorization = "authorization"
	ErrCategoryStateGuard    = "state_guard"
	ErrCategoryNotFound      = "not_found"
	ErrCategoryInternal      = "internal"
	ErrCategoryExternal      = "external"
)

// ErrorInfo contains structured error information for frontend debugging.
type ErrorInfo struct {
	Category    string // Machine-readable category (validation, state_guard, etc.)
	Title       string // Short human-readable title
	Description string // Detailed description for debugging
	Suggestion  string // What the frontend dev should do to fix it
}

// ErrorCategoryInfo maps error categories to frontend-friendly descriptions.
var ErrorCategoryInfo = map[string]ErrorInfo{
	ErrCategoryValidation: {
		Category:    ErrCategoryValidation,
		Title:       "Validation Error",
		Description: "The request payload is missing required fields or has invalid values.",
		Suggestion:  "Check the API docs for required fields. Ensure all required payload keys are present and have correct types.",
	},
	ErrCategoryAuthorization: {
		Category:    ErrCategoryAuthorization,
		Title:       "Authorization Error",
		Description: "The user doesn't have permission to perform this action, or is not a participant in this match/date.",
		Suggestion:  "Verify the user is part of this match. Check that the JWT token belongs to the correct user.",
	},
	ErrCategoryStateGuard: {
		Category:    ErrCategoryStateGuard,
		Title:       "Invalid State Transition",
		Description: "This action is not allowed in the current UI state or for the user's role (initiator/receiver).",
		Suggestion:  "Call GET /ui-state first to check available_actions. Only dispatch actions listed there.",
	},
	ErrCategoryNotFound: {
		Category:    ErrCategoryNotFound,
		Title:       "Resource Not Found",
		Description: "The requested resource (date instance, match, venue, etc.) does not exist.",
		Suggestion:  "Verify the ID is correct. The resource may have been deleted or never existed.",
	},
	ErrCategoryInternal: {
		Category:    ErrCategoryInternal,
		Title:       "Internal Server Error",
		Description: "An unexpected error occurred on the server. This is likely a bug.",
		Suggestion:  "Report this to backend team with the correlation_id from the response/logs.",
	},
	ErrCategoryExternal: {
		Category:    ErrCategoryExternal,
		Title:       "External Service Error",
		Description: "A third-party service (venue search, booking API, etc.) failed or is unavailable.",
		Suggestion:  "Retry the request. If it persists, the external service may be down.",
	},
}

// GetErrorInfo returns the ErrorInfo for a category, with a fallback for unknown categories.
func GetErrorInfo(category string) ErrorInfo {
	if info, ok := ErrorCategoryInfo[category]; ok {
		return info
	}
	return ErrorInfo{
		Category:    category,
		Title:       "Unknown Error",
		Description: "An unrecognized error category was returned.",
		Suggestion:  "Report this to backend team with the full error details.",
	}
}

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CtxKeyRequestID, id)
}

// RequestIDFromContext extracts the request ID from context.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(CtxKeyRequestID).(string); ok {
		return id
	}
	return ""
}

// WithCorrelationID adds a correlation ID to the context.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CtxKeyCorrelationID, id)
}

// CorrelationIDFromContext extracts the correlation ID from context.
func CorrelationIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(CtxKeyCorrelationID).(string); ok {
		return id
	}
	return ""
}

// WithUserID adds a user ID to the context.
func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CtxKeyUserID, id)
}

// UserIDFromContext extracts the user ID from context.
func UserIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(CtxKeyUserID).(string); ok {
		return id
	}
	return ""
}

// WithLogger adds a logger to the context.
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, CtxKeyLogger, logger)
}

// FromContext extracts the logger from context, returns nil if not present.
func FromContext(ctx context.Context) Logger {
	if l, ok := ctx.Value(CtxKeyLogger).(Logger); ok {
		return l
	}
	return nil
}
