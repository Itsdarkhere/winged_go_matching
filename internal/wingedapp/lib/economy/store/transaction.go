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

type TransactionStore struct {
	logger applog.Logger
	repo   *repo.Store
}

func (s *TransactionStore) Transaction(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *economy.QueryFilterTransactions,
) (*economy.Transaction, error) {
	t, err := s.Transactions(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("transactions: %w", err)
	}

	tLen := len(t)
	if tLen != 1 {
		return nil, fmt.Errorf("transaction count mismatch, have %d, want 1", tLen)
	}

	return &t[0], nil
}

func (s *TransactionStore) Transactions(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *economy.QueryFilterTransactions,
) ([]economy.Transaction, error) {
	var trans []economy.Transaction

	// tables
	tranTbl := pgmodel.TableNames.WingsEcnTransaction
	usrTbl := pgmodel.TableNames.Users

	// cols
	tranCols := pgmodel.WingsEcnTransactionColumns
	usrCols := pgmodel.UserColumns

	qMods := append(
		transactionFilters(f),
		qm.Select(
			"tx."+tranCols.ID+" AS id",
			"tx."+tranCols.Amount+" AS amount",
			"tx."+tranCols.IsCredit+" AS is_credit",
			"tx."+tranCols.IsActive+" AS is_active",
			"tx."+tranCols.UserRefID+" AS user_id",
		),
		qm.From(tranTbl+" tx"),
		qm.InnerJoin(usrTbl+" u ON u."+usrCols.ID+" = tx."+tranCols.UserRefID),
	)

	if err := pgmodel.NewQuery(qMods...).Bind(ctx, exec, &trans); err != nil {
		return nil, fmt.Errorf("transactions: %w", err)
	}

	return trans, nil
}

// transactionFilters builds query modifiers based on the provided filter.
func transactionFilters(f *economy.QueryFilterTransactions) []qm.QueryMod {
	transCols := pgmodel.WingsEcnTransactionColumns

	qMods := make([]qm.QueryMod, 0)
	if f.ID.Valid {
		qMods = append(qMods, qm.Where("tx."+transCols.ID+" = ?", f.ID.String))
	}
	if f.ActionLogID.Valid {
		qMods = append(qMods, qm.Where("tx."+transCols.ActionLogRefID+" = ?", f.ActionLogID.String))
	}
	if f.IsActive.Valid {
		qMods = append(qMods, qm.Where("tx."+transCols.IsActive+" = ?", f.IsActive.Int))
	}

	return qMods
}

func (s *TransactionStore) Insert(ctx context.Context,
	exec boil.ContextExecutor,
	inserter *economy.InsertTransaction,
) error {
	if err := s.repo.InsertWingsEcnTransaction(ctx, exec, &repo.InsertWingsEcnTransaction{
		UserRefID:      inserter.UserID,
		ActionLogType:  inserter.ActionTypeID,
		ActionLogRefID: inserter.ActionRefID,
		IsCredit:       inserter.IsCredit,
		Claimed:        inserter.Claimed,
		Amount:         inserter.WingsAmount,
		ExpiresAt:      inserter.ExpiresAt,
		ExtraInfo:      inserter.ExtraInfo,
	}); err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}

	return nil
}

func (a *TransactionStore) Delete(ctx context.Context,
	exec boil.ContextExecutor,
	uuid string,
) error {
	if err := a.repo.SoftDeleteWingsEcnTransaction(ctx, exec, uuid); err != nil {
		return fmt.Errorf("repo delete transaction: %w", err)
	}

	return nil
}
