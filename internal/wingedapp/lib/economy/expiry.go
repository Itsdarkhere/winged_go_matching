package economy

import (
	"context"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// expiryTransactionStorer handles transaction queries for expiry operations.
// Following CLAUDE.md: lowercase interface, ends in -er, small (1-3 methods).
// Store exposes generic verbs, logic layer composes specific queries.
type expiryTransactionStorer interface {
	Transactions(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterExpiryTransaction) ([]ExpiredUserAmount, error)
	Update(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterExpiryTransaction, u *UpdateExpiryTransaction) (int64, error)
}

// ExpiryLogic handles wings expiration business logic.
type ExpiryLogic struct {
	transactionStore expiryTransactionStorer
	userTotalsStore  userTotalsStorer
}

// NewExpiryLogic creates a new ExpiryLogic.
// Following CLAUDE.md: guard in constructor, fail fast.
func NewExpiryLogic(
	transactionStore expiryTransactionStorer,
	userTotalsStore userTotalsStorer,
) (*ExpiryLogic, error) {
	if transactionStore == nil {
		return nil, fmt.Errorf("transactionStore is required")
	}
	if userTotalsStore == nil {
		return nil, fmt.Errorf("userTotalsStore is required")
	}
	return &ExpiryLogic{
		transactionStore: transactionStore,
		userTotalsStore:  userTotalsStore,
	}, nil
}

// ExpireWings processes all expired transactions atomically:
// 1. Get sum of expired amounts per user
// 2. Decrement each user's total_wings
// 3. Mark transactions as is_expired=true
// Returns count of expired transactions processed.
//
// IMPORTANT: Caller must wrap in transaction for atomicity.
func (e *ExpiryLogic) ExpireWings(
	ctx context.Context,
	exec boil.ContextExecutor,
) (int64, error) {
	now := time.Now()

	// 1. Get expired amounts grouped by user (compose query in logic layer)
	expiredByUser, err := e.transactionStore.Transactions(ctx, exec, &QueryFilterExpiryTransaction{
		ExpiresAtBefore:  null.TimeFrom(now),
		ExpiresAtNotNull: null.BoolFrom(true),
		IsExpired:        null.BoolFrom(false),
		IsActive:         null.IntFrom(1),
		IsCredit:         null.BoolFrom(true), // only credits can expire
		GroupByUser:      null.BoolFrom(true),
	})
	if err != nil {
		return 0, fmt.Errorf("fetch expired by user: %w", err)
	}

	if len(expiredByUser) == 0 {
		return 0, nil // nothing to expire
	}

	// 2. Decrement each user's wings
	for _, eu := range expiredByUser {
		totals, err := e.userTotalsStore.Totals(ctx, exec, eu.UserID)
		if err != nil {
			return 0, fmt.Errorf("fetch totals for user %s: %w", eu.UserID, err)
		}
		if totals == nil {
			continue // user has no totals record, skip
		}

		newWings := totals.Wings - eu.Amount
		if newWings < 0 {
			newWings = 0 // floor at 0
		}

		if err := e.userTotalsStore.Update(ctx, exec, &UpdateUserTotals{
			ID:    totals.ID,
			Wings: null.IntFrom(newWings),
		}); err != nil {
			return 0, fmt.Errorf("update totals for user %s: %w", eu.UserID, err)
		}
	}

	// 3. Mark all expired transactions (compose query in logic layer)
	count, err := e.transactionStore.Update(ctx, exec,
		&QueryFilterExpiryTransaction{
			ExpiresAtBefore:  null.TimeFrom(now),
			ExpiresAtNotNull: null.BoolFrom(true),
			IsExpired:        null.BoolFrom(false),
			IsActive:         null.IntFrom(1),
		},
		&UpdateExpiryTransaction{
			IsExpired: null.BoolFrom(true),
		},
	)
	if err != nil {
		return 0, fmt.Errorf("mark expired: %w", err)
	}

	return count, nil
}
