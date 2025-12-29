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

// InsertBookingReminder represents params for inserting a booking reminder.
type InsertBookingReminder struct {
	DateInstanceRefID uuid.UUID
	RemindAt          time.Time
}

// InsertBookingReminder inserts a new booking reminder record.
func (s *Store) InsertBookingReminder(
	ctx context.Context,
	exec boil.ContextExecutor,
	inserter *InsertBookingReminder,
) (*pgmodel.BookingReminder, error) {
	if inserter.DateInstanceRefID == uuid.Nil {
		return nil, fmt.Errorf("date_instance_ref_id is required")
	}
	if inserter.RemindAt.IsZero() {
		return nil, fmt.Errorf("remind_at is required")
	}

	br := &pgmodel.BookingReminder{
		DateInstanceRefID: inserter.DateInstanceRefID.String(),
		Status:            "Pending",
		RemindAt:          inserter.RemindAt,
	}

	if err := br.Insert(ctx, exec, boil.Infer()); err != nil {
		return nil, fmt.Errorf("insert booking reminder: %w", err)
	}

	return br, nil
}

// UpdateBookingReminder represents params for updating a booking reminder.
type UpdateBookingReminder struct {
	ID      uuid.UUID
	Status  null.String
	FiredAt null.Time
}

// UpdateBookingReminder updates a booking reminder's status and fired_at timestamp.
func (s *Store) UpdateBookingReminder(
	ctx context.Context,
	exec boil.ContextExecutor,
	updater *UpdateBookingReminder,
) (*pgmodel.BookingReminder, error) {
	if updater.ID == uuid.Nil {
		return nil, fmt.Errorf("id is required")
	}

	br, err := pgmodel.FindBookingReminder(ctx, exec, updater.ID.String())
	if err != nil {
		return nil, fmt.Errorf("find booking reminder: %w", err)
	}

	cols := make([]string, 0)

	if updater.Status.Valid {
		br.Status = updater.Status.String
		cols = append(cols, pgmodel.BookingReminderColumns.Status)
	}

	if updater.FiredAt.Valid {
		br.FiredAt = updater.FiredAt
		cols = append(cols, pgmodel.BookingReminderColumns.FiredAt)
	}

	if len(cols) == 0 {
		return br, nil
	}

	if _, err := br.Update(ctx, exec, boil.Whitelist(cols...)); err != nil {
		return nil, fmt.Errorf("update booking reminder: %w", err)
	}

	return br, nil
}
