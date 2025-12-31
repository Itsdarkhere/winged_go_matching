package store

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/boilhelper"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/enums"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"
)

type MatchResultStore struct {
	l    applog.Logger
	repo *repo.Store
}

func (s *MatchResultStore) matchResultsRows(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterMatchResult,
) ([]matching.MatchResult, error) {
	var results []matching.MatchResult

	// tables
	matchResultTbl := pgmodel.TableNames.MatchResult

	// cols
	matchResultCols := pgmodel.MatchResultColumns

	qMods := append(
		qModsMatchResult(f, true),
		qm.Select(
			"mr."+matchResultCols.ID+" AS id",
			"mr."+matchResultCols.MatchSetRefID+" AS match_set_id",
			"mr."+matchResultCols.InitiatorUserRefID+" AS initiator_user_id",
			"mr."+matchResultCols.ReceiverUserRefID+" AS receiver_user_id",
			"mr."+matchResultCols.IsExpired+" AS is_expired",
			"mr."+matchResultCols.IsPossibleMatch+" AS is_possible_match",
			"mr."+matchResultCols.IsApproved+" AS is_approved",
			"mr."+matchResultCols.QualifierResults+" AS qualifier_results",
			"mr."+matchResultCols.MatchedQualitatively+" AS matched_qualitatively",
			"mr."+matchResultCols.MatchLifecycleStatus+" AS user_lifecycle_status", // Now a direct string enum
		),
		qm.From(matchResultTbl+" mr"),
	)

	if err := pgmodel.NewQuery(qMods...).Bind(ctx, exec, &results); err != nil {
		return nil, fmt.Errorf("match results: %w", err)
	}

	return results, nil
}

func (s *MatchResultStore) matchResultsCount(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterMatchResult,
) (int, error) {
	// tables
	matchResultTbl := pgmodel.TableNames.MatchResult

	// Must use same FROM/alias structure as matchResultsRows since qModsMatchResult uses "mr." prefix
	qMods := append(
		qModsMatchResult(f, false),
		qm.Select("COUNT(*)"),
		qm.From(matchResultTbl+" mr"),
	)

	var count int
	if err := pgmodel.NewQuery(qMods...).QueryRow(exec).Scan(&count); err != nil {
		return 0, fmt.Errorf("match results: %w", err)
	}

	return count, nil
}

func (s *MatchResultStore) MatchResults(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterMatchResult,
) (*matching.MatchResultPaginated, error) {
	rows, err := s.matchResultsRows(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("match results rows: %w", err)
	}

	count := len(rows)
	if f.Pagination != nil {
		count, err = s.matchResultsCount(ctx, exec, f)
		if err != nil {
			return nil, fmt.Errorf("match results count: %w", err)
		}
	}

	paginated := &matching.MatchResultPaginated{
		Data:       rows,
		Pagination: f.Pagination.Recalculated(count),
	}

	return paginated, nil
}

func (s *MatchResultStore) MatchResult(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterMatchResult,
) (*matching.MatchResult, error) {
	m, err := s.MatchResults(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("match results: %w", err)
	}

	mLen := len(m.Data)
	if mLen != 1 {
		return nil, fmt.Errorf("match result count mismatch, have %d, want 1", mLen)
	}

	return &m.Data[0], nil
}

func (s *MatchResultStore) Insert(
	ctx context.Context,
	exec boil.ContextExecutor,
	inserter *matching.InsertMatchResult,
) (*matching.MatchResult, error) {
	// Use string enum values directly instead of category lookups
	pgRes, err := s.repo.InsertMatchResult(ctx, exec, &repo.InsertMatchResult{
		MatchSetRefID:        inserter.MatchSetID.String(),
		MatchLifecycleStatus: string(enums.MatchLifecycleStatusScheduling), // Default to Scheduling
		InitiatorUserRefID:   inserter.InitiatorUserID.String(),
		ReceiverUserRefID:    inserter.ReceiverUserID.String(),
		InitiatorAction:          string(enums.MatchUserActionPending), // Default to Pending
		ReceiverAction:          string(enums.MatchUserActionPending), // Default to Pending
	})
	if err != nil {
		return nil, fmt.Errorf("insert match result: %w", err)
	}

	id, err := uuid.Parse(pgRes.ID)
	if err != nil {
		return nil, fmt.Errorf("parse id: %w", err)
	}
	msID, err := uuid.Parse(pgRes.MatchSetRefID)
	if err != nil {
		return nil, fmt.Errorf("parse match_set_ref_id: %w", err)
	}
	initiatorID, err := uuid.Parse(pgRes.InitiatorUserRefID)
	if err != nil {
		return nil, fmt.Errorf("parse initiator_user_ref_id: %w", err)
	}
	receiverID, err := uuid.Parse(pgRes.ReceiverUserRefID)
	if err != nil {
		return nil, fmt.Errorf("parse receiver_user_ref_id: %w", err)
	}

	return &matching.MatchResult{
		ID:              id,
		MatchSetID:      msID,
		InitiatorUserID: initiatorID,
		ReceiverUserID:  receiverID,
	}, nil
}

func (s *MatchResultStore) Update(
	ctx context.Context,
	exec boil.ContextExecutor,
	updater *matching.UpdateMatchResult,
) (*matching.MatchResult, error) {
	if err := s.repo.UpdateMatchResult(ctx, exec, &repo.UpdateMatchResult{
		ID:                   updater.ID.String(),
		QualifierResults:     updater.QualifierResults,
		MatchedQualitatively: updater.MatchedQualitatively,
		IsVerified:           updater.IsVerified,
		IsExpired:            updater.IsExpired,
		IsApproved:           updater.IsApproved,
		IsDropped:            updater.IsDropped,
		DroppedTS:            updater.DroppedTS,
		IsPossibleMatch:      updater.IsPossibleMatch,
	}); err != nil {
		return nil, fmt.Errorf("update match result: %w", err)
	}

	matchResult, err := s.MatchResult(ctx, exec, &matching.QueryFilterMatchResult{
		ID: null.StringFrom(updater.ID.String()),
	})
	if err != nil {
		return nil, fmt.Errorf("fetch updated match result: %w", err)
	}

	return matchResult, nil
}

// UpdateMatchForDateInstance links a match to its date instance.
func (s *MatchResultStore) UpdateMatchForDateInstance(
	ctx context.Context,
	exec boil.ContextExecutor,
	updater *matching.UpdateMatchForDateInstance,
) error {
	return s.repo.UpdateMatchForDateInstance(ctx, exec, &repo.UpdateMatchForDateInstance{
		MatchResultID:         updater.MatchResultID,
		CurrentDateInstanceID: updater.CurrentDateInstanceID,
		MatchLifecycleStatus:  updater.MatchLifecycleStatus, // Now a string enum
	})
}

func qModsMatchResult(f *matching.QueryFilterMatchResult, paginated bool) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)

	mrCols := pgmodel.MatchResultColumns

	if f.ID.Valid {
		qMods = append(qMods, qm.Where("mr."+mrCols.ID+" = ?", f.ID.String))
	}

	if f.MatchSetID.Valid {
		qMods = append(qMods, qm.Where("mr."+mrCols.MatchSetRefID+" = ?", f.MatchSetID.String))
	}

	// UserID matches user as EITHER initiator OR receiver (OR condition)
	if f.UserID.Valid {
		qMods = append(qMods, qm.Where(
			"(mr."+mrCols.InitiatorUserRefID+" = ? OR mr."+mrCols.ReceiverUserRefID+" = ?)",
			f.UserID.String, f.UserID.String,
		))
	}

	// InitiatorUserID matches initiator specifically
	if f.InitiatorUserID.Valid {
		qMods = append(qMods, qm.Where("mr."+mrCols.InitiatorUserRefID+" = ?", f.InitiatorUserID.String))
	}

	// ReceiverUserID matches receiver specifically
	if f.ReceiverUserID.Valid {
		qMods = append(qMods, qm.Where("mr."+mrCols.ReceiverUserRefID+" = ?", f.ReceiverUserID.String))
	}

	// Status filters - now using string enum columns directly
	if f.MatchLifecycleStatus.Valid {
		qMods = append(qMods, qm.Where("mr."+mrCols.MatchLifecycleStatus+" = ?", f.MatchLifecycleStatus.String))
	}

	if f.InitiatorAction.Valid {
		qMods = append(qMods, qm.Where("mr."+mrCols.InitiatorAction+" = ?", f.InitiatorAction.String))
	}

	if f.ReceiverAction.Valid {
		qMods = append(qMods, qm.Where("mr."+mrCols.ReceiverAction+" = ?", f.ReceiverAction.String))
	}

	// Boolean filters
	if f.MatchedQualitatively.Valid {
		qMods = append(qMods, qm.Where("mr."+mrCols.MatchedQualitatively+" = ?", f.MatchedQualitatively.Bool))
	}

	if f.IsApproved.Valid {
		qMods = append(qMods, qm.Where("mr."+mrCols.IsApproved+" = ?", f.IsApproved.Bool))
	}

	if f.IsPossibleMatch.Valid {
		qMods = append(qMods, qm.Where("mr."+mrCols.IsPossibleMatch+" = ?", f.IsPossibleMatch.Bool))
	}

	if f.IsExpired.Valid {
		qMods = append(qMods, qm.Where("mr."+mrCols.IsExpired+" = ?", f.IsExpired.Bool))
	}

	if f.IsDropped.Valid {
		qMods = append(qMods, qm.Where("mr."+mrCols.IsDropped+" = ?", f.IsDropped.Bool))
	}

	// Only apply ordering and pagination for data queries, not count queries
	if paginated {
		// Apply ordering with column whitelisting
		if f.OrderBy.Valid && f.Sort.Valid {
			// Whitelist of allowed columns for ordering
			allowedColumns := map[string]string{
				"created_at":            mrCols.CreatedAt,
				"is_approved":           mrCols.IsApproved,
				"is_possible_match":     mrCols.IsPossibleMatch,
				"matched_qualitatively": mrCols.MatchedQualitatively,
				"is_expired":            mrCols.IsExpired,
				"is_dropped":            mrCols.IsDropped,
				"dropped_ts":            mrCols.DroppedTS,
				"expires_at":            mrCols.ExpiresAt,
			}

			if col, ok := allowedColumns[f.OrderBy.String]; ok {
				orderDir := "DESC"
				if f.Sort.String == "+" {
					orderDir = "ASC"
				}
				qMods = append(qMods, qm.OrderBy("mr."+col+" "+orderDir))
			}
		}

		if f.Pagination != nil {
			qMods = boilhelper.ApplyPagination(qMods, f.Pagination)
		}
	}

	return qMods
}
