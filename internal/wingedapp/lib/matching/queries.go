package matching

import (
	"context"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
)

// userMatchingDetails fetches user details needed for matching.
func (l *Logic) userMatchingDetails(ctx context.Context, exec boil.ContextExecutor, uuid uuid.UUID) (*User, error) {
	user, err := l.user(ctx, exec, &QueryFilterUser{
		ID:                null.StringFrom(uuid.String()),
		EnrichDatingPrefs: true, // needed for matching
	})
	if err != nil {
		return nil, fmt.Errorf("fetch user by id %s: %w", uuid.String(), err)
	}

	return user, nil
}

func (l *Logic) user(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterUser) (*User, error) {
	users, err := l.users(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}
	return &users[0], nil
}

func (l *Logic) enrichUserDatingPrefs(ctx context.Context,
	exec boil.ContextExecutor,
	user *User,
) error {
	userDatingPrefs, err := l.userDatingPreferenceStorer.UserDatingPreferences(ctx, exec, &QueryFilterUserDatingPrefs{
		UserID: null.StringFrom(user.ID.String()),
	})
	if err != nil {
		return fmt.Errorf("user dating prefs storer: %w", err)
	}

	user.DatingPreferences = userDatingPrefs

	return nil
}

// users fetches all the users.
func (l *Logic) users(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterUser) ([]User, error) {
	users, err := l.userStorer.Users(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("user storer: %w", err)
	}

	for i := range users {
		if f.EnrichDatingPrefs {
			if err = l.enrichUserDatingPrefs(ctx, exec, &users[i]); err != nil {
				return nil, fmt.Errorf("enrich dating prefs storer: %w", err)
			}
		}
	}

	return users, nil
}

// MatchResults retrieves match results based on the provided context.
// Main use-case is to be a provider for API fetch handlers.
// aiExec is optional - if nil, profile enrichment is skipped.
func (l *Logic) MatchResults(ctx context.Context,
	exec boil.ContextExecutor,
	aiExec boil.ContextExecutor,
	f *QueryFilterMatchResult,
) (*MatchResultPaginated, error) {
	var err error

	matchResults, err := l.matchResultStorer.MatchResults(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("store match results: %w", err)
	}

	if f.EnrichUsers {
		for i := range matchResults.Data {
			mr := &matchResults.Data[i]

			mr.UserADetails, err = l.userMatchingDetails(ctx, exec, mr.UserAID)
			if err != nil {
				return nil, fmt.Errorf("enrich user-A by id %s: %w", mr.UserAID.String(), err)
			}

			mr.UserBDetails, err = l.userMatchingDetails(ctx, exec, mr.UserBID)
			if err != nil {
				return nil, fmt.Errorf("enrich user-B by id %s: %w", mr.UserBID.String(), err)
			}

			// Enrich user profiles
			mr.UserAProfile, err = l.profileStorer.Profile(ctx, aiExec, mr.UserAID)
			if err != nil {
				return nil, fmt.Errorf("enrich user-A profile by id %s: %w", mr.UserAID.String(), err)
			}
			mr.UserBProfile, err = l.profileStorer.Profile(ctx, aiExec, mr.UserBID)
			if err != nil {
				return nil, fmt.Errorf("enrich user-B profile by id %s: %w", mr.UserBID.String(), err)
			}
		}
	}

	return matchResults, nil
}

func (l *Logic) MatchResult(ctx context.Context,
	exec boil.ContextExecutor,
	aiExec boil.ContextExecutor,
	f *QueryFilterMatchResult,
) (*MatchResult, error) {
	paginated, err := l.MatchResults(ctx, exec, aiExec, f)
	if err != nil {
		return nil, fmt.Errorf("match results: %w", err)
	}

	if len(paginated.Data) == 0 {
		return nil, fmt.Errorf("no match results found")
	}

	return &paginated.Data[0], nil
}

// MatchResultsForDrop retrieves match results that are ready for drops.
// Criteria: is_approved=true, is_possible_match=true, is_expired=false
func (l *Logic) MatchResultsForDrop(ctx context.Context,
	exec boil.ContextExecutor,
) (*MatchResultPaginated, error) {
	return l.MatchResults(ctx, exec, nil, &QueryFilterMatchResult{
		IsApproved:      null.BoolFrom(true),
		IsPossibleMatch: null.BoolFrom(true),
		IsExpired:       null.BoolFrom(false),
	})
}

// MatchSets retrieves match sets based on the provided filter.
// Supports full pagination and filtering.
func (l *Logic) MatchSets(ctx context.Context,
	exec boil.ContextExecutor,
	f *QueryFilterMatchSet,
) (*MatchSetPaginated, error) {
	matchSets, err := l.matchSetStorer.MatchSets(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("match set storer: %w", err)
	}

	return matchSets, nil
}

// MatchSet retrieves a single match set.
// Returns error if zero or more than one match set found.
func (l *Logic) MatchSet(ctx context.Context,
	exec boil.ContextExecutor,
	f *QueryFilterMatchSet,
) (*MatchSet, error) {
	matchSet, err := l.matchSetStorer.MatchSet(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("match set storer: %w", err)
	}

	return matchSet, nil
}

// MatchConfigs retrieves match configurations based on the provided filter.
func (l *Logic) MatchConfigs(ctx context.Context,
	exec boil.ContextExecutor,
	f *QueryFilterMatchConfig,
) ([]Config, error) {
	configs, err := l.configStorer.Configs(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("config storer: %w", err)
	}

	return configs, nil
}

// MatchConfig retrieves a single match configuration.
// Returns error if zero or more than one config found.
func (l *Logic) MatchConfig(ctx context.Context,
	exec boil.ContextExecutor,
	f *QueryFilterMatchConfig,
) (*Config, error) {
	cfg, err := l.configStorer.Config(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("config storer: %w", err)
	}

	return cfg, nil
}
