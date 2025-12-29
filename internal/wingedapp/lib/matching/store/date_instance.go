package store

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// DateInstanceStore handles date instance data access.
// For Insert/Update, this uses db/repo.Store internally.
type DateInstanceStore struct {
	l    applog.Logger
	repo *repo.Store
}

// InsertDateInstance creates a new date instance and returns its ID.
func (s *DateInstanceStore) InsertDateInstance(
	ctx context.Context,
	exec boil.ContextExecutor,
	inserter *matching.InsertDateInstance,
) (string, error) {
	di, err := s.repo.InsertDateInstance(ctx, exec, &repo.InsertDateInstance{
		MatchResultRefID:  inserter.MatchResultRefID,
		Status:            inserter.Status, // String enum
		DecisionWindowEnd: inserter.DecisionWindowEnd,
		DateTypeCore:      inserter.DateTypeCore, // String enum
		ScheduledTimeUTC:  inserter.ScheduledTimeUTC,
		DurationMinutes:   inserter.DurationMinutes,
	})
	if err != nil {
		return "", fmt.Errorf("insert date instance: %w", err)
	}
	return di.ID, nil
}

// InsertDateInstanceLog creates a new date instance log entry.
func (s *DateInstanceStore) InsertDateInstanceLog(
	ctx context.Context,
	exec boil.ContextExecutor,
	inserter *matching.InsertDateInstanceLog,
) error {
	_, err := s.repo.InsertDateInstanceLog(ctx, exec, &repo.InsertDateInstanceLog{
		DateInstanceRefID: inserter.DateInstanceRefID,
		UserRefID:         inserter.UserRefID,
		EventType:         inserter.EventType,
		OldValue:          inserter.OldValue,
		NewValue:          inserter.NewValue,
		Details:           inserter.Details,
	})
	return err
}
