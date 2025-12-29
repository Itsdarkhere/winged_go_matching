package store

import (
	"context"
	"database/sql"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/economy"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// DailyCheckinStore handles daily check-in transaction data access.
type DailyCheckinStore struct {
	logger applog.Logger
	repo   *repo.Store
}

// NewDailyCheckinStore creates a new DailyCheckinStore.
func NewDailyCheckinStore(l applog.Logger) *DailyCheckinStore {
	return &DailyCheckinStore{logger: l, repo: &repo.Store{}}
}

// Transactions returns check-in transactions matching filter.
func (s *DailyCheckinStore) Transactions(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *economy.QueryFilterCheckinTransaction,
) ([]economy.CheckinTransaction, error) {
	var results []economy.CheckinTransaction

	cols := pgmodel.WingsEcnTransactionColumns

	qMods := []qm.QueryMod{
		qm.Select(
			cols.ID+" AS id",
			cols.UserRefID+" AS user_id",
			cols.Amount+" AS amount",
			cols.Claimed+" AS claimed",
			cols.CreatedDate+" AS created_date",
		),
	}

	// Apply conditional filters
	if f.ActionLogType.Valid {
		qMods = append(qMods, qm.Where(cols.ActionLogType+" = ?", f.ActionLogType.String))
	}
	if f.IsActive.Valid {
		qMods = append(qMods, qm.Where(cols.IsActive+" = ?", f.IsActive.Int))
	}
	if f.UserID.Valid {
		qMods = append(qMods, qm.Where(cols.UserRefID+" = ?", f.UserID.String))
	}
	if f.Claimed.Valid {
		qMods = append(qMods, qm.Where(cols.Claimed+" = ?", f.Claimed.Bool))
	}

	// Ordering
	orderBy := cols.CreatedDate
	if f.OrderBy.Valid {
		orderBy = f.OrderBy.String
	}
	sort := "ASC"
	if f.Sort.Valid && f.Sort.String == "-" {
		sort = "DESC"
	}
	qMods = append(qMods, qm.OrderBy(orderBy+" "+sort))

	// Limit
	if f.Limit.Valid {
		qMods = append(qMods, qm.Limit(f.Limit.Int))
	}

	if err := pgmodel.WingsEcnTransactions(qMods...).Bind(ctx, exec, &results); err != nil {
		if err == sql.ErrNoRows {
			return []economy.CheckinTransaction{}, nil
		}
		return nil, fmt.Errorf("query checkin transactions: %w", err)
	}
	return results, nil
}

// Transaction returns a single check-in transaction matching filter.
func (s *DailyCheckinStore) Transaction(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *economy.QueryFilterCheckinTransaction,
) (*economy.CheckinTransaction, error) {
	txns, err := s.Transactions(ctx, exec, f)
	if err != nil {
		return nil, err
	}
	if len(txns) != 1 {
		return nil, fmt.Errorf("transaction count mismatch, have %d, want 1", len(txns))
	}
	return &txns[0], nil
}

// TransactionCount returns count of transactions matching filter.
func (s *DailyCheckinStore) TransactionCount(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *economy.QueryFilterCheckinTransaction,
) (int, error) {
	cols := pgmodel.WingsEcnTransactionColumns

	var qMods []qm.QueryMod

	// Apply conditional filters
	if f.ActionLogType.Valid {
		qMods = append(qMods, qm.Where(cols.ActionLogType+" = ?", f.ActionLogType.String))
	}
	if f.IsActive.Valid {
		qMods = append(qMods, qm.Where(cols.IsActive+" = ?", f.IsActive.Int))
	}
	if f.UserID.Valid {
		qMods = append(qMods, qm.Where(cols.UserRefID+" = ?", f.UserID.String))
	}
	if f.Claimed.Valid {
		qMods = append(qMods, qm.Where(cols.Claimed+" = ?", f.Claimed.Bool))
	}

	count, err := pgmodel.WingsEcnTransactions(qMods...).Count(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("count checkin transactions: %w", err)
	}
	return int(count), nil
}

// Insert creates a new check-in transaction via repo.
func (s *DailyCheckinStore) Insert(
	ctx context.Context,
	exec boil.ContextExecutor,
	inserter *economy.InsertCheckinTransaction,
) (*pgmodel.WingsEcnTransaction, error) {
	if err := s.repo.InsertWingsEcnTransaction(ctx, exec, &repo.InsertWingsEcnTransaction{
		UserRefID:      inserter.UserID,
		ActionLogType:  string(economy.ActionDailyCheckIn),
		ActionLogRefID: inserter.ActionLogID,
		IsCredit:       true,
		Claimed:        false, // unclaimed by default
		Amount:         inserter.Amount,
		ExpiresAt:      inserter.ExpiresAt, // 30 days for earned wings
	}); err != nil {
		return nil, fmt.Errorf("insert checkin transaction: %w", err)
	}

	// Fetch the inserted transaction
	cols := pgmodel.WingsEcnTransactionColumns
	tx, err := pgmodel.WingsEcnTransactions(
		qm.Where(cols.ActionLogRefID+" = ?", inserter.ActionLogID),
		qm.Where(cols.ActionLogType+" = ?", string(economy.ActionDailyCheckIn)),
	).One(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("fetch inserted transaction: %w", err)
	}

	return tx, nil
}

// Update updates check-in transactions via repo.
func (s *DailyCheckinStore) Update(
	ctx context.Context,
	exec boil.ContextExecutor,
	updater *economy.UpdateCheckinTransaction,
) (int, error) {
	rowsAffected, err := s.repo.UpdateWingsEcnTransactions(ctx, exec, &repo.UpdateWingsEcnTransaction{
		IDs:     updater.IDs,
		Claimed: updater.Claimed,
	})
	if err != nil {
		return 0, fmt.Errorf("update checkin transactions: %w", err)
	}
	return int(rowsAffected), nil
}
