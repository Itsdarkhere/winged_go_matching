package scheduling

import (
	"context"
	"errors"
	"fmt"

	schedulingLib "wingedapp/pgtester/internal/wingedapp/lib/scheduling"
)

// ChangeTime allows either user to request time change.
func (b *Business) ChangeTime(
	ctx context.Context,
	params *schedulingLib.ChangeTimeParams,
) (*schedulingLib.ChangeTimeResult, error) {
	if b.modificationExecutor == nil {
		return nil, errors.New("modification executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.modificationExecutor.ChangeTime(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("change time: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// ChangePlace allows either user to request place change.
func (b *Business) ChangePlace(
	ctx context.Context,
	params *schedulingLib.ChangePlaceParams,
) (*schedulingLib.ChangePlaceResult, error) {
	if b.modificationExecutor == nil {
		return nil, errors.New("modification executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.modificationExecutor.ChangePlace(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("change place: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// CancelDate allows either user to cancel the date.
func (b *Business) CancelDate(
	ctx context.Context,
	params *schedulingLib.CancelDateParams,
) (*schedulingLib.CancelDateResult, error) {
	if b.modificationExecutor == nil {
		return nil, errors.New("modification executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.modificationExecutor.CancelDate(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("cancel date: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}
