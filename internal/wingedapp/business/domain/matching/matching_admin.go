package matching

import (
	"context"
	"fmt"
	jobqueueLib "wingedapp/pgtester/internal/wingedapp/lib/jobqueue"
	matchLib "wingedapp/pgtester/internal/wingedapp/lib/matching"
)

/*
	Admin functions for match result flag control and batch operations.

	Match visibility requires ALL three conditions:
	- is_approved = true
	- is_dropped = true
	- is_expired = false
*/

// BatchIngestResult contains the result of an async batch ingestion request.
type BatchIngestResult struct {
	MatchSet *matchLib.MatchSet `json:"match_set"`
	Job      *jobqueueLib.Job   `json:"job"`
}

// ============================================================================
// MATCH FLAG CONTROL ENDPOINTS
// Single-flag operations for granular admin control
// ============================================================================

// SetMatchApproved sets is_approved=true for a match.
func (b *Business) SetMatchApproved(ctx context.Context, matchResultID string) error {
	tx, err := b.transactor.TX()
	if err != nil {
		return fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	if err = b.approvalSetter.SetApproved(ctx, tx, matchResultID); err != nil {
		return fmt.Errorf("set approved: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// SetMatchUnapproved sets is_approved=false for a match.
func (b *Business) SetMatchUnapproved(ctx context.Context, matchResultID string) error {
	tx, err := b.transactor.TX()
	if err != nil {
		return fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	if err = b.approvalSetter.SetUnapproved(ctx, tx, matchResultID); err != nil {
		return fmt.Errorf("set unapproved: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// SetMatchDropped sets is_dropped=true for a match.
func (b *Business) SetMatchDropped(ctx context.Context, matchResultID string) error {
	tx, err := b.transactor.TX()
	if err != nil {
		return fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	if err = b.dropSetter.SetDropped(ctx, tx, matchResultID); err != nil {
		return fmt.Errorf("set dropped: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// SetMatchUndropped sets is_dropped=false for a match.
func (b *Business) SetMatchUndropped(ctx context.Context, matchResultID string) error {
	tx, err := b.transactor.TX()
	if err != nil {
		return fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	if err = b.dropSetter.SetUndropped(ctx, tx, matchResultID); err != nil {
		return fmt.Errorf("set undropped: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// SetMatchExpired sets is_expired=true for a match.
func (b *Business) SetMatchExpired(ctx context.Context, matchResultID string) error {
	tx, err := b.transactor.TX()
	if err != nil {
		return fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	if err = b.expirySetter.SetExpired(ctx, tx, matchResultID); err != nil {
		return fmt.Errorf("set expired: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// SetMatchUnexpired sets is_expired=false for a match.
func (b *Business) SetMatchUnexpired(ctx context.Context, matchResultID string) error {
	tx, err := b.transactor.TX()
	if err != nil {
		return fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	if err = b.expirySetter.SetUnexpired(ctx, tx, matchResultID); err != nil {
		return fmt.Errorf("set unexpired: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// ============================================================================
// LEGACY METHODS (backwards compatibility)
// ============================================================================

// ApproveMatch sets is_approved=true (legacy - use SetMatchApproved)
func (b *Business) ApproveMatch(ctx context.Context, matchResultID string) error {
	return b.SetMatchApproved(ctx, matchResultID)
}

// UnapproveMatch sets is_approved=false (legacy - use SetMatchUnapproved)
func (b *Business) UnapproveMatch(ctx context.Context, matchResultID string) error {
	return b.SetMatchUnapproved(ctx, matchResultID)
}

// ============================================================================
// ADMIN QUERY METHODS
// ============================================================================

// AdminMatchViews returns all match results for admin review
func (b *Business) AdminMatchViews(
	ctx context.Context,
	filter *matchLib.QueryFilterMatchResult,
) (*matchLib.MatchResultPaginated, error) {
	exec := b.transactor.DB()
	aiExec := b.transactorAI.DB()

	results, err := b.matchGetter.MatchResults(ctx, exec, aiExec, filter)
	if err != nil {
		return nil, fmt.Errorf("match results: %w", err)
	}

	return results, nil
}

// AdminMatchResult returns a single match result by ID for admin review
func (b *Business) AdminMatchResult(
	ctx context.Context,
	filter *matchLib.QueryFilterMatchResult,
) (*matchLib.MatchResult, error) {
	exec := b.transactor.DB()
	aiExec := b.transactorAI.DB()

	result, err := b.matchGetter.MatchResult(ctx, exec, aiExec, filter)
	if err != nil {
		return nil, fmt.Errorf("match result: %w", err)
	}

	return result, nil
}

// AdminMatchSets returns match sets with pagination and filtering (admin only)
func (b *Business) AdminMatchSets(
	ctx context.Context,
	filter *matchLib.QueryFilterMatchSet,
) (*matchLib.MatchSetPaginated, error) {
	exec := b.transactor.DB()

	results, err := b.matchSetGetter.MatchSets(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("match sets: %w", err)
	}

	return results, nil
}

// AdminGetConfig returns the current match configuration (admin only)
func (b *Business) AdminGetConfig(
	ctx context.Context,
) (*matchLib.Config, error) {
	config, err := b.configGetter.Config(ctx, b.transactor.DB(), nil)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	return config, nil
}

// AdminUpdateConfig updates the match configuration (admin only)
func (b *Business) AdminUpdateConfig(
	ctx context.Context,
	updater *matchLib.UpdateMatchConfig,
) (*matchLib.Config, error) {
	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	config, err := b.configUpdater.UpdateConfig(ctx, tx, updater)
	if err != nil {
		return nil, fmt.Errorf("update config: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return config, nil
}

// ============================================================================
// BATCH INGESTION
// ============================================================================

// AdminIngestAll runs the batch ingestion process for all active users (admin only).
// This assembles two operations:
// 1. IngestAll - creates MatchSet and MatchResults for all active user pairs
// 2. Enqueues a job to run the matching algorithm asynchronously
func (b *Business) AdminIngestAll(
	ctx context.Context,
) (*BatchIngestResult, error) {
	return b.AdminIngestWithOptions(ctx, nil)
}

// AdminIngestWithOptions runs the batch ingestion process with optional exclusion filters.
// If options is nil, all active users are included (same as AdminIngestAll).
//
// Phase 1 (IngestWithOptions) uses a transaction to atomically create the MatchSet and MatchResults.
// Phase 2: Enqueues a job for async processing - the worker will run the matching algorithm.
func (b *Business) AdminIngestWithOptions(
	ctx context.Context,
	options *matchLib.BatchIngestOptions,
) (*BatchIngestResult, error) {
	// Phase 1: Create MatchSet + MatchResults atomically in a transaction
	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	matchSet, err := b.ingester.IngestWithOptions(ctx, tx, options)
	if err != nil {
		return nil, fmt.Errorf("ingest with options: %w", err)
	}

	// Extract IsTestUser from options for job payload
	var isTestUser *bool
	if options != nil {
		isTestUser = options.IsTestUser
	}

	// Phase 2: Enqueue job for async processing (same TX for atomicity)
	job, err := b.jobEnqueuer.EnqueueBatchMatch(ctx, tx, matchSet.ID.String(), isTestUser)
	if err != nil {
		return nil, fmt.Errorf("enqueue batch match: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &BatchIngestResult{
		MatchSet: matchSet,
		Job:      job,
	}, nil
}

// AdminPopulateFromCSV populates users across all databases from parsed CSV data.
// This coordinates insertions to:
//   - backend_app.users (with dating preferences)
//   - ai_backend.profiles
//   - supabase auth.users
//
// If options.IsTestUser is true, all users will be marked as test users.
// Returns a PopulateResult with counts and any errors encountered.
// Note: Each DB uses its own transaction for atomicity within that DB.
func (b *Business) AdminPopulateFromCSV(
	ctx context.Context,
	rows []matchLib.PopulationRow,
	options *matchLib.PopulateOptions,
) (*matchLib.PopulateResult, error) {
	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	txAI, err := b.transactorAI.TX()
	if err != nil {
		return nil, fmt.Errorf("tx ai: %w", err)
	}
	defer b.transactorAI.Rollback(txAI)

	txSBAuth, err := b.transactorSBAuth.TX()
	if err != nil {
		return nil, fmt.Errorf("tx supabase auth: %w", err)
	}
	defer b.transactorSBAuth.Rollback(txSBAuth)

	execs := &matchLib.PopulationExecutors{
		BackendApp:   tx,
		AIBackend:    txAI,
		SupabaseAuth: txSBAuth,
	}

	result, err := b.populator.Populate(ctx, execs, rows, options)
	if err != nil {
		return nil, fmt.Errorf("populate: %w", err)
	}

	// Commit all transactions - each transactor has its own connection
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit backend: %w", err)
	}
	if err = txAI.Commit(); err != nil {
		return nil, fmt.Errorf("commit ai: %w", err)
	}
	if err = txSBAuth.Commit(); err != nil {
		return nil, fmt.Errorf("commit supabase auth: %w", err)
	}

	return result, nil
}
