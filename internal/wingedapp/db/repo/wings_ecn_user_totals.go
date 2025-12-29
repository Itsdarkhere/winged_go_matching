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

type InsertWingsEcnUserTotal struct {
	UserRefID           string `json:"user_ref_id"`
	TotalWings          int    `json:"total_wings"`
	CounterSentMessages int    `json:"counter_sent_messages"`
	CounterDailyCheckIn int    `json:"counter_daily_check_in"`
}

// InsertWingsEcnUserTotal inserts a new wings economy user total.
func (s *Store) InsertWingsEcnUserTotal(ctx context.Context, db boil.ContextExecutor, inserter *InsertWingsEcnUserTotal) error {
	if inserter.UserRefID == "" {
		return fmt.Errorf("user_ref_id is required for inserting wings ecn user total")
	}

	userTotal := pgmodel.WingsEcnUserTotal{
		UserRefID:           inserter.UserRefID,
		TotalWings:          inserter.TotalWings,
		CounterSentMessages: inserter.CounterSentMessages,
		CounterDailyCheckIn: inserter.CounterDailyCheckIn,
	}

	if err := userTotal.Insert(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("insert wings ecn user total: %w", err)
	}
	return nil
}

type QueryFilterWingsEcnUserTotal struct {
	ID     null.String `json:"id"`
	UserID null.String `json:"user_id"`
}

func qModsWingsEcnUserTotal(f *QueryFilterWingsEcnUserTotal) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)
	if f.ID.Valid {
		qMods = append(qMods, pgmodel.WingsEcnUserTotalWhere.ID.EQ(f.ID.String))
	}
	if f.UserID.Valid {
		qMods = append(qMods, pgmodel.WingsEcnUserTotalWhere.UserRefID.EQ(f.UserID.String))
	}

	return qMods
}

// WingsEcnUserTotals lists all the user totals based on the filter
func (s *Store) WingsEcnUserTotals(ctx context.Context,
	exec boil.ContextExecutor,
	f *QueryFilterWingsEcnUserTotal,
) (pgmodel.WingsEcnUserTotalSlice, error) {
	userTotals, err := pgmodel.WingsEcnUserTotals(qModsWingsEcnUserTotal(f)...).All(ctx, exec)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgmodel.WingsEcnUserTotalSlice{}, nil
		}
		return nil, fmt.Errorf("query wings ecn user totals: %w", err)
	}
	return userTotals, nil
}

// WingsEcnUserTotal returns one user total based on the filter
func (s *Store) WingsEcnUserTotal(ctx context.Context,
	exec boil.ContextExecutor,
	filter *QueryFilterWingsEcnUserTotal,
) (*pgmodel.WingsEcnUserTotal, error) {
	userTotals, err := s.WingsEcnUserTotals(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("wings ecn user total: %w", err)
	}

	if len(userTotals) == 0 {
		return nil, fmt.Errorf("wings ecn user total: none found")
	}

	if len(userTotals) != 1 {
		return nil, fmt.Errorf("wings ecn user total count mismatch, have %d, want 1", len(userTotals))
	}

	return userTotals[0], nil
}

// DeleteWingsEcnUserTotal deletes a wings economy user total.
func (s *Store) DeleteWingsEcnUserTotal(
	ctx context.Context,
	db boil.ContextExecutor,
	id string,
) error {
	userTotal := pgmodel.WingsEcnUserTotal{ID: id}
	count, err := userTotal.Delete(ctx, db)
	if err != nil {
		return fmt.Errorf("delete wings ecn user total: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("delete wings ecn user total: no rows affected")
	}

	return nil
}

// UpdateWingsEcnUserTotals updates user totals details in the database.
func (s *Store) UpdateWingsEcnUserTotals(ctx context.Context, exec boil.ContextExecutor, u *UpdateWingsEcnUserTotals) error {
	userTotals, err := pgmodel.FindWingsEcnUserTotal(ctx, exec, u.ID)
	if err != nil {
		return fmt.Errorf("find userTotals by UserID: %w", err)
	}

	if u.TotalWings.Valid {
		userTotals.TotalWings = u.TotalWings.Int
	}

	if u.PremiumExpiresIn.Valid {
		userTotals.PremiumExpiresIn = u.PremiumExpiresIn
	}

	if u.SentMessages.Valid {
		userTotals.CounterSentMessages = u.SentMessages.Int
	}

	if u.StreakLastDate.Valid {
		userTotals.StreakLastDate = u.StreakLastDate
	}

	if u.StreakCurrentDays.Valid {
		userTotals.StreakCurrentDays = u.StreakCurrentDays.Int
	}

	if u.StreakLongestDays.Valid {
		userTotals.StreakLongestDays = u.StreakLongestDays.Int
	}

	if _, err = userTotals.Update(ctx, exec, boil.Infer()); err != nil {
		return fmt.Errorf("update userTotals: %w", err)
	}

	return nil
}

// UpdateWingsEcnUserTotals represents the data needed to update a user's totals.
type UpdateWingsEcnUserTotals struct {
	ID                string
	TotalWings        null.Int  `json:"total_wings"`
	PremiumExpiresIn  null.Time `json:"premium_expires_in"`
	SentMessages      null.Int  `json:"sent_messages"`
	StreakLastDate    null.Time `json:"streak_last_date"`
	StreakCurrentDays null.Int  `json:"streak_current_days"`
	StreakLongestDays null.Int  `json:"streak_longest_days"`
}
