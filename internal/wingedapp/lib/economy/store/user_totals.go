package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/economy"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

type UserTotalsStore struct {
	logger applog.Logger
	repo   *repo.Store
}

// Totals retrieves the totals for a user by their UUID.
func (u *UserTotalsStore) Totals(ctx context.Context, exec boil.ContextExecutor, uuid string) (*economy.UserTotals, error) {
	var t economy.UserTotals

	col := pgmodel.WingsEcnUserTotalColumns
	where := pgmodel.WingsEcnUserTotalWhere
	sel := qm.Select

	if err := pgmodel.WingsEcnUserTotals(
		sel(
			col.ID+" AS id",
			col.TotalWings+" AS wings",
			col.CounterSentMessages+" AS sent_messages",
			col.PremiumExpiresIn+" AS premium_expires_in",
			col.StreakLastDate+" AS streak_last_date",
			col.StreakCurrentDays+" AS streak_current_days",
			col.StreakLongestDays+" AS streak_longest_days",
		),
		where.UserRefID.EQ(uuid),
	).Bind(ctx, exec, &t); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("pgmodel single user totals: %w", err)
	}

	return &t, nil
}

// Create creates a new user totals record for the given user UUID.
func (u *UserTotalsStore) Create(ctx context.Context, exec boil.ContextExecutor, uuid string) (*economy.UserTotals, error) {
	ut := pgmodel.WingsEcnUserTotal{UserRefID: uuid}
	if err := ut.Insert(ctx, exec, boil.Infer()); err != nil {
		return nil, fmt.Errorf("insert user totals: %w", err)
	}

	return &economy.UserTotals{
		ID: ut.ID,
	}, nil
}

// Update updates the user totals record for the given user UUID.
func (u *UserTotalsStore) Update(ctx context.Context, exec boil.ContextExecutor, updater *economy.UpdateUserTotals) error {
	if err := u.repo.UpdateWingsEcnUserTotals(ctx, exec, toRepoUpdaterUserTotals(updater)); err != nil {
		return fmt.Errorf("update user totals: %w", err)
	}

	return nil
}

func toRepoUpdaterUserTotals(u *economy.UpdateUserTotals) *repo.UpdateWingsEcnUserTotals {
	return &repo.UpdateWingsEcnUserTotals{
		ID:                u.ID,
		PremiumExpiresIn:  u.PremiumExpiresIn,
		TotalWings:        u.Wings,
		SentMessages:      u.SentMessages,
		StreakLastDate:    u.StreakLastDate,
		StreakCurrentDays: u.StreakCurrentDays,
		StreakLongestDays: u.StreakLongestDays,
	}
}
