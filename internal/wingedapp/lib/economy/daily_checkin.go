package economy

import (
	"context"
	"fmt"
	"time"

	"wingedapp/pgtester/internal/wingedapp/lib/applog"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
)

// DailyCheckinLogic handles daily check-in business logic.
// Per spec: Check-in is a UI action only. Wings are granted ONLY at streak milestones.
type DailyCheckinLogic struct {
	logger          applog.Logger
	userTotalsStore userTotalsStorer
	actionLogStore  actionLogStorer
	transactionStore transactionStorer
}

// NewDailyCheckinLogic creates a new DailyCheckinLogic.
func NewDailyCheckinLogic(
	l applog.Logger,
	userTotalsStore userTotalsStorer,
	actionLogStore actionLogStorer,
	transactionStore transactionStorer,
) (*DailyCheckinLogic, error) {
	if l == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if userTotalsStore == nil {
		return nil, fmt.Errorf("userTotalsStore is required")
	}
	if actionLogStore == nil {
		return nil, fmt.Errorf("actionLogStore is required")
	}
	if transactionStore == nil {
		return nil, fmt.Errorf("transactionStore is required")
	}
	return &DailyCheckinLogic{
		logger:           l,
		userTotalsStore:  userTotalsStore,
		actionLogStore:   actionLogStore,
		transactionStore: transactionStore,
	}, nil
}

// PerformCheckin performs a daily check-in for the user.
// Per spec: No wings granted directly. Updates streak. Awards wings at milestones (7, 30 days).
func (d *DailyCheckinLogic) PerformCheckin(
	ctx context.Context,
	exec boil.ContextExecutor,
	userID string,
) (*CheckinResult, error) {
	// 1. Get current user totals (includes streak state)
	totals, err := d.userTotalsStore.Totals(ctx, exec, userID)
	if err != nil {
		return nil, fmt.Errorf("get totals: %w", err)
	}
	if totals == nil {
		return nil, fmt.Errorf("user totals not found for user: %s", userID)
	}

	// 2. Check if already checked in today
	if totals.StreakLastDate.Valid && isToday(totals.StreakLastDate.Time) {
		return nil, ErrAlreadyCheckedInToday
	}

	// 3. Calculate new streak
	newStreak := d.calculateNewStreak(totals)

	// 4. Update streak fields
	newLongest := max(totals.StreakLongestDays, newStreak)
	todayDate := time.Now().UTC().Truncate(24 * time.Hour)

	err = d.userTotalsStore.Update(ctx, exec, &UpdateUserTotals{
		ID:                totals.ID,
		StreakLastDate:    null.TimeFrom(todayDate),
		StreakCurrentDays: null.IntFrom(newStreak),
		StreakLongestDays: null.IntFrom(newLongest),
	})
	if err != nil {
		return nil, fmt.Errorf("update streak: %w", err)
	}

	// 5. Log the check-in action
	_, err = d.actionLogStore.Insert(ctx, exec, string(ActionDailyCheckIn), &InsertActionLog{
		UserID: userID,
		RefID:  uuid.New().String(),
		Type:   ActionDailyCheckIn,
	})
	if err != nil {
		return nil, fmt.Errorf("insert action log: %w", err)
	}

	// 6. Check milestone and award wings if reached
	result := &CheckinResult{
		NewStreak:          newStreak,
		IsNewLongestStreak: newLongest > totals.StreakLongestDays,
	}

	milestone, wings := d.checkMilestone(newStreak)
	if milestone > 0 {
		err = d.awardMilestoneWings(ctx, exec, userID, totals.ID, milestone, wings)
		if err != nil {
			return nil, fmt.Errorf("award milestone wings: %w", err)
		}
		result.MilestoneReached = true
		result.MilestoneType = milestone
		result.WingsAwarded = wings
	}

	return result, nil
}

// calculateNewStreak determines the new streak based on last check-in date.
// Per spec: yesterday = increment, missed day = reset to 1
func (d *DailyCheckinLogic) calculateNewStreak(totals *UserTotals) int {
	if !totals.StreakLastDate.Valid {
		// First check-in ever
		return 1
	}

	lastDate := totals.StreakLastDate.Time.UTC().Truncate(24 * time.Hour)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	if lastDate.Equal(yesterday) {
		// Consecutive day - increment streak
		return totals.StreakCurrentDays + 1
	}

	// Missed a day - reset to 1
	return 1
}

// checkMilestone returns (milestone, wings) if a milestone was reached, (0, 0) otherwise.
// Per spec: 7-day = +2 wings, 30-day = +6 wings (once per streak run)
func (d *DailyCheckinLogic) checkMilestone(streak int) (milestone int, wings int) {
	switch streak {
	case 7:
		return 7, Streak7DayWings
	case 30:
		return 30, Streak30DayWings
	default:
		return 0, 0
	}
}

// awardMilestoneWings grants wings for reaching a streak milestone.
func (d *DailyCheckinLogic) awardMilestoneWings(
	ctx context.Context,
	exec boil.ContextExecutor,
	userID string,
	totalsID string,
	milestone int,
	wings int,
) error {
	// Determine action type based on milestone
	var actionType ActionType
	switch milestone {
	case 7:
		actionType = ActionStreak7Day
	case 30:
		actionType = ActionStreak30Day
	default:
		return fmt.Errorf("invalid milestone: %d", milestone)
	}

	// Insert action log for milestone
	extRefID := uuid.New().String()
	actionLog, err := d.actionLogStore.Insert(ctx, exec, string(actionType), &InsertActionLog{
		UserID: userID,
		RefID:  extRefID,
		Type:   actionType,
	})
	if err != nil {
		return fmt.Errorf("insert milestone action log: %w", err)
	}

	// Insert transaction (earned wings expire in 30 days)
	err = d.transactionStore.Insert(ctx, exec, &InsertTransaction{
		UserID:       userID,
		ActionTypeID: string(actionType),
		ActionRefID:  actionLog.ID,
		WingsAmount:  wings,
		IsCredit:     true,
		Claimed:      true, // Milestone wings are immediately credited
		ExpiresAt:    null.TimeFrom(time.Now().AddDate(0, 0, EarnedWingsExpiryDays)),
	})
	if err != nil {
		return fmt.Errorf("insert milestone transaction: %w", err)
	}

	// Update user's total wings
	totals, err := d.userTotalsStore.Totals(ctx, exec, userID)
	if err != nil {
		return fmt.Errorf("get totals for wing update: %w", err)
	}

	err = d.userTotalsStore.Update(ctx, exec, &UpdateUserTotals{
		ID:    totalsID,
		Wings: null.IntFrom(totals.Wings + wings),
	})
	if err != nil {
		return fmt.Errorf("update wings balance: %w", err)
	}

	return nil
}

// GetStatus returns the current check-in status for a user.
func (d *DailyCheckinLogic) GetStatus(
	ctx context.Context,
	exec boil.ContextExecutor,
	userID string,
) (*CheckinStatus, error) {
	totals, err := d.userTotalsStore.Totals(ctx, exec, userID)
	if err != nil {
		return nil, fmt.Errorf("get totals: %w", err)
	}
	if totals == nil {
		return nil, fmt.Errorf("user totals not found for user: %s", userID)
	}

	checkedInToday := totals.StreakLastDate.Valid && isToday(totals.StreakLastDate.Time)

	// Calculate next milestone info
	nextMilestone, daysTo, milestoneWings := d.getNextMilestoneInfo(totals.StreakCurrentDays)

	return &CheckinStatus{
		CheckedInToday:    checkedInToday,
		StreakCurrentDays: totals.StreakCurrentDays,
		StreakLongestDays: totals.StreakLongestDays,
		NextMilestone:     nextMilestone,
		DaysToMilestone:   daysTo,
		MilestoneWings:    milestoneWings,
	}, nil
}

// getNextMilestoneInfo returns (milestone, daysUntil, wings) for the next milestone.
func (d *DailyCheckinLogic) getNextMilestoneInfo(currentStreak int) (milestone, daysTo, wings int) {
	if currentStreak < 7 {
		return 7, 7 - currentStreak, Streak7DayWings
	}
	if currentStreak < 30 {
		return 30, 30 - currentStreak, Streak30DayWings
	}
	// Past all milestones - no next milestone
	return 0, 0, 0
}

// isToday checks if a timestamp is from today (UTC).
func isToday(t time.Time) bool {
	now := time.Now().UTC()
	tUTC := t.UTC()
	return tUTC.Year() == now.Year() && tUTC.YearDay() == now.YearDay()
}
