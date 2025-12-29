package economy

import (
	"context"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// AddMessage adds a new message to the accounting system.
func (a *ActionLogger) addWingedPlusWeeklyPayment(ctx context.Context,
	exec boil.ContextExecutor,
	userTotals *UserTotals,
	actionInserter *InsertActionLog,
) error {
	if err := a.addWingedPayment(ctx,
		exec,
		userTotals,
		actionInserter,
		&SubscriptionPayment{
			Type: SubscriptionTypeWingedPlus,
			Name: SubscriptionPaymentWeekly,
		},
	); err != nil {
		return fmt.Errorf("winged plus weekly payment: %w", err)
	}

	return nil
}

func (a *ActionLogger) addWingedPlusMonthlyPayment(ctx context.Context,
	exec boil.ContextExecutor,
	userTotals *UserTotals,
	actionInserter *InsertActionLog,
) error {
	if err := a.addWingedPayment(
		ctx,
		exec,
		userTotals,
		actionInserter,
		&SubscriptionPayment{
			Type: SubscriptionTypeWingedPlus,
			Name: SubscriptionPaymentMonthly,
		},
	); err != nil {
		return fmt.Errorf("winged plus monthly payment: %w", err)
	}
	return nil
}

func (a *ActionLogger) addWingedPlusThreeMonthlyPayment(ctx context.Context,
	exec boil.ContextExecutor,
	userTotals *UserTotals,
	actionInserter *InsertActionLog,
) error {
	if err := a.addWingedPayment(
		ctx,
		exec,
		userTotals,
		actionInserter,
		&SubscriptionPayment{
			Type: SubscriptionTypeWingedPlus,
			Name: SubscriptionPaymentThreeMonth,
		},
	); err != nil {
		return fmt.Errorf("winged plus three monthly payment: %w", err)
	}
	return nil
}

func (a *ActionLogger) addWingedPlusSixMonthlyPayment(ctx context.Context,
	exec boil.ContextExecutor,
	userTotals *UserTotals,
	actionInserter *InsertActionLog,
) error {
	if err := a.addWingedPayment(
		ctx,
		exec,
		userTotals,
		actionInserter,
		&SubscriptionPayment{
			Type: SubscriptionTypeWingedPlus,
			Name: SubscriptionPaymentSixMonth,
		},
	); err != nil {
		return fmt.Errorf("winged plus six monthly payment: %w", err)
	}
	return nil
}

func (a *ActionLogger) addWingedPayment(ctx context.Context,
	exec boil.ContextExecutor,
	userTotals *UserTotals,
	actionInserter *InsertActionLog,
	subPayment *SubscriptionPayment,
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

	subscriptionPlan, err := a.subscriptionStorer.SubscriptionPlan(ctx, exec,
		subPayment.Type,
		subPayment.Name,
	)
	if err != nil {
		return fmt.Errorf("fetch subscription plan: %w", err)
	}

	actionLog, err := a.actionLogStorer.Insert(ctx, exec, string(actionInserter.Type), actionInserter)
	if err != nil {
		return fmt.Errorf("insert action log: %w", err)
	}

	// Subscription wings expire in 30 days (same as earned wings)
	expiresAt := null.TimeFrom(time.Now().AddDate(0, 0, EarnedWingsExpiryDays))

	if err = a.transactionStorer.Insert(ctx, exec, &InsertTransaction{
		UserID:       actionInserter.UserID,       // userID
		ActionTypeID: string(actionInserter.Type), // enum string value
		ActionRefID:  actionLog.ID,                // shared duplicate in action log insert
		WingsAmount:  subscriptionPlan.Wings,      // the number of wings to credit
		Claimed:      true,                        // yes, true by default
		IsCredit:     true,                        // from the user's perspective
		ExtraInfo:    actionInserter.JSONDetails,  // pass along any extra details
		ExpiresAt:    expiresAt,                   // 30-day expiry
	}); err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}

	/* then update totals */

	if err = a.userTotalsStorer.Update(ctx, exec, &UpdateUserTotals{
		ID:    userTotals.ID,
		Wings: null.IntFrom(userTotals.Wings + subscriptionPlan.Wings),
	}); err != nil {
		return fmt.Errorf("update user totals: %w", err)
	}

	return nil
}
