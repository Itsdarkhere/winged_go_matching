package matching

import (
	"context"
	"fmt"
	"reflect"
	"time"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"

	"wingedapp/pgtester/internal/wingedapp/lib/enums"
)

// isNilExecutor checks if the executor interface is nil or has a nil underlying value.
func isNilExecutor(exec boil.ContextExecutor) bool {
	if exec == nil {
		return true
	}
	v := reflect.ValueOf(exec)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

// timeNow is a variable for testing purposes
var timeNow = time.Now

type Logic struct {
	userStorer                 userStorer
	userDatingPreferenceStorer userDatingPreferenceStorer
	matchSetStorer             matchSetStorer
	matchResultStorer          matchResultStorer
	qualitativeQuantifier      qualitativeQuantifier
	configStorer               configStorer
	profileStorer              profileStorer
	supabaseUserStorer         supabaseUserStorer
	userMatchActionsStorer     userMatchActionsStorer
	audioGetter                audioGetter
	lovestoryGetter            lovestoryGetter
	aiPublicURLer              publicURLer
	userDeleter                userDeleter

	// Date instance dependencies (Tier 1)
	dateInstanceInserter dateInstanceInserter
	matchResultUpdater   matchResultUpdater
}

func NewLogic(
	configStorer configStorer,
	userStorer userStorer,
	userDatingPreferenceStorer userDatingPreferenceStorer,
	matchSetStorer matchSetStorer,
	matchResultStorer matchResultStorer,
	qualitativeQuantifier qualitativeQuantifier,
	profileStorer profileStorer,
	supabaseUserStorer supabaseUserStorer,
	userMatchActionsStorer userMatchActionsStorer,
	audioGetter audioGetter,
	lovestoryGetter lovestoryGetter,
	aiPublicURLer publicURLer,
	userDeleter userDeleter,
	dateInstanceInserter dateInstanceInserter,
	matchResultUpdater matchResultUpdater,
) (*Logic, error) {
	if configStorer == nil {
		return nil, fmt.Errorf("configStorer is required")
	}
	if userStorer == nil {
		return nil, fmt.Errorf("userStorer is required")
	}
	if userDatingPreferenceStorer == nil {
		return nil, fmt.Errorf("userDatingPreferenceStorer is required")
	}
	if matchSetStorer == nil {
		return nil, fmt.Errorf("matchSetStorer is required")
	}
	if matchResultStorer == nil {
		return nil, fmt.Errorf("matchResultStorer is required")
	}
	if qualitativeQuantifier == nil {
		return nil, fmt.Errorf("qualitativeQuantifier is required")
	}
	if profileStorer == nil {
		return nil, fmt.Errorf("profileStorer is required")
	}
	if audioGetter == nil {
		return nil, fmt.Errorf("audioGetter is required")
	}
	if lovestoryGetter == nil {
		return nil, fmt.Errorf("lovestoryGetter is required")
	}
	if aiPublicURLer == nil {
		return nil, fmt.Errorf("aiPublicURLer is required")
	}
	// supabaseUserStorer, userMatchActionsStorer, userDeleter, dateInstanceInserter, matchResultUpdater are optional

	return &Logic{
		configStorer:               configStorer,
		userStorer:                 userStorer,
		userDatingPreferenceStorer: userDatingPreferenceStorer,
		matchSetStorer:             matchSetStorer,
		matchResultStorer:          matchResultStorer,
		qualitativeQuantifier:      qualitativeQuantifier,
		profileStorer:              profileStorer,
		supabaseUserStorer:         supabaseUserStorer,
		userMatchActionsStorer:     userMatchActionsStorer,
		audioGetter:                audioGetter,
		lovestoryGetter:            lovestoryGetter,
		aiPublicURLer:              aiPublicURLer,
		userDeleter:                userDeleter,
		dateInstanceInserter:       dateInstanceInserter,
		matchResultUpdater:         matchResultUpdater,
	}, nil
}

// SetQualitativeQuantifier sets the qualitativeQuantifier implementation.
func (l *Logic) SetQualitativeQuantifier(q qualitativeQuantifier) {
	l.qualitativeQuantifier = q
}

// SetAIPublicURLer sets the aiPublicURLer implementation.
func (l *Logic) SetAIPublicURLer(p publicURLer) {
	l.aiPublicURLer = p
}

// Config returns the match configuration (delegates to configStorer).
func (l *Logic) Config(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *QueryFilterMatchConfig,
) (*Config, error) {
	return l.configStorer.Config(ctx, exec, f)
}

// UserMatches returns user's dropped matches from their perspective.
func (l *Logic) UserMatches(
	ctx context.Context,
	exec boil.ContextExecutor,
	aiExec boil.ContextExecutor,
	f *QueryFilterUserMatch,
) ([]UserMatch, error) {
	matches, err := l.userMatchActionsStorer.UserMatches(ctx, exec, f)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return matches, nil
	}

	// Enrich audio URLs if requested
	if f.EnrichAudio {
		l.enrichMatchAudio(ctx, aiExec, matches)
	}

	// Enrich lovestory URLs if requested
	if f.EnrichLovestory {
		l.enrichLovestoryURLs(ctx, aiExec, matches)
	}

	return matches, nil
}

// UserMatchesPaginated returns user's matches with pagination info.
func (l *Logic) UserMatchesPaginated(
	ctx context.Context,
	exec boil.ContextExecutor,
	aiExec boil.ContextExecutor,
	f *QueryFilterUserMatch,
) (*UserMatchPaginated, error) {
	result, err := l.userMatchActionsStorer.UserMatchesPaginated(ctx, exec, f)
	if err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return result, nil
	}

	// Enrich audio URLs if requested
	if f.EnrichAudio {
		l.enrichMatchAudio(ctx, aiExec, result.Data)
	}

	// Enrich lovestory URLs if requested
	if f.EnrichLovestory {
		l.enrichLovestoryURLs(ctx, aiExec, result.Data)
	}

	return result, nil
}

// UserMatch returns a single match for a user.
func (l *Logic) UserMatch(
	ctx context.Context,
	exec boil.ContextExecutor,
	aiExec boil.ContextExecutor,
	f *QueryFilterUserMatch,
) (*UserMatch, error) {
	match, err := l.userMatchActionsStorer.UserMatch(ctx, exec, f)
	if err != nil {
		return nil, err
	}

	// Use slice for enrichment functions (mutate pattern)
	matches := []UserMatch{*match}

	// Enrich audio URLs if requested
	if f.EnrichAudio {
		l.enrichMatchAudio(ctx, aiExec, matches)
		match.YourIntroURL = matches[0].YourIntroURL
		match.PartnerIntroURL = matches[0].PartnerIntroURL
	}

	// Enrich lovestory URL if requested
	if f.EnrichLovestory {
		l.enrichLovestoryURLs(ctx, aiExec, matches)
		match.LovestoryURL = matches[0].LovestoryURL
	}

	return match, nil
}

// enrichMatchAudio enriches matches with intro audio URLs.
func (l *Logic) enrichMatchAudio(ctx context.Context, exec boil.ContextExecutor, matches []UserMatch) {
	// Skip if no AI database executor is provided
	if isNilExecutor(exec) {
		return
	}

	// Collect all unique supabase IDs
	supabaseIDs := make([]uuid.UUID, 0, len(matches)*2)
	seen := make(map[uuid.UUID]bool)

	for _, m := range matches {
		if m.YourSupabaseID.Valid {
			id, err := uuid.Parse(m.YourSupabaseID.String)
			if err == nil && !seen[id] {
				supabaseIDs = append(supabaseIDs, id)
				seen[id] = true
			}
		}
		if m.PartnerSupabaseID.Valid {
			id, err := uuid.Parse(m.PartnerSupabaseID.String)
			if err == nil && !seen[id] {
				supabaseIDs = append(supabaseIDs, id)
				seen[id] = true
			}
		}
	}

	if len(supabaseIDs) == 0 {
		return
	}

	// Batch fetch audio URLs
	audioURLs, err := l.audioGetter.IntroAudioURLs(ctx, exec, supabaseIDs)
	if err != nil {
		// Log and continue - audio is optional enrichment
		return
	}

	// Apply to matches
	for i := range matches {
		if matches[i].YourSupabaseID.Valid {
			id, _ := uuid.Parse(matches[i].YourSupabaseID.String)
			if url, ok := audioURLs[id]; ok {
				matches[i].YourIntroURL = url
			}
		}
		if matches[i].PartnerSupabaseID.Valid {
			id, _ := uuid.Parse(matches[i].PartnerSupabaseID.String)
			if url, ok := audioURLs[id]; ok {
				matches[i].PartnerIntroURL = url
			}
		}
	}
}

// enrichLovestoryURLs enriches matches with lovestory (first date simulation) audio URLs.
// Storage paths from the lovestory table are converted to public URLs via aiPublicURLer.
func (l *Logic) enrichLovestoryURLs(ctx context.Context, exec boil.ContextExecutor, matches []UserMatch) {
	// Skip if no AI database executor is provided
	if isNilExecutor(exec) {
		return
	}

	// Collect all match IDs
	matchIDs := make([]uuid.UUID, 0, len(matches))
	for _, m := range matches {
		matchIDs = append(matchIDs, m.ID)
	}

	if len(matchIDs) == 0 {
		return
	}

	// Batch fetch lovestory storage paths
	lovestoryPaths, err := l.lovestoryGetter.LovestoryURLsByMatchIDs(ctx, exec, matchIDs)
	if err != nil {
		// Log and continue - lovestory is optional enrichment
		return
	}

	// Apply to matches, converting storage paths to public URLs
	for i := range matches {
		if storagePath, ok := lovestoryPaths[matches[i].ID]; ok && storagePath.Valid && storagePath.String != "" {
			publicURL, err := l.aiPublicURLer.PublicURL(ctx, storagePath.String)
			if err != nil {
				continue
			}
			matches[i].LovestoryURL = null.StringFrom(publicURL)
		}
	}
}

// ProposeMatch sets the user's action to Proposed on the match.
// On mutual proposal (both users proposed), auto-creates a date_instance and
// transitions the match to "Scheduling" lifecycle status.
func (l *Logic) ProposeMatch(
	ctx context.Context,
	exec boil.ContextExecutor,
	params *ProposeMatchParams,
) (*ProposeMatchResult, error) {
	// Lock the match row to prevent concurrent proposals (SELECT FOR UPDATE)
	filter := &QueryFilterUserMatch{
		UserID:  params.UserID,
		MatchID: null.StringFrom(params.MatchResultID.String()),
	}
	match, err := l.userMatchActionsStorer.UserMatch(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("get match: %w", err)
	}

	// Update the user's action
	updateParams := &UpdateUserMatchAction{
		MatchResultID:    params.MatchResultID,
		UserID:           params.UserID,
		ActionCategoryID: null.StringFrom(string(enums.MatchUserActionProposed)),
	}

	// Check if partner already proposed - if so, this is a mutual proposal
	result := &ProposeMatchResult{Success: true}
	partnerAlreadyProposed := match.PartnerAction == MatchUserActionProposed

	if partnerAlreadyProposed {
		result.MutualProposal = true

		// Create date_instance on mutual proposal (if dependencies available)
		if l.dateInstanceInserter != nil && l.matchResultUpdater != nil {
			dateInstanceID, err := l.createDateInstanceOnMutualProposal(ctx, exec, params.MatchResultID.String())
			if err != nil {
				return nil, fmt.Errorf("create date instance: %w", err)
			}
			result.DateInstanceID = dateInstanceID
		}
	}

	if err := l.userMatchActionsStorer.Update(ctx, exec, updateParams); err != nil {
		return nil, fmt.Errorf("update match action: %w", err)
	}

	// Post-update reconciliation: Check if we became mutually proposed
	// This handles edge case where partner proposed between our read and update
	// (shouldn't happen with row lock, but belt-and-suspenders)
	if !partnerAlreadyProposed {
		matchAfter, err := l.userMatchActionsStorer.UserMatch(ctx, exec, filter)
		if err != nil {
			return nil, fmt.Errorf("reconciliation read: %w", err)
		}

		// Check if BOTH users now show as proposed
		if matchAfter.YourAction == MatchUserActionProposed &&
			matchAfter.PartnerAction == MatchUserActionProposed {
			result.MutualProposal = true

			// Create date_instance if not already created
			if l.dateInstanceInserter != nil && l.matchResultUpdater != nil && result.DateInstanceID == "" {
				dateInstanceID, err := l.createDateInstanceOnMutualProposal(ctx, exec, params.MatchResultID.String())
				if err != nil {
					return nil, fmt.Errorf("reconciliation create date instance: %w", err)
				}
				result.DateInstanceID = dateInstanceID
			}
		}
	}

	return result, nil
}

// createDateInstanceOnMutualProposal creates a date_instance when both users have proposed.
func (l *Logic) createDateInstanceOnMutualProposal(
	ctx context.Context,
	exec boil.ContextExecutor,
	matchResultID string,
) (string, error) {
	// 0. Check if date_instance already exists (idempotency)
	existingMatchResult, err := pgmodel.FindMatchResult(ctx, exec, matchResultID)
	if err != nil {
		return "", fmt.Errorf("find match result: %w", err)
	}

	if existingMatchResult.CurrentDateInstanceID.Valid {
		// Date instance already exists, return it
		return existingMatchResult.CurrentDateInstanceID.String, nil
	}

	// 1. Get match config for decision window duration
	config, err := l.configStorer.Config(ctx, exec, &QueryFilterMatchConfig{})
	if err != nil {
		return "", fmt.Errorf("get match config: %w", err)
	}

	// Calculate decision_window_end (default 72 hours from now)
	decisionWindowHours := 72
	if config != nil && config.MatchExpirationHours > 0 {
		decisionWindowHours = config.MatchExpirationHours
	}
	decisionWindowEnd := timeNow().Add(time.Duration(decisionWindowHours) * time.Hour)

	// 2. Create date_instance
	dateInstanceID, err := l.dateInstanceInserter.InsertDateInstance(ctx, exec, &InsertDateInstance{
		MatchResultRefID:  matchResultID,
		Status:            string(enums.DateInstanceStatusProposed),
		DecisionWindowEnd: decisionWindowEnd,
	})
	if err != nil {
		return "", fmt.Errorf("insert date instance: %w", err)
	}

	// 3. Log the creation event
	err = l.dateInstanceInserter.InsertDateInstanceLog(ctx, exec, &InsertDateInstanceLog{
		DateInstanceRefID: dateInstanceID,
		EventType:         "created",
		Details:           null.StringFrom("Date instance auto-created on mutual proposal"),
	})
	if err != nil {
		return "", fmt.Errorf("insert date instance log: %w", err)
	}

	// 4. Update match_result with date_instance reference and lifecycle status
	err = l.matchResultUpdater.UpdateMatchForDateInstance(ctx, exec, &UpdateMatchForDateInstance{
		MatchResultID:         matchResultID,
		CurrentDateInstanceID: dateInstanceID,
		MatchLifecycleStatus:  string(enums.MatchLifecycleStatusScheduling),
	})
	if err != nil {
		return "", fmt.Errorf("update match for date instance: %w", err)
	}

	return dateInstanceID, nil
}

// PassMatch sets the user's action to Passed on the match.
func (l *Logic) PassMatch(
	ctx context.Context,
	exec boil.ContextExecutor,
	params *PassMatchParams,
) error {
	updateParams := &UpdateUserMatchAction{
		MatchResultID:    params.MatchResultID,
		UserID:           params.UserID,
		ActionCategoryID: null.StringFrom(string(enums.MatchUserActionPassed)),
	}

	if err := l.userMatchActionsStorer.Update(ctx, exec, updateParams); err != nil {
		return fmt.Errorf("update match action: %w", err)
	}

	return nil
}

// MarkMatchesSeen marks matches as seen by the user using batch update.
func (l *Logic) MarkMatchesSeen(
	ctx context.Context,
	exec boil.ContextExecutor,
	params *MarkSeenParams,
) error {
	if len(params.MatchIDs) == 0 {
		return nil
	}

	if err := l.userMatchActionsStorer.UpdateSeenBatch(ctx, exec, params.UserID, params.MatchIDs); err != nil {
		return fmt.Errorf("batch update seen: %w", err)
	}

	return nil
}

// UnseenMatchCount returns the count of unseen dropped matches for a user.
func (l *Logic) UnseenMatchCount(
	ctx context.Context,
	exec boil.ContextExecutor,
	userID uuid.UUID,
) (int64, error) {
	// Get all unseen matches using the filter
	filter := &QueryFilterUserMatch{
		UserID: userID,
	}
	matches, err := l.userMatchActionsStorer.UserMatches(ctx, exec, filter)
	if err != nil {
		return 0, fmt.Errorf("get user matches: %w", err)
	}

	// Count unseen matches
	var count int64
	for _, m := range matches {
		if !m.SeenAt.Valid {
			count++
		}
	}

	return count, nil
}
