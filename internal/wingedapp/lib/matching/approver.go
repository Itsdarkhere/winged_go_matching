package matching

import (
	"context"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
)

/*
	This file contains admin control methods for match result flags.
	These are single-flag operations for granular admin control.

	Match visibility requires ALL three conditions:
	- is_approved = true
	- is_dropped = true
	- is_expired = false
*/

// SetApproved sets is_approved=true for a match result.
func (l *Logic) SetApproved(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error {
	if _, err := l.matchResultStorer.Update(ctx, exec, &UpdateMatchResult{
		ID:         uuidFromString(matchResultID),
		IsApproved: null.BoolFrom(true),
	}); err != nil {
		return fmt.Errorf("set approved for match %s: %w", matchResultID, err)
	}
	return nil
}

// SetUnapproved sets is_approved=false for a match result.
func (l *Logic) SetUnapproved(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error {
	if _, err := l.matchResultStorer.Update(ctx, exec, &UpdateMatchResult{
		ID:         uuidFromString(matchResultID),
		IsApproved: null.BoolFrom(false),
	}); err != nil {
		return fmt.Errorf("set unapproved for match %s: %w", matchResultID, err)
	}
	return nil
}

// SetDropped sets is_dropped=true for a match result.
func (l *Logic) SetDropped(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error {
	if _, err := l.matchResultStorer.Update(ctx, exec, &UpdateMatchResult{
		ID:        uuidFromString(matchResultID),
		IsDropped: null.BoolFrom(true),
	}); err != nil {
		return fmt.Errorf("set dropped for match %s: %w", matchResultID, err)
	}
	return nil
}

// SetUndroped sets is_dropped=false for a match result.
func (l *Logic) SetUndropped(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error {
	if _, err := l.matchResultStorer.Update(ctx, exec, &UpdateMatchResult{
		ID:        uuidFromString(matchResultID),
		IsDropped: null.BoolFrom(false),
	}); err != nil {
		return fmt.Errorf("set undropped for match %s: %w", matchResultID, err)
	}
	return nil
}

// SetExpired sets is_expired=true for a match result.
func (l *Logic) SetExpired(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error {
	if _, err := l.matchResultStorer.Update(ctx, exec, &UpdateMatchResult{
		ID:        uuidFromString(matchResultID),
		IsExpired: null.BoolFrom(true),
	}); err != nil {
		return fmt.Errorf("set expired for match %s: %w", matchResultID, err)
	}
	return nil
}

// SetUnexpired sets is_expired=false for a match result.
func (l *Logic) SetUnexpired(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error {
	if _, err := l.matchResultStorer.Update(ctx, exec, &UpdateMatchResult{
		ID:        uuidFromString(matchResultID),
		IsExpired: null.BoolFrom(false),
	}); err != nil {
		return fmt.Errorf("set unexpired for match %s: %w", matchResultID, err)
	}
	return nil
}

// uuidFromString is a helper to parse UUID from string for update operations.
func uuidFromString(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}

// Legacy methods for backwards compatibility

// Approve sets is_approved=true (legacy - use SetApproved)
func (l *Logic) Approve(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error {
	return l.SetApproved(ctx, exec, matchResultID)
}

// Unapprove sets is_approved=false (legacy - use SetUnapproved)
func (l *Logic) Unapprove(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error {
	return l.SetUnapproved(ctx, exec, matchResultID)
}
