package scheduling

import (
	"context"
	"errors"
	"fmt"

	schedulingLib "wingedapp/pgtester/internal/wingedapp/lib/scheduling"

	"github.com/google/uuid"
)

type Business struct {
	transactor           transactor
	availabilityGetter   availabilityGetter
	availabilitySyncer   availabilitySyncer
	overlapFinder        overlapFinder
	dateInstanceFetcher  dateInstanceFetcher
	timeFlowExecutor     timeFlowExecutor
	venueFlowExecutor    venueFlowExecutor
	venueSuggestionExec  venueSuggestionExecutor
	modificationExecutor modificationExecutor
	feedbackExecutor     feedbackExecutor
	logisticsExecutor    logisticsExecutor
	actionLogger         actionLogger
}

func NewBusiness(
	transactor transactor,
	availabilityGetter availabilityGetter,
	availabilitySyncer availabilitySyncer,
	overlapFinder overlapFinder,
	dateInstanceFetcher dateInstanceFetcher,
	timeFlowExecutor timeFlowExecutor,
	venueFlowExecutor venueFlowExecutor,
	venueSuggestionExec venueSuggestionExecutor,
	modificationExecutor modificationExecutor,
	feedbackExecutor feedbackExecutor,
	logisticsExecutor logisticsExecutor,
	actionLogger actionLogger,
) (*Business, error) {
	if transactor == nil {
		return nil, errors.New("transactor is required")
	}
	if availabilityGetter == nil {
		return nil, errors.New("availabilityGetter is required")
	}
	if availabilitySyncer == nil {
		return nil, errors.New("availabilitySyncer is required")
	}
	if overlapFinder == nil {
		return nil, errors.New("overlapFinder is required")
	}
	// All other executors are optional for backwards compatibility

	return &Business{
		transactor:           transactor,
		availabilityGetter:   availabilityGetter,
		availabilitySyncer:   availabilitySyncer,
		overlapFinder:        overlapFinder,
		dateInstanceFetcher:  dateInstanceFetcher,
		timeFlowExecutor:     timeFlowExecutor,
		venueFlowExecutor:    venueFlowExecutor,
		venueSuggestionExec:  venueSuggestionExec,
		modificationExecutor: modificationExecutor,
		feedbackExecutor:     feedbackExecutor,
		logisticsExecutor:    logisticsExecutor,
		actionLogger:         actionLogger,
	}, nil
}

// UserAvailabilities returns all availability time blocks for a user.
func (b *Business) UserAvailabilities(ctx context.Context, userID string) ([]schedulingLib.TimeBlock, error) {
	exec := b.transactor.DB()
	blocks, err := b.availabilityGetter.UserTimeBlocks(ctx, exec, userID)
	if err != nil {
		return nil, fmt.Errorf("get user availability: %w", err)
	}
	return blocks, nil
}

// SyncUserAvailability replaces all existing availability for a user with new blocks.
func (b *Business) SyncUserAvailability(ctx context.Context, params *schedulingLib.SyncUserAvailabilityParams) (*schedulingLib.SyncUserAvailabilityResult, error) {
	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.availabilitySyncer.SyncUserAvailability(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("sync user availability: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// FindOverlaps finds overlapping availability between two users.
func (b *Business) FindOverlaps(ctx context.Context, userAID, userBID string) ([]schedulingLib.TimeBlock, error) {
	exec := b.transactor.DB()
	blocks, err := b.overlapFinder.FindOverlaps(ctx, exec, userAID, userBID)
	if err != nil {
		return nil, fmt.Errorf("find overlaps: %w", err)
	}
	return blocks, nil
}

// UserDateInstance returns a single date instance with UI state for the authenticated user.
func (b *Business) UserDateInstance(
	ctx context.Context,
	dateInstanceID uuid.UUID,
	requestingUserID uuid.UUID,
) (*schedulingLib.DateInstanceUI, error) {
	if b.dateInstanceFetcher == nil {
		return nil, errors.New("date instance fetcher not configured")
	}
	exec := b.transactor.DB()
	return b.dateInstanceFetcher.DateInstanceForUser(ctx, exec, dateInstanceID, requestingUserID)
}

// UserDateInstances returns paginated date instances for the authenticated user.
func (b *Business) UserDateInstances(
	ctx context.Context,
	filter *schedulingLib.QueryFilterUserDateInstances,
) (*schedulingLib.DateInstanceUIPaginated, error) {
	if b.dateInstanceFetcher == nil {
		return nil, errors.New("date instance fetcher not configured")
	}
	exec := b.transactor.DB()
	return b.dateInstanceFetcher.DateInstancesForUser(ctx, exec, filter)
}

// ============================================================================
// SWAPPABLE SETTERS (for testing)
// ============================================================================

// SetTimeFlowExecutor swaps the time flow executor (for testing).
func (b *Business) SetTimeFlowExecutor(e timeFlowExecutor) {
	b.timeFlowExecutor = e
}

// SetVenueFlowExecutor swaps the venue flow executor (for testing).
func (b *Business) SetVenueFlowExecutor(e venueFlowExecutor) {
	b.venueFlowExecutor = e
}

// SetVenueSuggestionExecutor swaps the venue suggestion executor (for testing).
func (b *Business) SetVenueSuggestionExecutor(e venueSuggestionExecutor) {
	b.venueSuggestionExec = e
}

// SetModificationExecutor swaps the modification executor (for testing).
func (b *Business) SetModificationExecutor(e modificationExecutor) {
	b.modificationExecutor = e
}

// SetFeedbackExecutor swaps the feedback executor (for testing).
func (b *Business) SetFeedbackExecutor(e feedbackExecutor) {
	b.feedbackExecutor = e
}

// SetLogisticsExecutor swaps the logistics executor (for testing).
func (b *Business) SetLogisticsExecutor(e logisticsExecutor) {
	b.logisticsExecutor = e
}

// SetDateInstanceFetcher swaps the date instance fetcher (for testing).
func (b *Business) SetDateInstanceFetcher(f dateInstanceFetcher) {
	b.dateInstanceFetcher = f
}

// SetActionLogger swaps the action logger (for testing).
func (b *Business) SetActionLogger(l actionLogger) {
	b.actionLogger = l
}
