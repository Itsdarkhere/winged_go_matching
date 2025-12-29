package matching

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/alitto/pond"
	"github.com/google/uuid"
)

const (
	// batchMatchWorkers is the number of concurrent workers for batch matching.
	batchMatchWorkers = 50
)

/*
	This will contain all ingestion related logic.
*/

// BatchIngestOptions controls which users to include in batch matching.
// By default (nil options), all active users are included.
type BatchIngestOptions struct {
	// IsTestUser filters users by test user status:
	// - nil: include all users (default, backwards compatible)
	// - true: include ONLY test users (is_test_user = true)
	// - false: include ONLY non-test users (is_test_user = false)
	IsTestUser *bool
}

// Ingest will load users matching the given filters into a new match set.
func (l *Logic) Ingest(ctx context.Context, exec boil.ContextExecutor, userFilters *QueryFilterUser) (*MatchSet, error) {
	matchSet, err := l.ingestUsers(ctx, exec, userFilters)
	if err != nil {
		return nil, fmt.Errorf("ingest all users: %w", err)
	}

	return matchSet, nil
}

// IngestAll will load all users into a new match set.
// This will be the most used function in the alpha version that will,
// process all users against each other to see if the algorithm works.
// Note: This only creates the MatchSet and MatchResults - to run the matching
// algorithm, call RunIngestionSet separately or use the business layer which
// assembles both operations.
func (l *Logic) IngestAll(ctx context.Context, exec boil.ContextExecutor) (*MatchSet, error) {
	return l.IngestWithOptions(ctx, exec, nil)
}

// IngestWithOptions loads users into a new match set with optional filters.
// If options is nil, all active users are included (same as IngestAll).
func (l *Logic) IngestWithOptions(ctx context.Context, exec boil.ContextExecutor, options *BatchIngestOptions) (*MatchSet, error) {
	filter := &QueryFilterUser{
		IsActive: null.BoolFrom(true), // always filter for active users
	}

	// Apply is_test_user filter if provided
	if options != nil && options.IsTestUser != nil {
		filter.IsTestUser = null.BoolFrom(*options.IsTestUser)
	}

	matchSet, err := l.ingestUsers(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("ingest users: %w", err)
	}

	return matchSet, nil
}

// ingestUsers will load users matching the given filters into a new match set.
func (l *Logic) ingestUsers(ctx context.Context, exec boil.ContextExecutor, userFilters *QueryFilterUser) (*MatchSet, error) {
	users, err := l.userStorer.Users(ctx, exec, userFilters)
	if err != nil {
		return nil, fmt.Errorf("all users: %w", err)
	}

	matchSet, err := l.createMatchSetForUsers(ctx, exec, users)
	if err != nil {
		return nil, fmt.Errorf("create match set: %w", err)
	}

	return matchSet, nil
}

// createMatchSetForUsers creates a new match set and generates all unique
// pairings for the given users.
func (l *Logic) createMatchSetForUsers(ctx context.Context, exec boil.ContextExecutor, users []User) (*MatchSet, error) {
	settings, err := l.configStorer.Config(ctx, exec, nil)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}
	bytesSettings, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	// create match set
	matchSet, err := l.matchSetStorer.Insert(ctx, exec, &InsertMatchSet{
		Name:                 time.Now().Format(time.RFC3339),
		MatchingParameters:   bytesSettings,
		NumberOfParticipants: len(users),
	})
	if err != nil {
		return nil, fmt.Errorf("insert match set: %w", err)
	}

	// create match results
	for _, up := range UserPairsUniqPerm(users) {
		if _, err = l.matchResultStorer.Insert(ctx, exec, &InsertMatchResult{
			MatchSetID: matchSet.ID,
			UserAID:    up.UserA.ID,
			UserBID:    up.UserB.ID,
		}); err != nil {
			return nil, fmt.Errorf("insert match participant: %w", err)
		}
	}

	return matchSet, nil
}

// AllUsers will return all users in the matching engine.
// This can be optimised to return via cursor, and pre-filter by lat,
// and long.
func (l *Logic) allUsers(ctx context.Context, exec boil.ContextExecutor) ([]User, error) {
	users, err := l.userStorer.Users(ctx, exec, &QueryFilterUser{
		IsActive: null.BoolFrom(true),
	})
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}

	return users, nil
}

// RunIngestionSet runs the matching algorithm on all MatchResults in a given MatchSet.
// It processes each pair through ProcessMatchResult which evaluates hard qualifiers
// (age, dating prefs, height, distance) and qualitative matching if hard qualifiers pass.
// Processing continues even if individual pairs fail - errors are logged but don't stop the batch.
//
// Uses pond worker pool with batchMatchWorkers (50) concurrent workers for parallel processing.
// IMPORTANT: exec MUST be a DB pool (not a transaction) for safe concurrent access.
// Each goroutine will get its own connection from the pool.
//
// aiExec is the executor for ai_backend database (for profile lookups).
func (l *Logic) RunIngestionSet(ctx context.Context, exec boil.ContextExecutor, aiExec boil.ContextExecutor, matchSetID uuid.UUID) error {
	// Fetch all MatchResults for this MatchSet
	results, err := l.matchResultStorer.MatchResults(ctx, exec, &QueryFilterMatchResult{
		MatchSetID: null.StringFrom(matchSetID.String()),
	})
	if err != nil {
		return fmt.Errorf("fetch match results for set %s: %w", matchSetID, err)
	}

	// Process each MatchResult through the matching algorithm concurrently.
	// Using pond worker pool for parallel processing - each goroutine gets its own
	// connection from the DB pool, enabling safe concurrent database access.
	p := pond.New(batchMatchWorkers, len(results.Data))
	for i := range results.Data {
		result := &results.Data[i]
		p.Submit(func() {
			// Errors are logged but don't stop the batch - one bad pair shouldn't stop processing.
			// Error details are stored in QualifierResults for visibility.
			_, _ = l.ProcessMatchResult(ctx, exec, aiExec, result)
		})
	}
	p.StopAndWait()

	return nil
}
