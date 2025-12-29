package matching

import (
	"context"
	"errors"
	"fmt"

	"wingedapp/pgtester/internal/wingedapp/business/sdk"
	economyLib "wingedapp/pgtester/internal/wingedapp/lib/economy"
	matchLib "wingedapp/pgtester/internal/wingedapp/lib/matching"
	schedulingLib "wingedapp/pgtester/internal/wingedapp/lib/scheduling"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"
)

type Business struct {
	transactor        transactor
	transactorAI      transactor
	transactorSBAuth  transactor
	matchGetter       matchGetter
	approvalSetter    approvalSetter
	dropSetter        dropSetter
	expirySetter      expirySetter
	configGetter      configGetter
	configUpdater     configUpdater
	matchSetGetter    matchSetGetter
	ingester          ingester
	matchRunner       matchRunner
	populator         populator
	userMatchGetter   userMatchGetter
	userMatchActioner userMatchActioner
	schedulingLogic   schedulingLogic
	jobEnqueuer       jobEnqueuer
	actionLogger      actionLogger
}

func NewBusiness(
	transactor transactor,
	transactorAI transactor,
	transactorSBAuth transactor,
	matchGetter matchGetter,
	approvalSetter approvalSetter,
	dropSetter dropSetter,
	expirySetter expirySetter,
	configGetter configGetter,
	configUpdater configUpdater,
	matchSetGetter matchSetGetter,
	ingester ingester,
	matchRunner matchRunner,
	populator populator,
	userMatchGetter userMatchGetter,
	userMatchActioner userMatchActioner,
	schedulingLogic schedulingLogic,
	jobEnqueuer jobEnqueuer,
	actionLogger actionLogger,
) (*Business, error) {
	if transactor == nil {
		return nil, errors.New("transactor is required")
	}
	if matchGetter == nil {
		return nil, errors.New("matchGetter is required")
	}
	if approvalSetter == nil {
		return nil, errors.New("approvalSetter is required")
	}
	if dropSetter == nil {
		return nil, errors.New("dropSetter is required")
	}
	if expirySetter == nil {
		return nil, errors.New("expirySetter is required")
	}
	if configGetter == nil {
		return nil, errors.New("configGetter is required")
	}
	if configUpdater == nil {
		return nil, errors.New("configUpdater is required")
	}
	if matchSetGetter == nil {
		return nil, errors.New("matchSetGetter is required")
	}
	if ingester == nil {
		return nil, errors.New("ingester is required")
	}
	if matchRunner == nil {
		return nil, errors.New("matchRunner is required")
	}
	if userMatchGetter == nil {
		return nil, errors.New("userMatchGetter is required")
	}
	if userMatchActioner == nil {
		return nil, errors.New("userMatchActioner is required")
	}
	// transactorAI, transactorSBAuth, populator, and schedulingLogic are optional
	// jobEnqueuer is required for async batch matching
	if jobEnqueuer == nil {
		return nil, errors.New("jobEnqueuer is required")
	}
	if actionLogger == nil {
		return nil, errors.New("actionLogger is required")
	}

	return &Business{
		transactor:        transactor,
		transactorAI:      transactorAI,
		transactorSBAuth:  transactorSBAuth,
		matchGetter:       matchGetter,
		approvalSetter:    approvalSetter,
		dropSetter:        dropSetter,
		expirySetter:      expirySetter,
		configGetter:      configGetter,
		configUpdater:     configUpdater,
		matchSetGetter:    matchSetGetter,
		ingester:          ingester,
		matchRunner:       matchRunner,
		populator:         populator,
		userMatchGetter:   userMatchGetter,
		userMatchActioner: userMatchActioner,
		schedulingLogic:   schedulingLogic,
		jobEnqueuer:       jobEnqueuer,
		actionLogger:      actionLogger,
	}, nil
}

func sampleMatchMikaela() *Match {
	return &Match{
		ID:            uuid.NewString(),
		Name:          "Mikaela",
		Gender:        "female",
		Sexuality:     "straight",
		Age:           32,
		Seen:          true,
		YourStatus:    "Proposed",
		PartnerStatus: "Pending",
		IntroURL:      "https://file-examples.com/storage/feb6199105691b3e89c076a/2017/11/file_example_MP3_700KB.mp3",
		DistanceInKM:  1.5,
		MatchPercent:  0.80,
		Images: []Image{
			{
				URL:     "https://picsum.photos/id/237/200/300",
				OrderNo: 1,
			},
			{
				URL:     "https://picsum.photos/id/237/200/300",
				OrderNo: 2,
			},
			{
				URL:     "https://picsum.photos/seed/picsum/200/300",
				OrderNo: 3,
			},
			{
				URL:     "https://picsum.photos/200/300?grayscale",
				OrderNo: 4,
			},
		},
	}
}

func sampleMatchJohn() *Match {
	return &Match{
		ID:            uuid.NewString(),
		Name:          "John",
		Age:           35,
		Sexuality:     "Straight",
		Gender:        "Female",
		YourStatus:    "Proposed",
		PartnerStatus: "Pending",
		IntroURL:      "https://file-examples.com/storage/feb6199105691b3e89c076a/2017/11/file_example_MP3_700KB.mp3",
		DistanceInKM:  1.5,
		MatchPercent:  0.80,
		Images: []Image{
			{
				URL:     "https://picsum.photos/id/237/200/300",
				OrderNo: 1,
			},
			{
				URL:     "https://picsum.photos/id/237/200/300",
				OrderNo: 2,
			},
			{
				URL:     "https://picsum.photos/seed/picsum/200/300",
				OrderNo: 3,
			},
			{
				URL:     "https://picsum.photos/200/300?grayscale",
				OrderNo: 4,
			},
		},
	}
}

// UserMatches returns the user's dropped matches from their perspective.
// Only returns matches where:
// - User hasn't acted yet (YourAction = "Pending")
// - Both users haven't mutually proposed (MutualProposal = false)
// - Ordered by oldest dropped first (helps users see matches in order received)
func (b *Business) UserMatches(ctx context.Context, userID uuid.UUID) ([]Match, error) {
	exec := b.transactor.DB()
	aiExec := b.transactorAI.DB()
	filter := &matchLib.QueryFilterUserMatch{
		UserID:         userID,
		YourAction:     null.StringFrom("Pending"),      // Only show pending matches (not Proposed/Passed)
		MutualProposal: null.BoolFrom(false),            // Exclude matches where both proposed
		OrderBy:        null.StringFrom("dropped_at"),   // Sort by when match was dropped
		Sort:           null.StringFrom("+"),            // Ascending (oldest first)
	}

	libMatches, err := b.userMatchGetter.UserMatches(ctx, exec, aiExec, filter)
	if err != nil {
		return nil, fmt.Errorf("get user matches: %w", err)
	}

	matches := make([]Match, len(libMatches))
	for i, m := range libMatches {
		matches[i] = toBusinessMatch(m)
	}

	return matches, nil
}

// UserMatchesPaginated returns paginated user matches with filters and sorting.
func (b *Business) UserMatchesPaginated(ctx context.Context, f *UserMatchFilter) (*MatchesPaginated, error) {
	exec := b.transactor.DB()
	aiExec := b.transactorAI.DB()

	libFilter := &matchLib.QueryFilterUserMatch{
		UserID:          f.UserID,
		YourAction:      f.YourStatus,
		PartnerAction:   f.PartnerStatus,
		MutualProposal:  f.MutualOnly,
		Unseen:          null.Bool{}, // Only set if SeenOnly is false
		SeenOnly:        null.Bool{}, // Only set if SeenOnly is true
		IsExpired:       f.IsExpired,
		OrderBy:         f.OrderBy,
		Sort:            f.Sort,
		EnrichAudio:     true,
		EnrichLovestory: true,
	}

	// Apply sane defaults if not explicitly set
	// This prevents mutual proposals (date-instances) from appearing in match feed
	if !f.MutualOnly.Valid {
		libFilter.MutualProposal = null.BoolFrom(false) // Default: exclude mutual proposals
	}
	// Default to Pending status, but NOT when filtering for mutual (mutual implies both proposed)
	if !f.YourStatus.Valid && !f.MutualOnly.Bool {
		libFilter.YourAction = null.StringFrom("Pending") // Default: only pending matches
	}
	if !f.OrderBy.Valid {
		libFilter.OrderBy = null.StringFrom("dropped_at") // Default: sort by drop time
	}
	if !f.Sort.Valid {
		libFilter.Sort = null.StringFrom("+") // Default: ascending (oldest first)
	}

	// Handle SeenOnly filter (mutually exclusive with Unseen)
	if f.SeenOnly.Valid {
		if f.SeenOnly.Bool {
			libFilter.SeenOnly = null.BoolFrom(true)
		} else {
			libFilter.Unseen = null.BoolFrom(true)
		}
	}

	// Pagination
	if f.Pagination != nil {
		if f.Pagination.Page.Valid {
			libFilter.Page = f.Pagination.Page
		}
		if f.Pagination.Rows.Valid {
			libFilter.Rows = f.Pagination.Rows
		}
	}

	result, err := b.userMatchGetter.UserMatchesPaginated(ctx, exec, aiExec, libFilter)
	if err != nil {
		return nil, fmt.Errorf("get user matches paginated: %w", err)
	}

	// Convert to business matches
	matches := make([]Match, len(result.Data))
	for i, m := range result.Data {
		matches[i] = toBusinessMatch(m)
	}

	return &MatchesPaginated{
		Data: matches,
		Pagination: &sdk.Pagination{
			Page:  null.IntFrom(result.Page),
			Rows:  null.IntFrom(result.Rows),
			Total: result.TotalRows,
		},
	}, nil
}

// UserMatch returns a single match for the user by ID.
func (b *Business) UserMatch(ctx context.Context, userID, matchID uuid.UUID) (*Match, error) {
	exec := b.transactor.DB()
	aiExec := b.transactorAI.DB()
	filter := &matchLib.QueryFilterUserMatch{
		UserID:  userID,
		MatchID: null.StringFrom(matchID.String()),
	}

	libMatch, err := b.userMatchGetter.UserMatch(ctx, exec, aiExec, filter)
	if err != nil {
		return nil, fmt.Errorf("get user match: %w", err)
	}

	match := toBusinessMatch(*libMatch)
	return &match, nil
}

// ProposeOnMatch proposes on a match for the user.
// Returns whether the chat was unlocked (mutual proposal).
// Also triggers referral bonus for the referrer if this is the user's first paid action.
func (b *Business) ProposeOnMatch(ctx context.Context, userID, matchID uuid.UUID) (*ProposeResult, error) {
	tx, err := b.transactor.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	params := &matchLib.ProposeMatchParams{
		MatchResultID: matchID,
		UserID:        userID,
	}

	result, err := b.userMatchActioner.ProposeMatch(ctx, tx, params)
	if err != nil {
		return nil, fmt.Errorf("propose match: %w", err)
	}

	// Trigger referral bonus - lib handles all logic (user lookup, idempotency, credit referrer)
	// Returns nil if user not referred or already processed, error only for real failures
	if err := b.actionLogger.CreateActionLog(ctx, tx, &economyLib.InsertActionLog{
		UserID: userID.String(),
		Type:   economyLib.ActionReferralComplete,
	}); err != nil {
		return nil, fmt.Errorf("referral bonus: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &ProposeResult{
		Success:        result.Success,
		MutualProposal: result.MutualProposal,
		DateInstanceID: result.DateInstanceID,
	}, nil
}

// PassOnMatch passes on a match for the user.
func (b *Business) PassOnMatch(ctx context.Context, userID, matchID uuid.UUID) error {
	tx, err := b.transactor.TX()
	if err != nil {
		return fmt.Errorf("tx: %w", err)
	}
	defer b.transactor.Rollback(tx)

	params := &matchLib.PassMatchParams{
		MatchResultID: matchID,
		UserID:        userID,
	}

	if err = b.userMatchActioner.PassMatch(ctx, tx, params); err != nil {
		return fmt.Errorf("pass match: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// UnseenMatchCount returns the count of unseen dropped matches for a user.
func (b *Business) UnseenMatchCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	exec := b.transactor.DB()
	count, err := b.userMatchGetter.UnseenMatchCount(ctx, exec, userID)
	if err != nil {
		return 0, fmt.Errorf("count unseen matches: %w", err)
	}

	return count, nil
}

// toBusinessMatch converts a lib UserMatch to a business Match.
func toBusinessMatch(m matchLib.UserMatch) Match {
	images := make([]Image, len(m.PartnerPhotos))
	for i, p := range m.PartnerPhotos {
		images[i] = Image{URL: p.URL, OrderNo: p.OrderNo}
	}

	age := 0
	if m.PartnerAge.Valid {
		age = m.PartnerAge.Int
	}

	gender := ""
	if m.PartnerGender.Valid {
		gender = m.PartnerGender.String
	}

	sexuality := ""
	if m.PartnerSexuality.Valid {
		sexuality = m.PartnerSexuality.String
	}

	distance := 0.0
	if m.DistanceKM.Valid {
		distance = m.DistanceKM.Float64
	}

	matchPercent := 0.0
	if m.MatchScore.Valid {
		matchPercent = m.MatchScore.Float64
	}

	introURL := ""
	if m.PartnerIntroURL.Valid {
		introURL = m.PartnerIntroURL.String
	}

	lovestoryURL := ""
	if m.LovestoryURL.Valid {
		lovestoryURL = m.LovestoryURL.String
	}

	var expiresAt *string
	if m.ExpiresAt.Valid {
		formatted := m.ExpiresAt.Time.Format("2006-01-02T15:04:05Z07:00")
		expiresAt = &formatted
	}

	var droppedAt *string
	if m.DroppedAt.Valid {
		formatted := m.DroppedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		droppedAt = &formatted
	}

	return Match{
		ID:            m.ID.String(),
		Name:          m.PartnerName.String,
		Age:           age,
		Gender:        gender,
		Sexuality:     sexuality,
		IntroURL:      introURL,
		LovestoryURL:  lovestoryURL,
		Images:        images,
		DistanceInKM:  distance,
		MatchPercent:  matchPercent,
		YourStatus:    m.YourAction,
		PartnerStatus: m.PartnerAction,
		Seen:          m.SeenAt.Valid,
		ExpiresAt:     expiresAt,
		DroppedAt:     droppedAt,
	}
}

// Match returns a dummy match for demonstration purposes (legacy).
// Deprecated: Use UserMatch instead.
func (b *Business) Match() *Match {
	return sampleMatchMikaela()
}

// Matches returns dummy matches (legacy).
// Deprecated: Use UserMatches instead.
func (b *Business) Matches() []*Match {
	return []*Match{sampleMatchMikaela(), sampleMatchJohn()}
}

// ProposeMatch is a legacy stub (deprecated).
// Deprecated: Use ProposeOnMatch instead.
func (b *Business) ProposeMatch() error {
	return nil
}

// PassMatch is a legacy stub (deprecated).
// Deprecated: Use PassOnMatch instead.
func (b *Business) PassMatch() error {
	return nil
}

// MatchOverlaps returns scheduling overlaps between two matched users.
// Validates that the requesting user is part of the match and both have proposed.
func (b *Business) MatchOverlaps(ctx context.Context, userID, matchID uuid.UUID) (*MatchOverlapsResult, error) {
	exec := b.transactor.DB()
	aiExec := b.transactorAI.DB()

	// 1. Fetch match result
	matchResult, err := b.matchGetter.MatchResult(ctx, exec, aiExec, &matchLib.QueryFilterMatchResult{
		ID: null.StringFrom(matchID.String()),
	})
	if err != nil {
		return nil, ErrMatchNotFound
	}

	// 2. Validate user is in match
	userIDA := matchResult.UserAID.String()
	userIDB := matchResult.UserBID.String()
	userIDStr := userID.String()

	if userIDStr != userIDA && userIDStr != userIDB {
		return nil, ErrUserNotInMatch
	}

	// 3. Get the UserMatch to check action status
	// We need to fetch the user's match view to check both actions
	userMatchA, err := b.userMatchGetter.UserMatch(ctx, exec, aiExec, &matchLib.QueryFilterUserMatch{
		UserID:  matchResult.UserAID,
		MatchID: null.StringFrom(matchID.String()),
	})
	if err != nil {
		return nil, fmt.Errorf("get user A match: %w", err)
	}

	userMatchB, err := b.userMatchGetter.UserMatch(ctx, exec, aiExec, &matchLib.QueryFilterUserMatch{
		UserID:  matchResult.UserBID,
		MatchID: null.StringFrom(matchID.String()),
	})
	if err != nil {
		return nil, fmt.Errorf("get user B match: %w", err)
	}

	// 4. Validate both users have PROPOSED status
	// For userMatchA: YourAction is A's action, PartnerAction is B's action
	// For userMatchB: YourAction is B's action, PartnerAction is A's action
	if userMatchA.YourAction != matchLib.MatchUserActionProposed || userMatchB.YourAction != matchLib.MatchUserActionProposed {
		return nil, ErrNotMutuallyProposed
	}

	// 5. Get overlaps via scheduling biz
	overlaps, err := b.schedulingLogic.FindOverlaps(ctx, userIDA, userIDB)
	if err != nil {
		return nil, fmt.Errorf("find overlaps: %w", err)
	}

	if overlaps == nil {
		overlaps = []schedulingLib.TimeBlock{}
	}

	return &MatchOverlapsResult{
		OverlapBlocks:       overlaps,
		TotalOverlapMinutes: schedulingLib.TotalMinutes(overlaps),
	}, nil
}
