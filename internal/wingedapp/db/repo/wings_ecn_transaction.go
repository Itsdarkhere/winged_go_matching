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

type InsertWingsEcnTransaction struct {
	UserRefID      string    `json:"user_ref_id"`
	ActionLogType  string    `json:"action_log_type"` // String enum
	ActionLogRefID string    `json:"action_log_ref_id"`
	IsCredit       bool      `json:"is_credit"`
	Claimed        bool      `json:"claimed"`
	Amount         int       `json:"amount"`
	ExpiresAt      null.Time `json:"expires_at"` // nullable: NULL = never expires
	ExtraInfo      null.JSON `json:"extra_info"`
}

// InsertWingsEcnTransaction inserts a new wings economy transaction.
func (s *Store) InsertWingsEcnTransaction(ctx context.Context, db boil.ContextExecutor, inserter *InsertWingsEcnTransaction) error {
	if inserter.UserRefID == "" {
		return fmt.Errorf("user_ref_id is required for inserting wings ecn transaction")
	}
	if inserter.ActionLogType == "" {
		return fmt.Errorf("action_log_type is required for inserting wings ecn transaction")
	}
	if inserter.ActionLogRefID == "" {
		return fmt.Errorf("action_log_ref_id is required for inserting wings ecn transaction")
	}

	transaction := pgmodel.WingsEcnTransaction{
		UserRefID:      inserter.UserRefID,
		ActionLogType:  inserter.ActionLogType, // String enum
		ActionLogRefID: inserter.ActionLogRefID,
		IsCredit:       inserter.IsCredit,
		Claimed:        inserter.Claimed,
		Amount:         inserter.Amount,
		ExpiresAt:      inserter.ExpiresAt, // nullable: NULL = never expires
		ExtraInfo:      inserter.ExtraInfo,
	}

	if err := transaction.Insert(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("insert wings ecn transaction: %w", err)
	}
	return nil
}

type QueryFilterWingsEcnTransaction struct {
	ID            null.String `json:"id"`
	ActionLogType null.String `json:"action_log_type"` // String enum
	ActionLogID   null.String `json:"action_log_id"`
	UserID        null.String `json:"user_id"`
	IsCredit      null.Bool   `json:"is_credit"`
	Claimed       null.Bool   `json:"claimed"`
}

func qModsWingsEcnTransaction(f *QueryFilterWingsEcnTransaction) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)
	if f.ID.Valid {
		qMods = append(qMods, pgmodel.WingsEcnTransactionWhere.ID.EQ(f.ID.String))
	}
	if f.UserID.Valid {
		qMods = append(qMods, pgmodel.WingsEcnTransactionWhere.UserRefID.EQ(f.UserID.String))
	}
	if f.ActionLogType.Valid {
		qMods = append(qMods, pgmodel.WingsEcnTransactionWhere.ActionLogType.EQ(f.ActionLogType.String))
	}
	if f.ActionLogID.Valid {
		qMods = append(qMods, pgmodel.WingsEcnTransactionWhere.ActionLogRefID.EQ(f.ActionLogID.String))
	}
	if f.IsCredit.Valid {
		qMods = append(qMods, pgmodel.WingsEcnTransactionWhere.IsCredit.EQ(f.IsCredit.Bool))
	}
	if f.Claimed.Valid {
		qMods = append(qMods, pgmodel.WingsEcnTransactionWhere.Claimed.EQ(f.Claimed.Bool))
	}

	return qMods
}

// WingsEcnTransactions lists all the transactions based on the filter
func (s *Store) WingsEcnTransactions(ctx context.Context,
	exec boil.ContextExecutor,
	f *QueryFilterWingsEcnTransaction,
) (pgmodel.WingsEcnTransactionSlice, error) {
	trans, err := pgmodel.WingsEcnTransactions(qModsWingsEcnTransaction(f)...).All(ctx, exec)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgmodel.WingsEcnTransactionSlice{}, nil
		}
		return nil, fmt.Errorf("query wings ecn trans: %w", err)
	}

	return trans, nil
}

// WingsEcnTransaction returns one transaction based on the filter
func (s *Store) WingsEcnTransaction(ctx context.Context,
	exec boil.ContextExecutor,
	f *QueryFilterWingsEcnTransaction,
) (*pgmodel.WingsEcnTransaction, error) {
	trans, err := s.WingsEcnTransactions(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("wings ecn transaction: %w", err)
	}

	if len(trans) == 0 {
		return nil, fmt.Errorf("wings ecn transaction: none found")
	}

	if len(trans) != 1 {
		return nil, fmt.Errorf("wings ecn transaction count mismatch, have %d, want 1", len(trans))
	}

	return trans[0], nil
}

// SoftDeleteWingsEcnTransaction soft deletes a wings economy transaction.
func (s *Store) SoftDeleteWingsEcnTransaction(
	ctx context.Context,
	exec boil.ContextExecutor,
	uuid string,
) error {
	transaction, err := pgmodel.FindWingsEcnTransaction(ctx, exec, uuid)
	if err != nil {
		return fmt.Errorf("find wings ecn transaction: %w", err)
	}

	// soft delete
	transaction.IsActive = null.IntFrom(0)
	whiteList := boil.Whitelist(pgmodel.WingsEcnTransactionColumns.IsActive)

	if _, err = transaction.Update(ctx, exec, whiteList); err != nil {
		return fmt.Errorf("soft delete wings ecn transaction: %w", err)
	}

	return nil
}

// UpdateWingsEcnTransaction is the update struct for wings economy transactions.
type UpdateWingsEcnTransaction struct {
	IDs     []string  // transaction IDs to update
	Claimed null.Bool // set claimed status
}

// UpdateWingsEcnTransactions updates multiple wings economy transactions.
func (s *Store) UpdateWingsEcnTransactions(
	ctx context.Context,
	exec boil.ContextExecutor,
	updater *UpdateWingsEcnTransaction,
) (int64, error) {
	if len(updater.IDs) == 0 {
		return 0, nil
	}

	cols := pgmodel.WingsEcnTransactionColumns
	var whitelist []string

	// Build update model
	txn := &pgmodel.WingsEcnTransaction{}

	if updater.Claimed.Valid {
		txn.Claimed = updater.Claimed.Bool
		whitelist = append(whitelist, cols.Claimed)
	}

	if len(whitelist) == 0 {
		return 0, nil
	}

	rowsAffected, err := pgmodel.WingsEcnTransactions(
		pgmodel.WingsEcnTransactionWhere.ID.IN(updater.IDs),
	).UpdateAll(ctx, exec, pgmodel.M{
		cols.Claimed: txn.Claimed,
	})
	if err != nil {
		return 0, fmt.Errorf("update wings ecn transactions: %w", err)
	}

	return rowsAffected, nil
}
