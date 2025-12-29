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

type UserDatingPreferencesQueryFilter struct {
	ID                null.String `json:"id"`
	UserID            null.String `json:"user_id"`
	DatingPreferences []string    `json:"dating_preferences"`
}

func userDatingPreferencesFilter(f *UserDatingPreferencesQueryFilter) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)
	if f.UserID.Valid {
		qMods = append(qMods, pgmodel.UserDatingPreferenceWhere.UserID.EQ(f.UserID.String))
	}

	if len(f.DatingPreferences) > 0 {
		qMods = append(qMods, pgmodel.UserDatingPreferenceWhere.DatingPreference.IN(f.DatingPreferences))
	}

	return qMods
}

func (s *Store) UserDatingPreferences(ctx context.Context,
	exec boil.ContextExecutor,
	filter *UserDatingPreferencesQueryFilter,
) (pgmodel.UserDatingPreferenceSlice, error) {
	userDatingPrefs, err := pgmodel.UserDatingPreferences(userDatingPreferencesFilter(filter)...).All(ctx, exec)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgmodel.UserDatingPreferenceSlice{}, nil
		}
		return nil, fmt.Errorf("query user dating preferences: %w", err)
	}

	return userDatingPrefs, nil
}

func (s *Store) UpsertUserDatingPreference(
	ctx context.Context,
	db boil.ContextExecutor,
	userID,
	datingPreference string,
) error {
	userDatingPref := pgmodel.UserDatingPreference{
		UserID:           userID,
		DatingPreference: datingPreference,
	}

	conflictCols := []string{
		pgmodel.UserDatingPreferenceColumns.UserID,
		pgmodel.UserDatingPreferenceColumns.DatingPreference,
	}

	if err := userDatingPref.Upsert(ctx, db, true, conflictCols, boil.Infer(), boil.Infer()); err != nil {
		return fmt.Errorf("upsert user dating preference: %w", err)
	}

	return nil
}
