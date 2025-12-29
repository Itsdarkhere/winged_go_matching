package matching

import (
	"context"
	economyLib "wingedapp/pgtester/internal/wingedapp/lib/economy"
	jobqueueLib "wingedapp/pgtester/internal/wingedapp/lib/jobqueue"
	matchLib "wingedapp/pgtester/internal/wingedapp/lib/matching"
	schedulingLib "wingedapp/pgtester/internal/wingedapp/lib/scheduling"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
)

type transactor interface {
	TX() (boil.ContextTransactor, error)
	Rollback(boil.ContextTransactor)
	DB() boil.ContextExecutor
}

// matchGetter contains all the methods required to get matches.
// Must have full pagination support.
type matchGetter interface {
	MatchResults(ctx context.Context, exec boil.ContextExecutor, aiExec boil.ContextExecutor, f *matchLib.QueryFilterMatchResult) (*matchLib.MatchResultPaginated, error)
	MatchResult(ctx context.Context, exec boil.ContextExecutor, aiExec boil.ContextExecutor, f *matchLib.QueryFilterMatchResult) (*matchLib.MatchResult, error)
}

// approvalSetter controls is_approved flag on match results.
type approvalSetter interface {
	SetApproved(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error
	SetUnapproved(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error
	// Legacy
	Approve(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error
	Unapprove(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error
}

// dropSetter controls is_dropped flag on match results.
type dropSetter interface {
	SetDropped(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error
	SetUndropped(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error
}

// expirySetter controls is_expired flag on match results.
type expirySetter interface {
	SetExpired(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error
	SetUnexpired(ctx context.Context, exec boil.ContextExecutor, matchResultID string) error
}

// configGetter contains methods to get match configuration
type configGetter interface {
	Config(ctx context.Context, exec boil.ContextExecutor, f *matchLib.QueryFilterMatchConfig) (*matchLib.Config, error)
}

// matchSetGetter contains methods to get match sets with pagination
type matchSetGetter interface {
	MatchSets(ctx context.Context, exec boil.ContextExecutor, f *matchLib.QueryFilterMatchSet) (*matchLib.MatchSetPaginated, error)
	MatchSet(ctx context.Context, exec boil.ContextExecutor, f *matchLib.QueryFilterMatchSet) (*matchLib.MatchSet, error)
}

// configUpdater contains methods to update match configuration
type configUpdater interface {
	UpdateConfig(ctx context.Context, exec boil.ContextExecutor, updater *matchLib.UpdateMatchConfig) (*matchLib.Config, error)
}

// ingester contains methods to run the matching ingestion process
type ingester interface {
	IngestAll(ctx context.Context, exec boil.ContextExecutor) (*matchLib.MatchSet, error)
	IngestWithOptions(ctx context.Context, exec boil.ContextExecutor, options *matchLib.BatchIngestOptions) (*matchLib.MatchSet, error)
}

// matchRunner contains methods to run the matching algorithm on ingested sets
type matchRunner interface {
	RunIngestionSet(ctx context.Context, exec boil.ContextExecutor, aiExec boil.ContextExecutor, matchSetID uuid.UUID) error
}

// populator contains methods to populate users from CSV data
type populator interface {
	Populate(ctx context.Context, execs *matchLib.PopulationExecutors, rows []matchLib.PopulationRow, options *matchLib.PopulateOptions) (*matchLib.PopulateResult, error)
}

// userMatchGetter handles getting user match data
type userMatchGetter interface {
	UserMatches(ctx context.Context, exec boil.ContextExecutor, aiExec boil.ContextExecutor, f *matchLib.QueryFilterUserMatch) ([]matchLib.UserMatch, error)
	UserMatchesPaginated(ctx context.Context, exec boil.ContextExecutor, aiExec boil.ContextExecutor, f *matchLib.QueryFilterUserMatch) (*matchLib.UserMatchPaginated, error)
	UserMatch(ctx context.Context, exec boil.ContextExecutor, aiExec boil.ContextExecutor, f *matchLib.QueryFilterUserMatch) (*matchLib.UserMatch, error)
	UnseenMatchCount(ctx context.Context, exec boil.ContextExecutor, userID uuid.UUID) (int64, error)
}

// userMatchActioner handles user-facing match actions (propose, pass, seen)
type userMatchActioner interface {
	ProposeMatch(ctx context.Context, exec boil.ContextExecutor, params *matchLib.ProposeMatchParams) (*matchLib.ProposeMatchResult, error)
	PassMatch(ctx context.Context, exec boil.ContextExecutor, params *matchLib.PassMatchParams) error
	MarkMatchesSeen(ctx context.Context, exec boil.ContextExecutor, params *matchLib.MarkSeenParams) error
}

// schedulingLogic handles scheduling overlap operations
type schedulingLogic interface {
	FindOverlaps(ctx context.Context, userAID, userBID string) ([]schedulingLib.TimeBlock, error)
}

// jobEnqueuer handles job queue operations for async batch processing
type jobEnqueuer interface {
	EnqueueBatchMatch(ctx context.Context, exec boil.ContextExecutor, matchSetID string, isTestUser *bool) (*jobqueueLib.Job, error)
}

// actionLogger handles economy action logging (referral bonuses, etc.)
// Lib handles all internal logic (user lookup, idempotency, credit referrer).
type actionLogger interface {
	CreateActionLog(ctx context.Context, exec boil.ContextExecutor, inserter *economyLib.InsertActionLog) error
}
