package scheduling

import (
	"context"
	"errors"
	"fmt"

	"wingedapp/pgtester/internal/wingedapp/lib/economy"
	schedulingLib "wingedapp/pgtester/internal/wingedapp/lib/scheduling"
)

// PendingFeedback returns any date instance requiring feedback.
func (b *Business) PendingFeedback(
	ctx context.Context,
	params *schedulingLib.PendingFeedbackParams,
) (*schedulingLib.PendingFeedbackResult, error) {
	if b.feedbackExecutor == nil {
		return nil, errors.New("feedback executor not configured")
	}
	exec := b.transactor.DB()
	return b.feedbackExecutor.PendingFeedback(ctx, exec, params)
}

// SubmitDidYouMeet allows user to submit post-date meeting feedback.
func (b *Business) SubmitDidYouMeet(
	ctx context.Context,
	params *schedulingLib.SubmitDidYouMeetParams,
) (*schedulingLib.SubmitDidYouMeetResult, error) {
	if b.feedbackExecutor == nil {
		return nil, errors.New("feedback executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.feedbackExecutor.SubmitDidYouMeet(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("submit did you meet: %w", err)
	}

	// Award wings for attending the date (only when did_meet is "yes")
	if params.DidMeet == "yes" {
		if err := b.actionLogger.CreateActionLog(ctx, tx, &economy.InsertActionLog{
			UserID: params.RequestingUserID.String(),
			RefID:  params.DateInstanceID.String(),
			Type:   economy.ActionAttendDate,
		}); err != nil {
			return nil, fmt.Errorf("attend date bonus: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// SubmitDecision allows user to submit second-date decision.
func (b *Business) SubmitDecision(
	ctx context.Context,
	params *schedulingLib.SubmitDecisionParams,
) (*schedulingLib.SubmitDecisionResult, error) {
	if b.feedbackExecutor == nil {
		return nil, errors.New("feedback executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.feedbackExecutor.SubmitDecision(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("submit decision: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}
