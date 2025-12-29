package scheduling

import (
	"context"
	"errors"
	"fmt"

	schedulingLib "wingedapp/pgtester/internal/wingedapp/lib/scheduling"
)

// LogisticsArrived signals that user has arrived.
func (b *Business) LogisticsArrived(
	ctx context.Context,
	params *schedulingLib.LogisticsArrivedParams,
) (*schedulingLib.LogisticsArrivedResult, error) {
	if b.logisticsExecutor == nil {
		return nil, errors.New("logistics executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.logisticsExecutor.LogisticsArrived(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("logistics arrived: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// LogisticsRunningLate signals that user is running late.
func (b *Business) LogisticsRunningLate(
	ctx context.Context,
	params *schedulingLib.LogisticsRunningLateParams,
) (*schedulingLib.LogisticsRunningLateResult, error) {
	if b.logisticsExecutor == nil {
		return nil, errors.New("logistics executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.logisticsExecutor.LogisticsRunningLate(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("logistics running late: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// LogisticsNeedHelp signals that user needs help.
func (b *Business) LogisticsNeedHelp(
	ctx context.Context,
	params *schedulingLib.LogisticsNeedHelpParams,
) (*schedulingLib.LogisticsNeedHelpResult, error) {
	if b.logisticsExecutor == nil {
		return nil, errors.New("logistics executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.logisticsExecutor.LogisticsNeedHelp(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("logistics need help: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// LogisticsCancelInWindow cancels the date during active window.
func (b *Business) LogisticsCancelInWindow(
	ctx context.Context,
	params *schedulingLib.LogisticsCancelInWindowParams,
) (*schedulingLib.LogisticsCancelInWindowResult, error) {
	if b.logisticsExecutor == nil {
		return nil, errors.New("logistics executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.logisticsExecutor.LogisticsCancelInWindow(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("logistics cancel in window: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}
