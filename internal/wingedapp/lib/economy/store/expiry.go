package store

import (
	"context"
	"database/sql"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/economy"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// ExpiryStore handles transaction queries for expiry operations.
// Following CLAUDE.md: store exposes generic verbs (Transactions, Update),
// logic layer upstream composes the specific queries.
type ExpiryStore struct{}

// NewExpiryStore creates a new ExpiryStore.
func NewExpiryStore() *ExpiryStore {
	return &ExpiryStore{}
}

// Transactions returns transactions matching the filter.
// Supports grouping by user with SUM aggregation when GroupByUser is set.
func (s *ExpiryStore) Transactions(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *economy.QueryFilterExpiryTransaction,
) ([]economy.ExpiredUserAmount, error) {
	var results []economy.ExpiredUserAmount

	cols := pgmodel.WingsEcnTransactionColumns

	qMods := []qm.QueryMod{}

	// Select: either grouped or individual
	if f.GroupByUser.Valid && f.GroupByUser.Bool {
		qMods = append(qMods,
			qm.Select(
				cols.UserRefID+" AS user_id",
				"SUM("+cols.Amount+") AS amount",
			),
			qm.GroupBy(cols.UserRefID),
		)
	}

	// Filters
	if f.ExpiresAtBefore.Valid {
		qMods = append(qMods, qm.Where(cols.ExpiresAt+" < ?", f.ExpiresAtBefore.Time))
	}
	if f.ExpiresAtNotNull.Valid && f.ExpiresAtNotNull.Bool {
		qMods = append(qMods, qm.Where(cols.ExpiresAt+" IS NOT NULL"))
	}
	if f.IsExpired.Valid {
		qMods = append(qMods, qm.Where(cols.IsExpired+" = ?", f.IsExpired.Bool))
	}
	if f.IsActive.Valid {
		qMods = append(qMods, qm.Where(cols.IsActive+" = ?", f.IsActive.Int))
	}
	if f.IsCredit.Valid {
		qMods = append(qMods, qm.Where(cols.IsCredit+" = ?", f.IsCredit.Bool))
	}

	if err := pgmodel.WingsEcnTransactions(qMods...).Bind(ctx, exec, &results); err != nil {
		if err == sql.ErrNoRows {
			return []economy.ExpiredUserAmount{}, nil
		}
		return nil, fmt.Errorf("query transactions: %w", err)
	}

	return results, nil
}

// Update updates transactions matching the filter.
// Returns count of rows updated.
func (s *ExpiryStore) Update(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *economy.QueryFilterExpiryTransaction,
	u *economy.UpdateExpiryTransaction,
) (int64, error) {
	cols := pgmodel.WingsEcnTransactionColumns

	qMods := []qm.QueryMod{}

	// Filters
	if f.ExpiresAtBefore.Valid {
		qMods = append(qMods, qm.Where(cols.ExpiresAt+" < ?", f.ExpiresAtBefore.Time))
	}
	if f.ExpiresAtNotNull.Valid && f.ExpiresAtNotNull.Bool {
		qMods = append(qMods, qm.Where(cols.ExpiresAt+" IS NOT NULL"))
	}
	if f.IsExpired.Valid {
		qMods = append(qMods, qm.Where(cols.IsExpired+" = ?", f.IsExpired.Bool))
	}
	if f.IsActive.Valid {
		qMods = append(qMods, qm.Where(cols.IsActive+" = ?", f.IsActive.Int))
	}

	// Build update map
	updateMap := pgmodel.M{}
	if u.IsExpired.Valid {
		updateMap[cols.IsExpired] = u.IsExpired.Bool
	}

	if len(updateMap) == 0 {
		return 0, nil
	}

	rowsAffected, err := pgmodel.WingsEcnTransactions(qMods...).UpdateAll(ctx, exec, updateMap)
	if err != nil {
		return 0, fmt.Errorf("update transactions: %w", err)
	}

	return rowsAffected, nil
}
