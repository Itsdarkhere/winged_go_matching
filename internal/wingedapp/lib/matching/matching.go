package matching

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
)

/*
	This will contain all logic for the matching.
*/

// ProcessMatchResult will process a validatedUserMatchingDetails pair, and return a match result.
// aiExec is the executor for ai_backend database (for profile lookups).
func (l *Logic) ProcessMatchResult(ctx context.Context, exec boil.ContextExecutor, aiExec boil.ContextExecutor, matchResult *MatchResult) (*MatchResult, error) {
	fmt.Printf("[ProcessMatchResult] processing pair: %s <-> %s\n", matchResult.InitiatorUserID, matchResult.ReceiverUserID)

	initiatorUser, err := l.validatedUserMatchingDetails(ctx, exec, matchResult.InitiatorUserID)
	if err != nil {
		fmt.Printf("[ProcessMatchResult] ERROR fetching initiator: %v\n", err)
		return nil, fmt.Errorf("fetch initiator user matching details: %w", err)
	}

	receiverUser, err := l.validatedUserMatchingDetails(ctx, exec, matchResult.ReceiverUserID)
	if err != nil {
		fmt.Printf("[ProcessMatchResult] ERROR fetching receiver: %v\n", err)
		return nil, fmt.Errorf("fetch receiver user matching details: %w", err)
	}

	// TODO: get config from MatchSet, only hesitation is you need fixed versioned fields,
	// because the JSON might drift from the stored config.
	config, err := l.configStorer.Config(ctx, exec, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch config: %w", err)
	}

	qualifierResults := &QualifierResults{}
	hardQualifiers := qualifiers{
		l.ageQualifier,
		l.datePrefsQualifier,
		l.heightQualifier,
		l.distanceQualifier,

		// extend as needed
	}

	res := hardQualifiers.ExecuteAll(ctx, config, qualifierResults, &QualifierParameters{
		config: config,
		UserA:  initiatorUser,
		UserB:  receiverUser,
		// keep expanding this as needed
	})

	// save intermediary qualifier results
	bytesQualifierResults, err := qualifierResults.AsJSON()
	if err != nil {
		return nil, fmt.Errorf("serialize qualifier results: %w", err)
	}
	if matchResult, err = l.matchResultStorer.Update(ctx, exec, &UpdateMatchResult{
		ID:                   matchResult.ID,
		MatchedQualitatively: null.BoolFrom(false), // did not pass hard qualifiers
		QualifierResults:     null.JSONFrom(bytesQualifierResults),
	}); err != nil {
		return nil, fmt.Errorf("updating match set with errors: %w", err)
	}

	if res.Error() != nil {
		fmt.Printf("[ProcessMatchResult] hard qualifiers FAILED: %v\n", res.Error())
		return matchResult, nil
	}

	fmt.Println("[ProcessMatchResult] hard qualifiers PASSED, running qualitative matching...")

	/* test qualitative matches if hard filters passed */
	initiatorProf, err := l.profileStorer.Profile(ctx, aiExec, initiatorUser.ID)
	if err != nil {
		fmt.Printf("[ProcessMatchResult] ERROR fetching initiator profile: %v\n", err)
		return nil, fmt.Errorf("fetch initiator user profile: %w", err)
	}

	receiverProf, err := l.profileStorer.Profile(ctx, aiExec, receiverUser.ID)
	if err != nil {
		fmt.Printf("[ProcessMatchResult] ERROR fetching receiver profile: %v\n", err)
		return nil, fmt.Errorf("fetch receiver user profile: %w", err)
	}

	matchCompatibilityResult, err := l.qualitativeQuantifier.Qualify(ctx, &QualitativeMatchRequest{
		Romeo:  initiatorProf,
		Juliet: receiverProf,
	})
	if err != nil {
		fmt.Printf("[ProcessMatchResult] ERROR qualitative quantify: %v\n", err)
		return nil, fmt.Errorf("qualitative quantify: %w", err)
	}

	bytesMatchCompatibility, err := json.Marshal(matchCompatibilityResult)
	if err != nil {
		return nil, fmt.Errorf("marshal qualitative match result: %w", err)
	}

	// update with qualitative match result
	matchResult, err = l.matchResultStorer.Update(ctx, exec, &UpdateMatchResult{
		ID:                   matchResult.ID,
		MatchedQualitatively: null.BoolFrom(true),
		IsPossibleMatch:      null.BoolFrom(true),
		QualifierResults:     null.JSONFrom(bytesMatchCompatibility),
	})
	if err != nil {
		return nil, fmt.Errorf("match result storer update: %w", err)
	}

	return matchResult, nil
}

// validatedUserMatchingDetails fetches a validatedUserMatchingDetails by uuid, and enriches all related data.
// It also performs validation on the user to ensure it is fit for matching, and has
// no missing critical data.
func (l *Logic) validatedUserMatchingDetails(ctx context.Context, exec boil.ContextExecutor, uuid uuid.UUID) (*User, error) {
	user, err := l.userMatchingDetails(ctx, exec, uuid)
	if err != nil {
		return nil, fmt.Errorf("userByID: %w", err)
	}

	// extra guardrail to keep matching safe
	if err = user.Validate(); err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	return user, nil
}

// RunMatchForUnmatchedUsers creates a new match set for users who don't have
// any pending match results (approved but not yet dropped).
// This ensures users without active matches get new potential pairings.
func (l *Logic) RunMatchForUnmatchedUsers(ctx context.Context, exec boil.ContextExecutor) (*MatchSet, error) {
	unmatchedUsers, err := l.usersWithoutPendingMatches(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("users without pending matches: %w", err)
	}

	if len(unmatchedUsers) < 2 {
		return nil, nil // need at least 2 users to create pairings
	}

	matchSet, err := l.createMatchSetForUsers(ctx, exec, unmatchedUsers)
	if err != nil {
		return nil, fmt.Errorf("create match set for users: %w", err)
	}

	return matchSet, nil
}

// usersWithoutPendingMatches returns active users who don't have any
// match results that are approved but not yet dropped.
func (l *Logic) usersWithoutPendingMatches(ctx context.Context, exec boil.ContextExecutor) ([]User, error) {
	allUsers, err := l.userStorer.Users(ctx, exec, &QueryFilterUser{
		IsActive: null.BoolFrom(true),
	})
	if err != nil {
		return nil, fmt.Errorf("fetch all active users: %w", err)
	}

	pendingMatches, err := l.matchResultStorer.MatchResults(ctx, exec, &QueryFilterMatchResult{
		IsApproved: null.BoolFrom(true),
		IsDropped:  null.BoolFrom(false),
	})
	if err != nil {
		return nil, fmt.Errorf("fetch pending matches: %w", err)
	}

	usersWithPending := l.extractUsersFromMatches(pendingMatches.Data)

	unmatched := make([]User, 0, len(allUsers))
	for _, user := range allUsers {
		if !usersWithPending[user.ID] {
			unmatched = append(unmatched, user)
		}
	}

	return unmatched, nil
}

// extractUsersFromMatches builds a set of user IDs from match results.
func (l *Logic) extractUsersFromMatches(matches []MatchResult) map[uuid.UUID]bool {
	seen := make(map[uuid.UUID]bool)
	for _, mr := range matches {
		seen[mr.InitiatorUserID] = true
		seen[mr.ReceiverUserID] = true
	}
	return seen
}
