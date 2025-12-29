package repo

import (
	"context"
	"errors"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jackc/pgx/v4"
)

type InsertWingsEcnActionLog struct {
	UserRefID      string    `json:"user_ref_id"`
	ActionLogType  string    `json:"action_log_type"` // String enum
	ExtDomainRefID string    `json:"ext_domain_ref_id"`
	IsCredit       bool      `json:"is_credit"`
	ExtraInfo      null.JSON `json:"extra_info"`
}

// InsertWingsEcnActionLog inserts a new wings economy action log.
func (s *Store) InsertWingsEcnActionLog(ctx context.Context, db boil.ContextExecutor, inserter *InsertWingsEcnActionLog) (*pgmodel.WingsEcnActionLog, error) {
	if inserter.UserRefID == "" {
		return nil, fmt.Errorf("user_ref_id is required for inserting wings ecn action log")
	}
	if inserter.ExtDomainRefID == "" {
		return nil, fmt.Errorf("ext_domain_ref_id is required for inserting wings ecn action log")
	}
	if inserter.ActionLogType == "" {
		return nil, fmt.Errorf("action_log_type is required for inserting wings ecn action log")
	}

	actionLog := pgmodel.WingsEcnActionLog{
		UserRefID:      inserter.UserRefID,
		ExtDomainRefID: inserter.ExtDomainRefID,
		ActionLogType:  inserter.ActionLogType, // String enum
		IsCredit:       inserter.IsCredit,
		ExtraInfo:      inserter.ExtraInfo,
	}

	if err := actionLog.Insert(ctx, db, boil.Infer()); err != nil {
		return nil, fmt.Errorf("insert wings ecn action log: %w", err)
	}

	return &actionLog, nil
}

type QueryFilterWingsEcnActionLog struct {
	ID             null.String `json:"id"`
	UserRefID      null.String `json:"user_ref_id"`
	ExtDomainRefID null.String `json:"ext_domain_ref_id"`
	IsCredit       null.Bool   `json:"is_credit"`
}

func qModsWingsEcnActionLog(f *QueryFilterWingsEcnActionLog) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)

	if f.ID.Valid {
		qMods = append(qMods, pgmodel.WingsEcnActionLogWhere.ID.EQ(f.ID.String))
	}
	if f.UserRefID.Valid {
		qMods = append(qMods, pgmodel.WingsEcnActionLogWhere.UserRefID.EQ(f.UserRefID.String))
	}
	if f.ExtDomainRefID.Valid {
		qMods = append(qMods, pgmodel.WingsEcnActionLogWhere.ExtDomainRefID.EQ(f.ExtDomainRefID.String))
	}
	if f.IsCredit.Valid {
		qMods = append(qMods, pgmodel.WingsEcnActionLogWhere.IsCredit.EQ(f.IsCredit.Bool))
	}

	return qMods
}

// WingsEcnActionLogs lists all the action logs based on the filter
func (s *Store) WingsEcnActionLogs(ctx context.Context,
	exec boil.ContextExecutor,
	f *QueryFilterWingsEcnActionLog,
) (pgmodel.WingsEcnActionLogSlice, error) {
	actionLogs, err := pgmodel.WingsEcnActionLogs(qModsWingsEcnActionLog(f)...).All(ctx, exec)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgmodel.WingsEcnActionLogSlice{}, nil
		}
		return nil, fmt.Errorf("query wings ecn action logs: %w", err)
	}
	return actionLogs, nil
}

// WingsEcnActionLog returns one action log based on the filter
func (s *Store) WingsEcnActionLog(ctx context.Context,
	exec boil.ContextExecutor,
	filter *QueryFilterWingsEcnActionLog,
) (*pgmodel.WingsEcnActionLog, error) {
	actionLogs, err := s.WingsEcnActionLogs(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("wings ecn action log: %w", err)
	}

	if len(actionLogs) == 0 {
		return nil, fmt.Errorf("wings ecn action log: none found")
	}

	if len(actionLogs) != 1 {
		return nil, fmt.Errorf("wings ecn action log count mismatch, have %d, want 1", len(actionLogs))
	}

	return actionLogs[0], nil
}

// SoftDeleteWingsEcnActionLog soft deletes a wings economy action log.
func (s *Store) SoftDeleteWingsEcnActionLog(
	ctx context.Context,
	exec boil.ContextExecutor,
	uuid string,
) error {
	actionLog, err := pgmodel.FindWingsEcnActionLog(ctx, exec, uuid)
	if err != nil {
		return fmt.Errorf("find wings ecn action log: %w", err)
	}

	// soft delete
	actionLog.IsActive = null.IntFrom(0)
	whiteList := boil.Whitelist(pgmodel.WingsEcnActionLogColumns.IsActive)

	if _, err = actionLog.Update(ctx, exec, whiteList); err != nil {
		return fmt.Errorf("soft delete wings ecn action log: %w", err)
	}

	return nil
}
