package repo

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// InsertNotification represents a notification to be inserted.
type InsertNotification struct {
	UserRefID        string
	NotificationType null.String
	Title            null.String
	Message          string
	Payload          null.JSON
}

// InsertNotification inserts a new notification.
func (s *Store) InsertNotification(
	ctx context.Context,
	exec boil.ContextExecutor,
	inserter *InsertNotification,
) (*pgmodel.Notification, error) {
	if inserter.UserRefID == "" {
		return nil, fmt.Errorf("user_ref_id is required")
	}
	if inserter.Message == "" {
		return nil, fmt.Errorf("message is required")
	}

	n := &pgmodel.Notification{
		UserRefID:        inserter.UserRefID,
		NotificationType: inserter.NotificationType,
		Title:            inserter.Title,
		Message:          inserter.Message,
		Payload:          inserter.Payload,
	}

	if err := n.Insert(ctx, exec, boil.Infer()); err != nil {
		return nil, fmt.Errorf("insert notification: %w", err)
	}

	return n, nil
}

// UpdateNotification contains optional fields for updating a notification.
type UpdateNotification struct {
	ID     string
	ReadAt null.Time
}

// UpdateNotification updates an existing notification (e.g., mark as read).
func (s *Store) UpdateNotification(
	ctx context.Context,
	exec boil.ContextExecutor,
	updater *UpdateNotification,
) (*pgmodel.Notification, error) {
	if updater.ID == "" {
		return nil, fmt.Errorf("id is required")
	}

	n, err := pgmodel.FindNotification(ctx, exec, updater.ID)
	if err != nil {
		return nil, fmt.Errorf("find notification: %w", err)
	}

	cols := make([]string, 0)

	if updater.ReadAt.Valid {
		n.ReadAt = updater.ReadAt
		cols = append(cols, pgmodel.NotificationColumns.ReadAt)
	}

	if len(cols) == 0 {
		return n, nil
	}

	if _, err := n.Update(ctx, exec, boil.Whitelist(cols...)); err != nil {
		return nil, fmt.Errorf("update notification: %w", err)
	}

	return n, nil
}
