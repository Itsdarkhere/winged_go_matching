package repo

import (
	"context"
	"fmt"
	"time"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

type SysParamQueryFilter struct {
	Key null.String
}

func sysParamQMods(f *SysParamQueryFilter) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)
	if f.Key.Valid {
		qMods = append(qMods, pgmodel.SysParamWhere.Key.EQ(f.Key.String))
	}
	return qMods
}

// SysParams retrieves system parameters from the database.
func (s *Store) SysParams(ctx context.Context,
	exec boil.ContextExecutor,
	f *SysParamQueryFilter,
) (pgmodel.SysParamSlice, error) {
	// 5 seconds - propagate context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sp, err := pgmodel.SysParams(sysParamQMods(f)...).All(ctxWithTimeout, exec)
	if err != nil {
		return nil, fmt.Errorf("get sysParam slice: %w", err)
	}
	return sp, nil
}

func (s *Store) SysParam(ctx context.Context,
	exec boil.ContextExecutor,
	f *SysParamQueryFilter,
) (*pgmodel.SysParam, error) {
	sysParams, err := s.SysParams(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	if len(sysParams) != 1 {
		return nil, fmt.Errorf("sysparam count mismatch, have %d, want 1", len(sysParams))
	}

	return sysParams[0], nil
}
