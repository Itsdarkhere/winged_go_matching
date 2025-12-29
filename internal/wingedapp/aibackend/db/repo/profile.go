package repo

import (
	"context"
	"errors"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jackc/pgx/v4"
)

type ProfileQueryFilter struct {
	ID     string      `json:"id"`
	UserID null.String `json:"user_id"`
}

func profilesFilter(filter *ProfileQueryFilter) []qm.QueryMod {
	filters := make([]qm.QueryMod, 0)

	if filter.ID != "" {
		filters = append(filters, aipgmodel.ProfileWhere.ID.EQ(filter.ID))
	}
	if filter.UserID.Valid {
		filters = append(filters, aipgmodel.ProfileWhere.UserID.EQ(filter.UserID))
	}

	return filters
}

func (s *Store) Profiles(
	ctx context.Context,
	exec boil.ContextExecutor,
	filter *ProfileQueryFilter,
) (aipgmodel.ProfileSlice, error) {
	profiles, err := aipgmodel.Profiles(profilesFilter(filter)...).All(ctx, exec)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return aipgmodel.ProfileSlice{}, nil
		}
		return nil, fmt.Errorf("query profiles: %w", err)
	}
	return profiles, nil
}

// Profile returns a single profile matching the filter.
// Returns nil, nil if no profile found.
func (s *Store) Profile(
	ctx context.Context,
	exec boil.ContextExecutor,
	filter *ProfileQueryFilter,
) (*aipgmodel.Profile, error) {
	profiles, err := s.Profiles(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}

	if len(profiles) == 0 {
		return nil, nil
	}

	if len(profiles) != 1 {
		return nil, fmt.Errorf("profile count mismatch, have %d, want 1", len(profiles))
	}

	return profiles[0], nil
}

// InsertProfile inserts a profile into ai_backend.profiles.
func (s *Store) InsertProfile(
	ctx context.Context,
	exec boil.ContextExecutor,
	profile *aipgmodel.Profile,
) error {
	if err := profile.Insert(ctx, exec, boil.Infer()); err != nil {
		return fmt.Errorf("insert profile: %w", err)
	}
	return nil
}
