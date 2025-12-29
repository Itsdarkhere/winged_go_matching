package store

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/economy"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

type ActionLogStore struct {
	logger applog.Logger
	repo   *repo.Store
}

// Insert inserts a new action log entry.
func (a *ActionLogStore) Insert(ctx context.Context,
	exec boil.ContextExecutor,
	actionLogType string,
	inserter *economy.InsertActionLog,
) (*economy.ActionLog, error) {
	pgActionLog, err := a.repo.InsertWingsEcnActionLog(ctx, exec, &repo.InsertWingsEcnActionLog{
		UserRefID:      inserter.UserID,
		ExtDomainRefID: inserter.RefID,
		ExtraInfo:      inserter.JSONDetails,
		ActionLogType:  actionLogType,
	})
	if err != nil {
		return nil, fmt.Errorf("insert action log: %w", err)
	}

	return &economy.ActionLog{
		ID:          pgActionLog.ID,
		RefID:       pgActionLog.ExtDomainRefID,
		Type:        inserter.Type,
		JSONDetails: pgActionLog.ExtraInfo,
	}, nil
}

// actionLogFilters builds query modifiers based on the provided filter.
func actionLogFilters(f *economy.QueryFilterActionLog) []qm.QueryMod {
	actLogCols := pgmodel.WingsEcnActionLogColumns

	qMods := make([]qm.QueryMod, 0)
	if f.ID.Valid {
		qMods = append(qMods, qm.Where("al."+actLogCols.ID+" = ?", f.ID.String))
	}
	if f.Category.Valid {
		qMods = append(qMods, qm.Where("al."+actLogCols.ActionLogType+" = ?", f.Category.String))
	}
	if f.UserID.Valid {
		qMods = append(qMods, qm.Where("al."+actLogCols.UserRefID+" = ?", f.UserID.String))
	}
	if f.RefID.Valid {
		qMods = append(qMods, qm.Where("al."+actLogCols.ExtDomainRefID+" = ?", f.RefID.String))
	}
	if f.IsActive.Valid {
		qMods = append(qMods, qm.Where("al."+actLogCols.IsActive+" = ?", f.IsActive.Int))
	}

	return qMods
}

// ActionLogs retrieves action log entries based on the provided filter.
func (a *ActionLogStore) ActionLogs(ctx context.Context,
	exec boil.ContextExecutor,
	f *economy.QueryFilterActionLog,
) ([]economy.ActionLog, error) {
	var aLogs []economy.ActionLog

	actLogTbl := pgmodel.TableNames.WingsEcnActionLog

	actLogCols := pgmodel.WingsEcnActionLogColumns

	qMods := append(
		actionLogFilters(f),
		qm.Select(
			"al."+actLogCols.ID+" AS id",
			"al."+actLogCols.UserRefID+" AS user_id",
			"al."+actLogCols.ExtDomainRefID+" AS ref_id",
			"al."+actLogCols.ActionLogType+" AS type",
			"al."+actLogCols.ExtraInfo+" AS json_details",
			"al."+actLogCols.IsActive+" AS is_active",
		),
		qm.From(actLogTbl+" al"),
	)

	if err := pgmodel.NewQuery(qMods...).Bind(ctx, exec, &aLogs); err != nil {
		return nil, fmt.Errorf("action logs: %w", err)
	}

	return aLogs, nil
}

func (a *ActionLogStore) ActionLog(ctx context.Context,
	exec boil.ContextExecutor,
	f *economy.QueryFilterActionLog,
) (*economy.ActionLog, error) {
	aLogs, err := a.ActionLogs(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("action logs: %w", err)
	}

	actLogLen := len(aLogs)
	if actLogLen != 1 {
		return nil, fmt.Errorf("action logs count mismatch, have %d, want 1", actLogLen)
	}

	return &aLogs[0], nil
}

func (a *ActionLogStore) Delete(ctx context.Context,
	exec boil.ContextExecutor,
	uuid string,
) error {
	if err := a.repo.SoftDeleteWingsEcnActionLog(ctx, exec, uuid); err != nil {
		return fmt.Errorf("repo delete action log: %w", err)
	}

	return nil
}
