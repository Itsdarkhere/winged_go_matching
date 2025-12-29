package store

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/boilhelper"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

type MatchSetStore struct {
	l    applog.Logger
	repo *repo.Store
}

func (s *MatchSetStore) matchSetsRows(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterMatchSet,
) ([]matching.MatchSet, error) {
	var matchSets []matching.MatchSet

	msCols := pgmodel.MatchSetColumns

	qMods := append(
		qModsMatchSet(f, true),
		qm.Select(
			"ms."+msCols.ID+" AS id",
			"ms."+msCols.Name+" AS name",
			"ms."+msCols.NumberOfParticipants+" AS number_of_participants",
			"ms."+msCols.MatchConfiguration+" AS match_configuration",
			"ms."+msCols.TimeStart+" AS time_start",
			"ms."+msCols.TimeEnd+" AS time_end",
			"ms."+msCols.CreatedAt+" AS created_at",
			"ms."+msCols.UpdatedAt+" AS updated_at",
		),
		qm.From(pgmodel.TableNames.MatchSet+" ms"),
	)

	if err := pgmodel.NewQuery(qMods...).Bind(ctx, exec, &matchSets); err != nil {
		return nil, fmt.Errorf("query match sets: %w", err)
	}
	return matchSets, nil
}

func (s *MatchSetStore) matchSetsCount(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterMatchSet,
) (int, error) {
	count, err := pgmodel.MatchSets(qModsMatchSet(f, false)...).Count(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("count match sets: %w", err)
	}
	return int(count), nil
}

func (s *MatchSetStore) MatchSets(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterMatchSet,
) (*matching.MatchSetPaginated, error) {
	rows, err := s.matchSetsRows(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("match sets rows: %w", err)
	}

	count := len(rows)
	if f.Pagination != nil {
		count, err = s.matchSetsCount(ctx, exec, f)
		if err != nil {
			return nil, fmt.Errorf("match sets count: %w", err)
		}
	}

	paginated := &matching.MatchSetPaginated{
		Data:       rows,
		Pagination: f.Pagination.Recalculated(count),
	}

	return paginated, nil
}

func (s *MatchSetStore) MatchSet(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterMatchSet,
) (*matching.MatchSet, error) {
	m, err := s.MatchSets(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("match sets: %w", err)
	}

	mLen := len(m.Data)
	if mLen != 1 {
		return nil, fmt.Errorf("match set count mismatch, have %d, want 1", mLen)
	}

	return &m.Data[0], nil
}

func qModsMatchSet(f *matching.QueryFilterMatchSet, paginated bool) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)

	msCols := pgmodel.MatchSetColumns

	// Use "ms." alias for paginated queries, bare column names for count queries
	colPrefix := ""
	if paginated {
		colPrefix = "ms."
	}

	if f.ID.Valid {
		qMods = append(qMods, qm.Where(colPrefix+msCols.ID+" = ?", f.ID.String))
	}

	if f.Name.Valid {
		qMods = append(qMods, qm.Where(colPrefix+msCols.Name+" ILIKE ?", "%"+f.Name.String+"%"))
	}

	if f.NumberOfParticipantsMin.Valid {
		qMods = append(qMods, qm.Where(colPrefix+msCols.NumberOfParticipants+" >= ?", f.NumberOfParticipantsMin.Int))
	}

	if f.NumberOfParticipantsMax.Valid {
		qMods = append(qMods, qm.Where(colPrefix+msCols.NumberOfParticipants+" <= ?", f.NumberOfParticipantsMax.Int))
	}

	if f.TimeStartAfter.Valid {
		qMods = append(qMods, qm.Where(colPrefix+msCols.TimeStart+" >= ?", f.TimeStartAfter.Time))
	}

	if f.TimeStartBefore.Valid {
		qMods = append(qMods, qm.Where(colPrefix+msCols.TimeStart+" <= ?", f.TimeStartBefore.Time))
	}

	if f.TimeEndAfter.Valid {
		qMods = append(qMods, qm.Where(colPrefix+msCols.TimeEnd+" >= ?", f.TimeEndAfter.Time))
	}

	if f.TimeEndBefore.Valid {
		qMods = append(qMods, qm.Where(colPrefix+msCols.TimeEnd+" <= ?", f.TimeEndBefore.Time))
	}

	if f.CreatedAfter.Valid {
		qMods = append(qMods, qm.Where(colPrefix+msCols.CreatedAt+" >= ?", f.CreatedAfter.Time))
	}

	if f.CreatedBefore.Valid {
		qMods = append(qMods, qm.Where(colPrefix+msCols.CreatedAt+" <= ?", f.CreatedBefore.Time))
	}

	// Only apply ordering and pagination for data queries, not count queries
	if paginated {
		// Apply ordering with column whitelisting
		if f.OrderBy.Valid && f.Sort.Valid {
			allowedColumns := map[string]string{
				"created_at":             msCols.CreatedAt,
				"updated_at":             msCols.UpdatedAt,
				"name":                   msCols.Name,
				"number_of_participants": msCols.NumberOfParticipants,
				"time_start":             msCols.TimeStart,
				"time_end":               msCols.TimeEnd,
			}

			if col, ok := allowedColumns[f.OrderBy.String]; ok {
				orderDir := "DESC"
				if f.Sort.String == "+" {
					orderDir = "ASC"
				}
				qMods = append(qMods, qm.OrderBy("ms."+col+" "+orderDir))
			}
		}

		if f.Pagination != nil {
			qMods = boilhelper.ApplyPagination(qMods, f.Pagination)
		}
	}

	return qMods
}

// Insert inserts a new match set into the database.
func (s *MatchSetStore) Insert(
	ctx context.Context,
	exec boil.ContextExecutor,
	inserter *matching.InsertMatchSet,
) (*matching.MatchSet, error) {
	pgMatchSet, err := s.repo.InsertMatchSet(ctx, exec, &repo.InsertMatchSet{
		Name:                  inserter.Name,
		NumberOfParticipants:  inserter.NumberOfParticipants,
		MatchingConfiguration: inserter.MatchingParameters,
	})
	if err != nil {
		return nil, fmt.Errorf("insert match set: %w", err)
	}

	matchSet, err := s.MatchSet(ctx, exec, &matching.QueryFilterMatchSet{
		ID: null.StringFrom(pgMatchSet.ID),
	})
	if err != nil {
		return nil, fmt.Errorf("query match set after insert: %w", err)
	}

	return matchSet, nil
}

func (s *MatchSetStore) Update(
	ctx context.Context,
	exec boil.ContextExecutor,
	updater *matching.UpdateMatchSet,
) error {
	return nil
}
