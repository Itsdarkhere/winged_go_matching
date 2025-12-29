package matching

import (
	"context"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type profileStorer interface {
	Profile(ctx context.Context, exec boil.ContextExecutor, userID uuid.UUID) (*PersonProfile, error)
	Insert(ctx context.Context, exec boil.ContextExecutor, params *InsertPopulationProfile) error
	DeleteByUserIDs(ctx context.Context, exec boil.ContextExecutor, userIDs []string) (int64, error)
}

type userDatingPreferenceStorer interface {
	UserDatingPreferences(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterUserDatingPrefs) ([]UserDatingPreference, error)
	Insert(ctx context.Context, exec boil.ContextExecutor, userID string, preferences []string) error
}

// supabaseUserStorer is the interface for supabase user operations.
type supabaseUserStorer interface {
	Insert(ctx context.Context, exec boil.ContextExecutor, params *InsertSupabaseUser) (string, error)
	DeleteByEmails(ctx context.Context, exec boil.ContextExecutor, emails []string) (int64, error)
}

// userStorer is the interface to store and retrieve users.
type userStorer interface {
	Users(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterUser) ([]User, error)
	User(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterUser) (*User, error)
	Insert(ctx context.Context, exec boil.ContextExecutor, params *InsertPopulationUser) (string, error)
	DeleteByEmails(ctx context.Context, exec boil.ContextExecutor, emails []string) (int64, error)
}

// configStorer is the interface to store and retrieve match configurations.
type configStorer interface {
	Configs(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterMatchConfig) ([]Config, error)
	Config(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterMatchConfig) (*Config, error)
	Update(ctx context.Context, exec boil.ContextExecutor, updater *UpdateMatchConfig) (*Config, error)
}

// matchSetStorer is the interface to store and retrieve match sets.
type matchSetStorer interface {
	MatchSets(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterMatchSet) (*MatchSetPaginated, error)
	MatchSet(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterMatchSet) (*MatchSet, error)
	Insert(ctx context.Context, exec boil.ContextExecutor, inserter *InsertMatchSet) (*MatchSet, error)
	Update(ctx context.Context, exec boil.ContextExecutor, updater *UpdateMatchSet) error
}

// matchResultStorer is the interface to store and retrieve match participants.
type matchResultStorer interface {
	MatchResults(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterMatchResult) (*MatchResultPaginated, error)
	MatchResult(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterMatchResult) (*MatchResult, error)
	Insert(ctx context.Context, exec boil.ContextExecutor, inserter *InsertMatchResult) (*MatchResult, error)
	Update(ctx context.Context, exec boil.ContextExecutor, inserter *UpdateMatchResult) (*MatchResult, error)
}

type QualitativeMatchRequest = ProfileData

// qualitativeQuantifier fetches a qualitative match from an external service.
//
//counterfeiter:generate . qualitativeQuantifier
type qualitativeQuantifier interface {
	Qualify(ctx context.Context, req *QualitativeMatchRequest) (*MatchCompatibilityResult, error)
}

// userMatchActionsStorer handles user-facing match data access (CRUD only).
type userMatchActionsStorer interface {
	UserMatches(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterUserMatch) ([]UserMatch, error)
	UserMatchesPaginated(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterUserMatch) (*UserMatchPaginated, error)
	UserMatch(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterUserMatch) (*UserMatch, error)
	Update(ctx context.Context, exec boil.ContextExecutor, params *UpdateUserMatchAction) error
	UpdateSeenBatch(ctx context.Context, exec boil.ContextExecutor, userID uuid.UUID, matchIDs []uuid.UUID) error
}

// audioGetter fetches intro audio URLs from ai_backend.
type audioGetter interface {
	IntroAudioURL(ctx context.Context, exec boil.ContextExecutor, supabaseUserID uuid.UUID) null.String
	IntroAudioURLs(ctx context.Context, exec boil.ContextExecutor, supabaseUserIDs []uuid.UUID) (map[uuid.UUID]null.String, error)
}

// lovestoryGetter fetches lovestory audio URLs from ai_backend.lovestory table.
type lovestoryGetter interface {
	LovestoryURLsByMatchIDs(ctx context.Context, exec boil.ContextExecutor, matchResultIDs []uuid.UUID) (map[uuid.UUID]null.String, error)
}

// publicURLer converts storage keys to public URLs.
// Used for Supabase storage paths that need to be converted to accessible URLs.
//
//counterfeiter:generate . publicURLer
type publicURLer interface {
	PublicURL(ctx context.Context, key string) (string, error)
}

// userDeleter handles cascading user data deletion across all tables.
type userDeleter interface {
	// DeleteUserDataByEmail deletes all user data for users matching the given emails.
	// This includes all FK-dependent tables (match results, chat messages, audio_files, supabase auth.users, etc.).
	// supabaseExec is optional - pass nil to skip supabase auth.users deletion.
	DeleteUserDataByEmail(ctx context.Context, beExec, aiExec, supabaseExec boil.ContextExecutor, emails []string) (int64, error)
}

// dateInstanceInserter handles date instance creation.
// For Insert/Update, use db/repo.Store.
type dateInstanceInserter interface {
	InsertDateInstance(ctx context.Context, exec boil.ContextExecutor, inserter *InsertDateInstance) (string, error)
	InsertDateInstanceLog(ctx context.Context, exec boil.ContextExecutor, inserter *InsertDateInstanceLog) error
}

// matchResultUpdater handles match result updates for date instance linking.
type matchResultUpdater interface {
	UpdateMatchForDateInstance(ctx context.Context, exec boil.ContextExecutor, updater *UpdateMatchForDateInstance) error
}
