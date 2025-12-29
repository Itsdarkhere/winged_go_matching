package scheduling

import (
	"context"
	"errors"
	"fmt"

	schedulingLib "wingedapp/pgtester/internal/wingedapp/lib/scheduling"
)

// VenueOptions returns venue options for the initiator.
func (b *Business) VenueOptions(
	ctx context.Context,
	params *schedulingLib.VenueOptionsParams,
) (*schedulingLib.VenueOptionsResult, error) {
	if b.venueFlowExecutor == nil {
		return nil, errors.New("venue flow executor not configured")
	}
	exec := b.transactor.DB()
	return b.venueFlowExecutor.VenueOptions(ctx, exec, params)
}

// SelectVenue allows the initiator to select a venue.
func (b *Business) SelectVenue(
	ctx context.Context,
	params *schedulingLib.SelectVenueParams,
) (*schedulingLib.SelectVenueResult, error) {
	if b.venueFlowExecutor == nil {
		return nil, errors.New("venue flow executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.venueFlowExecutor.SelectVenue(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("select venue: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// ConfirmBooking allows the initiator to confirm booking status.
func (b *Business) ConfirmBooking(
	ctx context.Context,
	params *schedulingLib.ConfirmBookingParams,
) (*schedulingLib.ConfirmBookingResult, error) {
	if b.venueFlowExecutor == nil {
		return nil, errors.New("venue flow executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.venueFlowExecutor.ConfirmBooking(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("confirm booking: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// RequestVenueChange allows the receiver to request a venue change.
func (b *Business) RequestVenueChange(
	ctx context.Context,
	params *schedulingLib.RequestVenueChangeParams,
) (*schedulingLib.RequestVenueChangeResult, error) {
	if b.venueFlowExecutor == nil {
		return nil, errors.New("venue flow executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.venueFlowExecutor.RequestVenueChange(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("request venue change: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// SuggestVenue allows the receiver to suggest their own venue.
func (b *Business) SuggestVenue(
	ctx context.Context,
	params *schedulingLib.SuggestVenueParams,
) (*schedulingLib.SuggestVenueResult, error) {
	if b.venueSuggestionExec == nil {
		return nil, errors.New("venue suggestion executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.venueSuggestionExec.SuggestVenue(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("suggest venue: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// RespondToVenueSuggestion allows the initiator to respond to a venue suggestion.
func (b *Business) RespondToVenueSuggestion(
	ctx context.Context,
	params *schedulingLib.RespondToVenueSuggestionParams,
) (*schedulingLib.RespondToVenueSuggestionResult, error) {
	if b.venueSuggestionExec == nil {
		return nil, errors.New("venue suggestion executor not configured")
	}

	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.venueSuggestionExec.RespondToVenueSuggestion(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("respond to venue suggestion: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}
