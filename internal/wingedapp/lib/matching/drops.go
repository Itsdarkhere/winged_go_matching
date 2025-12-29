package matching

import (
	"context"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
)

// DropOneMatchPerUser checks all match results that are "approved", but not "dropped".
// It sets one of them to "dropped", in ascending order of created_at.
// Only drops 1 match per user per invocation.
func (l *Logic) DropOneMatchPerUser(ctx context.Context, exec boil.ContextExecutor) error {
	matchResults, err := l.matchResultsNotDropped(ctx, exec)
	if err != nil {
		return fmt.Errorf("match results not dropped: %w", err)
	}

	droppedUsers := make(map[string]bool)
	for _, mr := range matchResults {
		userA := mr.UserAID.String()
		userB := mr.UserBID.String()

		if droppedUsers[userA] || droppedUsers[userB] {
			continue
		}

		if err = l.setMatchResultDropped(ctx, exec, mr.ID.String()); err != nil {
			return fmt.Errorf("set match result dropped: %w", err)
		}

		// record user-pair have now dropped a match
		droppedUsers[userA] = true
		droppedUsers[userB] = true
	}

	return nil
}

func (l *Logic) matchResultsNotDropped(ctx context.Context,
	exec boil.ContextExecutor,
) ([]MatchResult, error) {
	paginated, err := l.matchResultStorer.MatchResults(ctx, exec, &QueryFilterMatchResult{
		IsApproved: null.BoolFrom(true),
		IsDropped:  null.BoolFrom(false),
		OrderBy:    null.StringFrom("created_at"),
		Sort:       null.StringFrom("+"),
	})
	if err != nil {
		return nil, fmt.Errorf("fetch match results not dropped: %w", err)
	}

	return paginated.Data, nil
}

// setMatchResultDropped sets the given match result as dropped.
func (l *Logic) setMatchResultDropped(ctx context.Context,
	exec boil.ContextExecutor,
	matchResultID string,
) error {
	id, err := uuid.Parse(matchResultID)
	if err != nil {
		return fmt.Errorf("parse match result id: %w", err)
	}

	if _, err = l.matchResultStorer.Update(ctx, exec, &UpdateMatchResult{
		ID:        id,
		IsDropped: null.BoolFrom(true),
		DroppedTS: null.TimeFrom(time.Now()),
	}); err != nil {
		return fmt.Errorf("set match result dropped: %w", err)
	}

	return nil
}
