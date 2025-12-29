package repo

import (
	"context"
	"fmt"
	"time"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
)

// InsertDateInstance represents a date instance entry to be inserted.
type InsertDateInstance struct {
	MatchResultRefID  string
	Status            string // Required: Date Instance Status (string enum)
	DecisionWindowEnd time.Time
	DateTypeCore      null.String // Date type core (string enum)
	ScheduledTimeUTC  null.Time
	DurationMinutes   null.Int
}

// InsertDateInstance inserts a new date instance.
func (s *Store) InsertDateInstance(
	ctx context.Context,
	exec boil.ContextExecutor,
	inserter *InsertDateInstance,
) (*pgmodel.DateInstance, error) {
	if inserter.MatchResultRefID == "" {
		return nil, fmt.Errorf("match_result_ref_id is required")
	}
	if inserter.Status == "" {
		return nil, fmt.Errorf("status is required")
	}
	if inserter.DecisionWindowEnd.IsZero() {
		return nil, fmt.Errorf("decision_window_end is required")
	}

	di := &pgmodel.DateInstance{
		MatchResultRefID:  inserter.MatchResultRefID,
		Status:            inserter.Status,
		DecisionWindowEnd: inserter.DecisionWindowEnd,
	}

	if inserter.DateTypeCore.Valid {
		di.DateTypeCore = inserter.DateTypeCore
	}
	if inserter.ScheduledTimeUTC.Valid {
		di.ScheduledTimeUtc = inserter.ScheduledTimeUTC
	}
	if inserter.DurationMinutes.Valid {
		di.DurationMinutes = inserter.DurationMinutes
	}

	if err := di.Insert(ctx, exec, boil.Infer()); err != nil {
		return nil, fmt.Errorf("insert date instance: %w", err)
	}

	return di, nil
}

// UpdateDateInstance contains optional fields for updating a date instance.
type UpdateDateInstance struct {
	ID                   uuid.UUID
	VenueRefID           null.String
	ClearVenueRefID      bool        // Set to true to explicitly clear venue_ref_id to NULL
	DateTypeCore         null.String // String enum
	Status               null.String // String enum
	ScheduledTimeUTC     null.Time
	DurationMinutes      null.Int
	BookingStatus        null.String // String enum
	InitiatorConfirmedAt null.Time   // When initiator confirmed attendance
	ReceiverConfirmedAt  null.Time   // When receiver confirmed attendance
	FeedbackStatusUserA  null.String // String enum
	DecisionUserA        null.String // String enum
	DidMeetUserA         null.String // String enum
	FeedbackTextUserA    null.String
	FeedbackStatusUserB  null.String // String enum
	DecisionUserB        null.String // String enum
	DidMeetUserB         null.String // String enum
	FeedbackTextUserB    null.String
}

// UpdateDateInstance updates an existing date instance.
func (s *Store) UpdateDateInstance(
	ctx context.Context,
	exec boil.ContextExecutor,
	updater *UpdateDateInstance,
) (*pgmodel.DateInstance, error) {
	if updater.ID == uuid.Nil {
		return nil, fmt.Errorf("id is required")
	}

	di, err := pgmodel.FindDateInstance(ctx, exec, updater.ID.String())
	if err != nil {
		return nil, fmt.Errorf("find date instance: %w", err)
	}

	cols := make([]string, 0)

	// Handle VenueRefID: either set to a value or explicitly clear to NULL
	if updater.ClearVenueRefID {
		di.VenueRefID = null.String{Valid: false}
		cols = append(cols, pgmodel.DateInstanceColumns.VenueRefID)
	} else if updater.VenueRefID.Valid {
		di.VenueRefID = updater.VenueRefID
		cols = append(cols, pgmodel.DateInstanceColumns.VenueRefID)
	}
	if updater.DateTypeCore.Valid {
		di.DateTypeCore = updater.DateTypeCore
		cols = append(cols, pgmodel.DateInstanceColumns.DateTypeCore)
	}
	if updater.Status.Valid {
		di.Status = updater.Status.String
		cols = append(cols, pgmodel.DateInstanceColumns.Status)
	}
	if updater.ScheduledTimeUTC.Valid {
		di.ScheduledTimeUtc = updater.ScheduledTimeUTC
		cols = append(cols, pgmodel.DateInstanceColumns.ScheduledTimeUtc)
	}
	if updater.DurationMinutes.Valid {
		di.DurationMinutes = updater.DurationMinutes
		cols = append(cols, pgmodel.DateInstanceColumns.DurationMinutes)
	}
	if updater.BookingStatus.Valid {
		di.BookingStatus = updater.BookingStatus
		cols = append(cols, pgmodel.DateInstanceColumns.BookingStatus)
	}

	// Attendance confirmation timestamps
	if updater.InitiatorConfirmedAt.Valid {
		di.InitiatorConfirmedAt = updater.InitiatorConfirmedAt
		cols = append(cols, pgmodel.DateInstanceColumns.InitiatorConfirmedAt)
	}
	if updater.ReceiverConfirmedAt.Valid {
		di.ReceiverConfirmedAt = updater.ReceiverConfirmedAt
		cols = append(cols, pgmodel.DateInstanceColumns.ReceiverConfirmedAt)
	}

	// Feedback: User A
	if updater.FeedbackStatusUserA.Valid {
		di.FeedbackStatusUserA = updater.FeedbackStatusUserA
		cols = append(cols, pgmodel.DateInstanceColumns.FeedbackStatusUserA)
	}
	if updater.DecisionUserA.Valid {
		di.DecisionUserA = updater.DecisionUserA
		cols = append(cols, pgmodel.DateInstanceColumns.DecisionUserA)
	}
	if updater.DidMeetUserA.Valid {
		di.DidMeetUserA = updater.DidMeetUserA
		cols = append(cols, pgmodel.DateInstanceColumns.DidMeetUserA)
	}
	if updater.FeedbackTextUserA.Valid {
		di.FeedbackTextUserA = updater.FeedbackTextUserA
		cols = append(cols, pgmodel.DateInstanceColumns.FeedbackTextUserA)
	}

	// Feedback: User B
	if updater.FeedbackStatusUserB.Valid {
		di.FeedbackStatusUserB = updater.FeedbackStatusUserB
		cols = append(cols, pgmodel.DateInstanceColumns.FeedbackStatusUserB)
	}
	if updater.DecisionUserB.Valid {
		di.DecisionUserB = updater.DecisionUserB
		cols = append(cols, pgmodel.DateInstanceColumns.DecisionUserB)
	}
	if updater.DidMeetUserB.Valid {
		di.DidMeetUserB = updater.DidMeetUserB
		cols = append(cols, pgmodel.DateInstanceColumns.DidMeetUserB)
	}
	if updater.FeedbackTextUserB.Valid {
		di.FeedbackTextUserB = updater.FeedbackTextUserB
		cols = append(cols, pgmodel.DateInstanceColumns.FeedbackTextUserB)
	}

	if len(cols) == 0 {
		return di, nil
	}

	if _, err := di.Update(ctx, exec, boil.Whitelist(cols...)); err != nil {
		return nil, fmt.Errorf("update date instance: %w", err)
	}

	return di, nil
}

// InsertDateInstanceLog represents a log entry to be inserted.
type InsertDateInstanceLog struct {
	DateInstanceRefID string
	UserRefID         null.String
	EventType         string
	OldValue          null.JSON
	NewValue          null.JSON
	Details           null.String
}

// InsertDateInstanceLog inserts a new date instance log entry.
func (s *Store) InsertDateInstanceLog(
	ctx context.Context,
	exec boil.ContextExecutor,
	inserter *InsertDateInstanceLog,
) (*pgmodel.DateInstanceLog, error) {
	if inserter.DateInstanceRefID == "" {
		return nil, fmt.Errorf("date_instance_ref_id is required")
	}
	if inserter.EventType == "" {
		return nil, fmt.Errorf("event_type is required")
	}

	log := &pgmodel.DateInstanceLog{
		DateInstanceRefID: inserter.DateInstanceRefID,
		UserRefID:         inserter.UserRefID,
		EventType:         inserter.EventType,
		OldValue:          inserter.OldValue,
		NewValue:          inserter.NewValue,
		Details:           inserter.Details,
	}

	if err := log.Insert(ctx, exec, boil.Infer()); err != nil {
		return nil, fmt.Errorf("insert date instance log: %w", err)
	}

	return log, nil
}
