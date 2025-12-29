package economy

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

/*
	Entrypoints.go houses the main entrypoints for the economy package.
		- CreateActionLog
		- CanPerformAction
*/

type actLoggerHandlerFn func(ctx context.Context,
	exec boil.ContextExecutor,
	userTotals *UserTotals,
	inserter *InsertActionLog,
) error

// validateUserExists checks if user exists
func (a *ActionLogger) validateUserExists(ctx context.Context, exec boil.ContextExecutor, userID string) error {
	exists, err := a.userStorer.Exists(ctx, exec, userID)
	if err != nil {
		return fmt.Errorf("checking user exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("user with ID %s does not exist", userID)
	}

	return nil
}

// ensureUserHasTotals creates totals for user if not exists
func (a *ActionLogger) ensureUserHasTotals(ctx context.Context, exec boil.ContextExecutor, userID string) error {
	userTotals, err := a.userTotalsStorer.Totals(ctx, exec, userID)
	if err != nil {
		return fmt.Errorf("checking user exists: %w", err)
	}
	if userTotals != nil {
		// create totals
	}

	return nil
}

// CreateActionLog adds a new entry to action log,
// cause earning, or spending of wings currency.
func (a *ActionLogger) CreateActionLog(ctx context.Context,
	exec boil.ContextExecutor,
	inserter *InsertActionLog,
) error {
	if err := inserter.validateParams(); err != nil {
		return fmt.Errorf("param validation: %w", err)
	}

	userTotals, err := a.userTotalsStorer.Totals(ctx, exec, inserter.UserID)
	if err != nil {
		return fmt.Errorf("get user totals: %w", err)
	}

	actionLoggerHandlers := map[ActionType]actLoggerHandlerFn{
		ActionWingedXWeeklyPayment:        a.addWingedXWeeklyPayment,
		ActionWingedXMonthlyPayment:       a.addWingedXMonthlyPayment,
		ActionWingedPlusWeeklyPayment:     a.addWingedPlusWeeklyPayment,
		ActionWingedPlusMonthlyPayment:    a.addWingedPlusMonthlyPayment,
		ActionWingedPlusThreeMonthPayment: a.addWingedPlusThreeMonthlyPayment,
		ActionWingedPlusSixMonthPayment:   a.addWingedPlusSixMonthlyPayment,
		ActionReferralComplete:            a.processReferralBonus,  // Referrer gets 4 wings on invitee's first paid action
		ActionAttendDate:                  a.processAttendDate,     // When user confirms they attended a date
		ActionSendMessage:                 a.processSendMessage,    // Deduct 1 wing per 5 messages sent
	}

	var handler actLoggerHandlerFn
	if handler = actionLoggerHandlers[inserter.Type]; handler == nil {
		return errors.Join(ErrInvalidEntryType, errInvalidAction(inserter.Type))
	}

	if err = handler(ctx, exec, userTotals, inserter); err != nil {
		return fmt.Errorf("insert actionLog err: %s, %w", inserter.Type, err)
	}

	return nil
}

// isPremiumActive checks if the user has an active premium subscription.
// Used to bypass wings checks and short-circuit spend recording.
func isPremiumActive(userTotals *UserTotals) bool {
	return userTotals != nil &&
		userTotals.PremiumExpiresIn.Valid &&
		userTotals.PremiumExpiresIn.Time.After(time.Now())
}

// CanPerformAction checks if the action can be performed based on user's wings balance.
// Returns true if action is allowed, false with appropriate error if not.
// Premium subscribers with active subscription bypass wings requirements.
func (a *ActionLogger) CanPerformAction(ctx context.Context,
	exec boil.ContextExecutor,
	params *CanPerformActionParams,
) (bool, error) {
	userTotals, err := a.userTotalsStorer.Totals(ctx, exec, params.UserID)
	if err != nil {
		return false, fmt.Errorf("get user totals: %w", err)
	}

	// User has no totals record yet - they have 0 wings
	if userTotals == nil {
		return a.canPerformWithBalance(params.ActionType, 0)
	}

	// Premium subscribers bypass wings check
	if isPremiumActive(userTotals) {
		return true, nil
	}

	return a.canPerformWithBalance(params.ActionType, userTotals.Wings)
}

// canPerformWithBalance checks if an action can be performed given a wings balance.
func (a *ActionLogger) canPerformWithBalance(actionType ActionType, wings int) (bool, error) {
	switch actionType {
	case ActionSendMessage:
		// Need at least 1 wing to send messages (will be deducted every 5 msgs)
		if wings < SendMessageWingsCost {
			return false, ErrInsufficientWings
		}
		return true, nil
	default:
		// For earning actions (payments, referrals, etc.), always allow
		return true, nil
	}
}

// actionTypeToSubscription maps an ActionType to subscription type and plan name.
func actionTypeToSubscription(actionType ActionType) (subscriptionType, planName string) {
	switch actionType {
	case ActionWingedPlusWeeklyPayment:
		return SubscriptionTypeWingedPlus, SubscriptionPaymentWeekly
	case ActionWingedPlusMonthlyPayment:
		return SubscriptionTypeWingedPlus, SubscriptionPaymentMonthly
	case ActionWingedPlusThreeMonthPayment:
		return SubscriptionTypeWingedPlus, SubscriptionPaymentThreeMonth
	case ActionWingedPlusSixMonthPayment:
		return SubscriptionTypeWingedPlus, SubscriptionPaymentSixMonth
	case ActionWingedXWeeklyPayment:
		return SubscriptionTypeWingedX, SubscriptionPaymentWeekly
	case ActionWingedXMonthlyPayment:
		return SubscriptionTypeWingedX, SubscriptionPaymentMonthly
	default:
		return "", ""
	}
}

// DeleteActionLog performs a lazy delete on ActionLog.
func (a *ActionLogger) DeleteActionLog(ctx context.Context,
	exec boil.ContextExecutor,
	uuid string,
) error {
	actionLog, err := a.actionLogStorer.ActionLog(ctx, exec, &QueryFilterActionLog{
		ID:       null.StringFrom(uuid),
		IsActive: null.IntFrom(1),
	})
	if err != nil {
		return fmt.Errorf("fetch action log: %w", err)
	}

	// void action log
	if err = a.voidActionLog(ctx, exec, actionLog.ID); err != nil {
		return fmt.Errorf("void action log: %w", err)
	}

	transactions, err := a.transactionStorer.Transactions(ctx, exec, &QueryFilterTransactions{
		ActionLogID: null.StringFrom(actionLog.ID),
		IsActive:    null.IntFrom(1),
	})
	if err != nil {
		return fmt.Errorf("fetch transaction: %w", err)
	}

	if len(transactions) == 1 {
		if err = a.handleDeleteTransaction(ctx, exec, &transactions[0]); err != nil {
			return fmt.Errorf("handle delete transaction: %w", err)
		}
	}

	/* to scale: keep adding more tables to void here */

	return nil
}

// handleDeleteTransaction handles the voiding of a transaction and its effects
func (a *ActionLogger) handleDeleteTransaction(ctx context.Context,
	exec boil.ContextExecutor,
	transaction *Transaction,
) error {
	// void transactions
	if err := a.voidTransaction(ctx, exec, transaction.ID); err != nil {
		return fmt.Errorf("void transaction: %w", err)
	}

	// void (revert) user total wings balance
	userTotals, err := a.userTotalsStorer.Totals(ctx, exec, transaction.UserID)
	if err != nil {
		return fmt.Errorf("fetch user totals: %w", err)
	}
	newBalance := userTotals.Wings
	if transaction.IsCredit {
		newBalance -= transaction.Amount
	} else {
		newBalance += transaction.Amount
	}
	if err = a.userTotalsStorer.Update(ctx, exec, &UpdateUserTotals{
		ID:    userTotals.ID,
		Wings: null.IntFrom(newBalance),
	}); err != nil {
		return fmt.Errorf("update user totals: %w", err)
	}

	return nil
}

func (a *ActionLogger) voidActionLog(ctx context.Context, exec boil.ContextExecutor, uuid string) error {
	if err := a.actionLogStorer.Delete(ctx, exec, uuid); err != nil {
		return fmt.Errorf("delete action log: %w", err)
	}

	return nil
}

func (a *ActionLogger) voidTransaction(ctx context.Context, exec boil.ContextExecutor, uuid string) error {
	if err := a.transactionStorer.Delete(ctx, exec, uuid); err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}

	return nil
}

