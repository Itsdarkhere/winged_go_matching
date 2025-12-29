package economy

import (
	"context"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

type timeAdder func(d *time.Time)

func (a *ActionLogger) addWingedXPayment(ctx context.Context,
	exec boil.ContextExecutor,
	userTotals *UserTotals,
	actionInserter *InsertActionLog,
	categoryPayment string,
	timeAdder timeAdder,
) error {
	// Idempotency: check if this RefID already processed (webhook replay protection)
	if actionInserter.RefID != "" {
		existingLogs, err := a.actionLogStorer.ActionLogs(ctx, exec, &QueryFilterActionLog{
			RefID:    null.StringFrom(actionInserter.RefID),
			IsActive: null.IntFrom(1),
		})
		if err != nil {
			return fmt.Errorf("check idempotency: %w", err)
		}
		if len(existingLogs) > 0 {
			return nil // Already processed, idempotent success
		}
	}

	if _, err := a.actionLogStorer.Insert(ctx, exec, categoryPayment, actionInserter); err != nil {
		return fmt.Errorf("insert action log: %w", err)
	}

	premiumsExpiresIn := userTotals.PremiumExpiresIn
	if premiumsExpiresIn.IsZero() {
		premiumsExpiresIn = null.TimeFrom(time.Now().UTC())
	}

	// add X to time
	timeAdder(&premiumsExpiresIn.Time)
	updater := &UpdateUserTotals{
		ID:               userTotals.ID,
		PremiumExpiresIn: premiumsExpiresIn,
	}
	if err := a.userTotalsStorer.Update(ctx, exec, updater); err != nil {
		return fmt.Errorf("update user totals: %w", err)
	}

	return nil
}

// AddMessage adds a new message to the accounting system.
func (a *ActionLogger) addWingedXMonthlyPayment(ctx context.Context,
	exec boil.ContextExecutor,
	userTotals *UserTotals,
	actionInserter *InsertActionLog,
) error {
	if err := a.addWingedXPayment(ctx,
		exec,
		userTotals,
		actionInserter,
		CategoryWingedxMonthlyPayment,
		func(d *time.Time) {
			*d = d.AddDate(0, CountWingedXPaymentMonth, 0) // add 1 month (normalised via Go time package)
		},
	); err != nil {
		return fmt.Errorf("add wingedx monthly payment: %w", err)
	}

	return nil
}

// AddMessage adds a new message to the accounting system.
func (a *ActionLogger) addWingedXWeeklyPayment(ctx context.Context,
	exec boil.ContextExecutor,
	userTotals *UserTotals,
	actionInserter *InsertActionLog,
) error {
	if err := a.addWingedXPayment(ctx,
		exec,
		userTotals,
		actionInserter,
		CategoryWingedxWeeklyPayment,
		func(d *time.Time) {
			*d = d.AddDate(0, 0, CountWingedXPaymentWeek) // add 7 days
		},
	); err != nil {
		return fmt.Errorf("add wingedx weekly payment: %w", err)
	}

	return nil
}
