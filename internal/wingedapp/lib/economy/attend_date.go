package economy

import (
	"context"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// processAttendDate handles crediting wings to a user when they confirm
// they attended a scheduled date (did_meet: true).
// actionInserter.RefID should be the date_instance_id.
func (a *ActionLogger) processAttendDate(ctx context.Context,
	exec boil.ContextExecutor,
	userTotals *UserTotals,
	actionInserter *InsertActionLog,
) error {
	// 1. Check idempotency - skip if already processed for this user + date instance
	existingLogs, err := a.actionLogStorer.ActionLogs(ctx, exec, &QueryFilterActionLog{
		UserID:   null.StringFrom(actionInserter.UserID),
		Category: null.StringFrom(string(ActionAttendDate)),
		RefID:    null.StringFrom(actionInserter.RefID),
		IsActive: null.IntFrom(1),
	})
	if err != nil {
		return fmt.Errorf("check idempotency: %w", err)
	}
	if len(existingLogs) > 0 {
		return nil // Already processed
	}

	// 2. Get or create user totals
	if userTotals == nil {
		userTotals, err = a.userTotalsStorer.Create(ctx, exec, actionInserter.UserID)
		if err != nil {
			return fmt.Errorf("create user totals: %w", err)
		}
	}

	// 3. Insert action log
	actionLog, err := a.actionLogStorer.Insert(ctx, exec, string(ActionAttendDate), &InsertActionLog{
		UserID: actionInserter.UserID,
		RefID:  actionInserter.RefID,
		Type:   ActionAttendDate,
	})
	if err != nil {
		return fmt.Errorf("insert action log: %w", err)
	}

	// 4. Insert transaction (earned wings expire in 30 days)
	if err := a.transactionStorer.Insert(ctx, exec, &InsertTransaction{
		UserID:       actionInserter.UserID,
		ActionTypeID: string(ActionAttendDate),
		ActionRefID:  actionLog.ID,
		WingsAmount:  AttendDateWings,
		Claimed:      true,
		IsCredit:     true,
		ExpiresAt:    null.TimeFrom(time.Now().AddDate(0, 0, EarnedWingsExpiryDays)),
	}); err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}

	// 5. Update user totals
	newWings := userTotals.Wings + AttendDateWings
	if err := a.userTotalsStorer.Update(ctx, exec, &UpdateUserTotals{
		ID:    userTotals.ID,
		Wings: null.IntFrom(newWings),
	}); err != nil {
		return fmt.Errorf("update user wings: %w", err)
	}

	return nil
}
