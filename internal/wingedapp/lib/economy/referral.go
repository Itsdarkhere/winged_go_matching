package economy

import (
	"context"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// processReferralBonus handles crediting wings to the referrer only
// when a referred user performs their first paid action (connect or schedule).
// Per MVP spec: +4 wings to referrer, invitee gets nothing.
// actionInserter.UserID = invitee (the one who was referred)
// actionInserter.RefID = invite_code_id (optional - will be looked up if empty)
//
// Business layer just calls: CreateActionLog(ctx, exec, &InsertActionLog{Type: ActionReferralComplete, UserID: userID})
// This handler does ALL the work: user lookup, check if referred, idempotency, credit referrer.
func (a *ActionLogger) processReferralBonus(ctx context.Context,
	exec boil.ContextExecutor,
	_ *UserTotals, // userTotals for invitee - not used, we only credit referrer
	actionInserter *InsertActionLog,
) error {
	// 1. If RefID not provided, look up user to get invite_code_ref_id
	if actionInserter.RefID == "" {
		user, err := a.userStorer.User(ctx, exec, &QueryFilterUser{
			ID: null.StringFrom(actionInserter.UserID),
		})
		if err != nil {
			return fmt.Errorf("fetch user: %w", err)
		}
		if user == nil {
			return nil // User not found, nothing to do
		}
		if !user.UserInviteCodeRefID.Valid {
			return nil // User wasn't referred, nothing to do
		}
		actionInserter.RefID = user.UserInviteCodeRefID.String
	}

	// 2. Check idempotency - skip if already processed for this invitee+invite_code
	existingLogs, err := a.actionLogStorer.ActionLogs(ctx, exec, &QueryFilterActionLog{
		UserID:   null.StringFrom(actionInserter.UserID),
		Category: null.StringFrom(string(ActionReferralComplete)),
		RefID:    null.StringFrom(actionInserter.RefID),
		IsActive: null.IntFrom(1),
	})
	if err != nil {
		return fmt.Errorf("check idempotency: %w", err)
	}
	if len(existingLogs) > 0 {
		// Already processed, return silently
		return nil
	}

	// 4. Load invite code to get referrer info
	inviteCode, err := a.inviteCodeStorer.InviteCode(ctx, exec, &QueryFilterInviteCode{
		ID: null.StringFrom(actionInserter.RefID),
	})
	if err != nil {
		return fmt.Errorf("fetch invite code: %w", err)
	}
	if inviteCode == nil {
		return fmt.Errorf("invite code not found: %s", actionInserter.RefID)
	}

	// 5. Check if has referrer - system-generated codes have no referrer
	if !inviteCode.ReferrerNumberHash.Valid {
		// No referrer for this invite code, nothing to credit
		// Still insert action log to mark as processed (idempotency)
		_, err := a.actionLogStorer.Insert(ctx, exec, string(ActionReferralComplete), actionInserter)
		if err != nil {
			return fmt.Errorf("insert action log (no referrer): %w", err)
		}
		return nil
	}

	// 6. Look up referrer by mobile hash
	referrer, err := a.userStorer.User(ctx, exec, &QueryFilterUser{
		Sha256Hash: inviteCode.ReferrerNumberHash,
	})
	if err != nil {
		return fmt.Errorf("referrer lookup: %w", err)
	}
	if referrer == nil {
		// Referrer account doesn't exist (deleted?), mark as processed and return nil
		_, err := a.actionLogStorer.Insert(ctx, exec, string(ActionReferralComplete), actionInserter)
		if err != nil {
			return fmt.Errorf("insert action log (referrer not found): %w", err)
		}
		return nil
	}

	// 7. Credit referrer only (per MVP spec, invitee gets nothing)
	if err := a.creditReferrer(ctx, exec, referrer.ID, actionInserter.RefID); err != nil {
		return fmt.Errorf("credit referrer: %w", err)
	}

	// 8. Insert action log for invitee to mark as processed (idempotency)
	if _, err := a.actionLogStorer.Insert(ctx, exec, string(ActionReferralComplete), actionInserter); err != nil {
		return fmt.Errorf("insert invitee action log: %w", err)
	}

	return nil
}

// creditReferrer credits wings to the referrer
func (a *ActionLogger) creditReferrer(ctx context.Context,
	exec boil.ContextExecutor,
	referrerID string,
	inviteCodeID string,
) error {
	// Check idempotency - skip if referrer already credited for this invite code
	existingLogs, err := a.actionLogStorer.ActionLogs(ctx, exec, &QueryFilterActionLog{
		UserID:   null.StringFrom(referrerID),
		Category: null.StringFrom(string(ActionReferralComplete)),
		RefID:    null.StringFrom(inviteCodeID),
		IsActive: null.IntFrom(1),
	})
	if err != nil {
		return fmt.Errorf("check referrer idempotency: %w", err)
	}
	if len(existingLogs) > 0 {
		// Already credited, return silently
		return nil
	}

	// Get or create referrer totals
	referrerTotals, err := a.userTotalsStorer.Totals(ctx, exec, referrerID)
	if err != nil {
		return fmt.Errorf("fetch referrer totals: %w", err)
	}
	if referrerTotals == nil {
		referrerTotals, err = a.userTotalsStorer.Create(ctx, exec, referrerID)
		if err != nil {
			return fmt.Errorf("create referrer totals: %w", err)
		}
	}

	// Insert action log for referrer
	actionLog, err := a.actionLogStorer.Insert(ctx, exec, string(ActionReferralComplete), &InsertActionLog{
		UserID: referrerID,
		RefID:  inviteCodeID,
		Type:   ActionReferralComplete,
	})
	if err != nil {
		return fmt.Errorf("insert referrer action log: %w", err)
	}

	// Insert transaction (earned wings expire in 30 days)
	if err := a.transactionStorer.Insert(ctx, exec, &InsertTransaction{
		UserID:       referrerID,
		ActionTypeID: string(ActionReferralComplete),
		ActionRefID:  actionLog.ID,
		WingsAmount:  ReferralBonusWings,
		Claimed:      true,
		IsCredit:     true,
		ExpiresAt:    null.TimeFrom(time.Now().AddDate(0, 0, EarnedWingsExpiryDays)),
	}); err != nil {
		return fmt.Errorf("insert referrer transaction: %w", err)
	}

	// Update totals
	newWings := referrerTotals.Wings + ReferralBonusWings
	if err := a.userTotalsStorer.Update(ctx, exec, &UpdateUserTotals{
		ID:    referrerTotals.ID,
		Wings: null.IntFrom(newWings),
	}); err != nil {
		return fmt.Errorf("update referrer wings: %w", err)
	}

	return nil
}
