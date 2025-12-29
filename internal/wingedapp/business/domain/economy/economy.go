package economy

import (
	"context"
	"errors"
	"fmt"

	economyLib "wingedapp/pgtester/internal/wingedapp/lib/economy"
)

type Business struct {
	transactor       transactor
	checkinPerformer checkinPerformer
	actionLogger     actionLogger
}

func NewBusiness(
	transactor transactor,
	checkinPerformer checkinPerformer,
	actionLogger actionLogger,
) (*Business, error) {
	if transactor == nil {
		return nil, errors.New("transactor is required")
	}
	if checkinPerformer == nil {
		return nil, errors.New("checkinPerformer is required")
	}
	if actionLogger == nil {
		return nil, errors.New("actionLogger is required")
	}

	return &Business{
		transactor:       transactor,
		checkinPerformer: checkinPerformer,
		actionLogger:     actionLogger,
	}, nil
}

// DailyCheckin performs a daily check-in for the user.
// Per spec: Check-in is a UI action only. No wings granted directly.
// Wings are awarded ONLY at streak milestones (7-day: +2, 30-day: +6).
// Idempotent: returns success even if already checked in today.
func (b *Business) DailyCheckin(ctx context.Context, userID string) (*CheckinResponse, error) {
	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	result, err := b.checkinPerformer.PerformCheckin(ctx, tx, userID)
	if err != nil {
		// Idempotent: already checked in is not an error
		if errors.Is(err, economyLib.ErrAlreadyCheckedInToday) {
			return &CheckinResponse{
				Success:          true,
				AlreadyCheckedIn: true,
				Message:          "Already checked in today",
			}, nil
		}
		return nil, fmt.Errorf("perform checkin: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	// Build response message
	message := fmt.Sprintf("Check-in successful! Streak: %d days", result.NewStreak)
	if result.MilestoneReached {
		message = fmt.Sprintf("🎉 %d-day streak milestone! You earned %d wings!", result.MilestoneType, result.WingsAwarded)
	}

	return &CheckinResponse{
		Success:            true,
		NewStreak:          result.NewStreak,
		IsNewLongestStreak: result.IsNewLongestStreak,
		MilestoneReached:   result.MilestoneReached,
		MilestoneType:      result.MilestoneType,
		WingsAwarded:       result.WingsAwarded,
		AlreadyCheckedIn:   false,
		Message:            message,
	}, nil
}

// GetCheckinStatus returns the current check-in status for a user.
func (b *Business) GetCheckinStatus(ctx context.Context, userID string) (*CheckinStatusResponse, error) {
	status, err := b.checkinPerformer.GetStatus(ctx, b.transactor.DB(), userID)
	if err != nil {
		return nil, fmt.Errorf("get status: %w", err)
	}

	return &CheckinStatusResponse{
		CheckedInToday:    status.CheckedInToday,
		StreakCurrentDays: status.StreakCurrentDays,
		StreakLongestDays: status.StreakLongestDays,
		NextMilestone:     status.NextMilestone,
		DaysToMilestone:   status.DaysToMilestone,
		MilestoneWings:    status.MilestoneWings,
	}, nil
}

// ProcessRevenueCatEvent processes a RevenueCat webhook event.
// Uses existing ActionLogger.CreateActionLog flow - no duplicate logic.
func (b *Business) ProcessRevenueCatEvent(ctx context.Context, req *RevenueCatWebhookRequest) (*RevenueCatWebhookResponse, error) {
	event := req.Event

	// Only handle purchase/renewal events
	if event.Type != "INITIAL_PURCHASE" && event.Type != "RENEWAL" {
		return &RevenueCatWebhookResponse{
			Success: true,
			EventID: event.ID,
			Action:  "ignored",
			Reason:  "event_type_not_handled",
		}, nil
	}

	// Map product_id to action type
	actionType, ok := economyLib.ProductIDToActionType[event.ProductID]
	if !ok {
		return &RevenueCatWebhookResponse{
			Success: true,
			EventID: event.ID,
			Action:  "ignored",
			Reason:  "unknown_product_id",
		}, nil
	}

	// Process payment using existing CreateActionLog flow
	// Idempotency is handled inside the handler - duplicate RefID is a no-op
	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	// Create action log - handles action log → transaction → user totals
	// Handler has built-in idempotency (checks RefID, returns nil if exists)
	if err := b.actionLogger.CreateActionLog(ctx, tx, &economyLib.InsertActionLog{
		UserID: event.AppUserID,
		RefID:  event.ID, // webhook event ID for idempotency
		Type:   actionType,
	}); err != nil {
		return nil, fmt.Errorf("create action log: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &RevenueCatWebhookResponse{
		Success: true,
		EventID: event.ID,
		Action:  "processed",
	}, nil
}
