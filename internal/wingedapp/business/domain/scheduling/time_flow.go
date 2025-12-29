package scheduling

import (
	"context"
	"errors"
	"fmt"

	schedulingLib "wingedapp/pgtester/internal/wingedapp/lib/scheduling"
)

// SuggestDateInstanceTimes allows the receiver to suggest proposed times.
func (b *Business) SuggestDateInstanceTimes(
	ctx context.Context,
	params *schedulingLib.SuggestDateInstanceTimesParams,
) (*schedulingLib.SuggestDateInstanceTimesResult, error) {
	if b.timeFlowExecutor == nil {
		return nil, errors.New("time flow executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.timeFlowExecutor.SuggestDateInstanceTimes(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("suggest times: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// RequestMoreTimes allows the receiver to request more time options.
func (b *Business) RequestMoreTimes(
	ctx context.Context,
	params *schedulingLib.RequestMoreTimesParams,
) (*schedulingLib.RequestMoreTimesResult, error) {
	if b.timeFlowExecutor == nil {
		return nil, errors.New("time flow executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.timeFlowExecutor.RequestMoreTimes(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("request more times: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// ConfirmDateInstanceTime allows the initiator to confirm a selected time.
func (b *Business) ConfirmDateInstanceTime(
	ctx context.Context,
	params *schedulingLib.ConfirmDateInstanceTimeParams,
) (*schedulingLib.ConfirmDateInstanceTimeResult, error) {
	if b.timeFlowExecutor == nil {
		return nil, errors.New("time flow executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.timeFlowExecutor.ConfirmDateInstanceTime(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("confirm time: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// RejectDateInstanceTimes allows the initiator to reject all proposed times.
func (b *Business) RejectDateInstanceTimes(
	ctx context.Context,
	params *schedulingLib.RejectDateInstanceTimesParams,
) (*schedulingLib.RejectDateInstanceTimesResult, error) {
	if b.timeFlowExecutor == nil {
		return nil, errors.New("time flow executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.timeFlowExecutor.RejectDateInstanceTimes(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("reject times: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}
