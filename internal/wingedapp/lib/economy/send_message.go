package economy

import (
	"context"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// processSendMessage handles deducting wings when a user sends messages.
// Every SendMessageThreshold (5) messages costs SendMessageWingsCost (1) wing.
// Premium subscribers short-circuit entirely - no records created.
// actionInserter.RefID should be the message_id.
func (a *ActionLogger) processSendMessage(ctx context.Context,
	exec boil.ContextExecutor,
	userTotals *UserTotals,
	actionInserter *InsertActionLog,
) error {
	// 1. Premium subscribers skip spend recording entirely
	if isPremiumActive(userTotals) {
		return nil
	}

	// 2. Check idempotency - skip if already processed for this message
	existingLogs, err := a.actionLogStorer.ActionLogs(ctx, exec, &QueryFilterActionLog{
		UserID:   null.StringFrom(actionInserter.UserID),
		Category: null.StringFrom(string(ActionSendMessage)),
		RefID:    null.StringFrom(actionInserter.RefID),
		IsActive: null.IntFrom(1),
	})
	if err != nil {
		return fmt.Errorf("check idempotency: %w", err)
	}
	if len(existingLogs) > 0 {
		return nil // Already processed
	}

	// 3. Get or create user totals
	if userTotals == nil {
		userTotals, err = a.userTotalsStorer.Create(ctx, exec, actionInserter.UserID)
		if err != nil {
			return fmt.Errorf("create user totals: %w", err)
		}
	}

	// 4. Check if user has enough wings
	if userTotals.Wings < SendMessageWingsCost {
		return ErrInsufficientWings
	}

	// 5. Insert action log
	actionLog, err := a.actionLogStorer.Insert(ctx, exec, string(ActionSendMessage), &InsertActionLog{
		UserID: actionInserter.UserID,
		RefID:  actionInserter.RefID,
		Type:   ActionSendMessage,
	})
	if err != nil {
		return fmt.Errorf("insert action log: %w", err)
	}

	// 6. Increment sent messages counter
	newSentMessages := userTotals.SentMessages + 1

	// 7. Check if we hit the threshold - deduct wing
	shouldDeductWing := newSentMessages%SendMessageThreshold == 0
	newWings := userTotals.Wings
	if shouldDeductWing {
		newWings -= SendMessageWingsCost

		// 8. Insert debit transaction only when wing is deducted
		if err := a.transactionStorer.Insert(ctx, exec, &InsertTransaction{
			UserID:       actionInserter.UserID,
			ActionTypeID: string(ActionSendMessage),
			ActionRefID:  actionLog.ID,
			WingsAmount:  SendMessageWingsCost,
			Claimed:      true,
			IsCredit:     false, // debit
		}); err != nil {
			return fmt.Errorf("insert transaction: %w", err)
		}
	}

	// 9. Update user totals (always update sent_messages, conditionally update wings)
	updateTotals := &UpdateUserTotals{
		ID:           userTotals.ID,
		SentMessages: null.IntFrom(newSentMessages),
	}
	if shouldDeductWing {
		updateTotals.Wings = null.IntFrom(newWings)
	}
	if err := a.userTotalsStorer.Update(ctx, exec, updateTotals); err != nil {
		return fmt.Errorf("update user totals: %w", err)
	}

	return nil
}
