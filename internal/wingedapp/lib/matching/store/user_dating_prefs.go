package store

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"
)

type UserDatingPrefsStore struct {
	l    applog.Logger
	repo *repo.Store
}

func (s *UserDatingPrefsStore) UserDatingPreferences(ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterUserDatingPrefs,
) ([]matching.UserDatingPreference, error) {
	var userDatingPrefs []matching.UserDatingPreference

	if err := pgmodel.UserDatingPreferences(qModsUserDatingPrefs(f)...).Bind(ctx, exec, &userDatingPrefs); err != nil {
		return nil, fmt.Errorf("user dating prefs: %w", err)
	}

	return userDatingPrefs, nil
}

func qModsUserDatingPrefs(f *matching.QueryFilterUserDatingPrefs) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)

	if f.UserID.Valid {
		qMods = append(qMods, pgmodel.UserDatingPreferenceWhere.UserID.EQ(f.UserID.String))
	}

	return qMods
}

// Insert inserts dating preferences for a user.
func (s *UserDatingPrefsStore) Insert(
	ctx context.Context,
	exec boil.ContextExecutor,
	userID string,
	preferences []string,
) error {
	for _, pref := range preferences {
		datingPref := &pgmodel.UserDatingPreference{
			ID:               uuid.NewString(),
			UserID:           userID,
			DatingPreference: pref,
		}

		if err := datingPref.Insert(ctx, exec, boil.Infer()); err != nil {
			return fmt.Errorf("insert dating preference '%s': %w", pref, err)
		}
	}

	s.l.Debug(ctx, "inserted dating preferences", applog.F("user_id", userID), applog.F("preferences", preferences))

	return nil
}
