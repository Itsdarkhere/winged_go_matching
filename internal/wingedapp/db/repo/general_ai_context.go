package repo

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

type GeneralAIContextQueryFilter struct {
	ID     null.String
	UserID null.String
}

func qModGeneralAIContext(f *GeneralAIContextQueryFilter) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)

	if f.ID.Valid {
		qMods = append(qMods, pgmodel.GeneralAIContextWhere.ID.EQ(f.ID.String))
	}

	return qMods
}

func (s *Store) GeneralAIContexts(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *GeneralAIContextQueryFilter,
) (pgmodel.GeneralAIContextSlice, error) {
	gCtxs, err := pgmodel.GeneralAIContexts(qModGeneralAIContext(f)...).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("get GeneralAIContext slice: %w", err)
	}
	return gCtxs, nil
}

func (s *Store) GeneralAIContext(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *GeneralAIContextQueryFilter,
) (*pgmodel.GeneralAIContext, error) {
	gCtxs, err := pgmodel.GeneralAIContexts(qModGeneralAIContext(f)...).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("get GeneralAIContext slice: %w", err)
	}

	if len(gCtxs) != 1 {
		return nil, fmt.Errorf("generalAIContext count mismatch, have %d, want 1", len(gCtxs))
	}

	return gCtxs[0], nil
}
