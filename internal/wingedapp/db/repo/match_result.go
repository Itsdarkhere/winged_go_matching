package repo

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/enums"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

type InsertMatchResult struct {
	/* required FKs */
	MatchSetRefID        string
	MatchLifecycleStatus string // String enum (optional at insert, defaults to empty)
	InitiatorUserRefID   string // initiator of the match
	ReceiverUserRefID    string // receiver of the match
	InitiatorAction          string // String enum - defaults to Pending
	ReceiverAction          string // String enum - defaults to Pending
}

func (s *Store) InsertMatchResult(
	ctx context.Context,
	exec boil.ContextExecutor,
	inserter *InsertMatchResult,
) (*pgmodel.MatchResult, error) {
	/* guard */
	if inserter.MatchSetRefID == "" {
		return nil, fmt.Errorf("match_set_ref_id is required")
	}
	if inserter.InitiatorUserRefID == "" {
		return nil, fmt.Errorf("initiator_user_ref_id is required")
	}
	if inserter.ReceiverUserRefID == "" {
		return nil, fmt.Errorf("receiver_user_ref_id is required")
	}

	/* Set default actions to Pending if not provided */
	userAAction := inserter.InitiatorAction
	if userAAction == "" {
		userAAction = string(enums.MatchUserActionPending)
	}
	userBAction := inserter.ReceiverAction
	if userBAction == "" {
		userBAction = string(enums.MatchUserActionPending)
	}

	/* only insert fields from the initial set */
	matchSet := pgmodel.MatchResult{
		MatchSetRefID:      inserter.MatchSetRefID,
		InitiatorUserRefID: inserter.InitiatorUserRefID,
		ReceiverUserRefID:  inserter.ReceiverUserRefID,
		InitiatorAction:        userAAction,
		ReceiverAction:        userBAction,
	}

	// MatchLifecycleStatus is optional (nullable)
	if inserter.MatchLifecycleStatus != "" {
		matchSet.MatchLifecycleStatus = null.StringFrom(inserter.MatchLifecycleStatus)
	}

	if err := matchSet.Insert(ctx, exec, boil.Infer()); err != nil {
		return nil, fmt.Errorf("insert match set: %w", err)
	}

	return &matchSet, nil
}

// UpdateMatchResult updates fields of a match result.
func (s *Store) UpdateMatchResult(
	ctx context.Context,
	exec boil.ContextExecutor,
	updater *UpdateMatchResult,
) error {
	matchResult, err := pgmodel.FindMatchResult(ctx, exec, updater.ID)
	if err != nil {
		return fmt.Errorf("find match result by ID: %w", err)
	}

	if updater.MatchLifecycleStatus.Valid {
		matchResult.MatchLifecycleStatus = updater.MatchLifecycleStatus
	}
	if updater.QualifierResults.Valid {
		matchResult.QualifierResults = updater.QualifierResults
	}
	if updater.MatchedQualitatively.Valid {
		matchResult.MatchedQualitatively = updater.MatchedQualitatively.Bool
	}
	if updater.DeliveredToUserAt.Valid {
		matchResult.DeliveredToUserAt = updater.DeliveredToUserAt
	}
	if updater.IsPossibleMatch.Valid {
		matchResult.IsPossibleMatch = updater.IsPossibleMatch.Bool
	}
	if updater.IsApproved.Valid {
		matchResult.IsApproved = updater.IsApproved.Bool
	}
	if updater.IsDropped.Valid {
		matchResult.IsDropped = updater.IsDropped.Bool
	}
	if updater.DroppedTS.Valid {
		matchResult.DroppedTS = updater.DroppedTS
	}
	if updater.IsExpired.Valid {
		matchResult.IsExpired = updater.IsExpired.Bool
	}

	// Per-user action fields (string enums)
	if updater.InitiatorAction.Valid {
		matchResult.InitiatorAction = updater.InitiatorAction.String
	}
	if updater.InitiatorActionAt.Valid {
		matchResult.InitiatorActionAt = updater.InitiatorActionAt
	}
	if updater.InitiatorSeenAt.Valid {
		matchResult.InitiatorSeenAt = updater.InitiatorSeenAt
	}
	if updater.ReceiverAction.Valid {
		matchResult.ReceiverAction = updater.ReceiverAction.String
	}
	if updater.ReceiverActionAt.Valid {
		matchResult.ReceiverActionAt = updater.ReceiverActionAt
	}
	if updater.ReceiverSeenAt.Valid {
		matchResult.ReceiverSeenAt = updater.ReceiverSeenAt
	}
	if updater.ExpiresAt.Valid {
		matchResult.ExpiresAt = updater.ExpiresAt
	}

	if _, err := matchResult.Update(ctx, exec, boil.Infer()); err != nil {
		return fmt.Errorf("update match result: %w", err)
	}

	return nil
}

// UpdateMatchResult represents the fields that can be updated on a match result.
type UpdateMatchResult struct {
	ID                   string
	MatchLifecycleStatus null.String // String enum
	QualifierResults     null.JSON
	MatchedQualitatively null.Bool
	DeliveredToUserAt    null.Time
	IsVerified           null.Bool
	IsExpired            null.Bool
	IsApproved           null.Bool
	IsDropped            null.Bool
	DroppedTS            null.Time
	IsPossibleMatch      null.Bool

	// Per-user action fields (string enums)
	InitiatorAction   null.String // String enum
	InitiatorActionAt null.Time
	InitiatorSeenAt   null.Time
	ReceiverAction   null.String // String enum
	ReceiverActionAt null.Time
	ReceiverSeenAt   null.Time
	ExpiresAt     null.Time
}

// UpdateMatchForDateInstance links a match to its date instance.
// This is called when both users have proposed (mutual proposal).
type UpdateMatchForDateInstance struct {
	MatchResultID         string
	CurrentDateInstanceID string
	MatchLifecycleStatus  string // String enum
}

// UpdateMatchForDateInstance updates match_result with date instance reference.
func (s *Store) UpdateMatchForDateInstance(
	ctx context.Context,
	exec boil.ContextExecutor,
	updater *UpdateMatchForDateInstance,
) error {
	if updater.MatchResultID == "" {
		return fmt.Errorf("match_result_id is required")
	}
	if updater.CurrentDateInstanceID == "" {
		return fmt.Errorf("current_date_instance_id is required")
	}
	if updater.MatchLifecycleStatus == "" {
		return fmt.Errorf("match_lifecycle_status is required")
	}

	matchResult, err := pgmodel.FindMatchResult(ctx, exec, updater.MatchResultID)
	if err != nil {
		return fmt.Errorf("find match result: %w", err)
	}

	matchResult.CurrentDateInstanceID = null.StringFrom(updater.CurrentDateInstanceID)
	matchResult.MatchLifecycleStatus = null.StringFrom(updater.MatchLifecycleStatus)

	cols := []string{
		pgmodel.MatchResultColumns.CurrentDateInstanceID,
		pgmodel.MatchResultColumns.MatchLifecycleStatus,
	}

	if _, err := matchResult.Update(ctx, exec, boil.Whitelist(cols...)); err != nil {
		return fmt.Errorf("update match result: %w", err)
	}

	return nil
}
