package store

import (
	"context"
	"fmt"
	"strings"
	"time"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"
)

// UserMatchActionsStore handles user-facing match data access (CRUD only).
type UserMatchActionsStore struct {
	l applog.Logger
}

// NewUserMatchActionsStore creates a new UserMatchActionsStore.
func NewUserMatchActionsStore(l applog.Logger) *UserMatchActionsStore {
	return &UserMatchActionsStore{l: l}
}

// buildUserMatchBaseQuery builds the base WHERE conditions for user match queries.
// This is shared between the data query and count query.
func (s *UserMatchActionsStore) buildUserMatchBaseQuery(f *matching.QueryFilterUserMatch) []qm.QueryMod {
	matchResultTbl := pgmodel.TableNames.MatchResult
	usersTbl := pgmodel.TableNames.Users

	qMods := []qm.QueryMod{
		qm.From(matchResultTbl + " mr"),
		qm.InnerJoin(usersTbl + " ua ON ua.id = mr.initiator_user_ref_id"),
		qm.InnerJoin(usersTbl + " ub ON ub.id = mr.receiver_user_ref_id"),
		// Base filters: user must be in match + visibility conditions
		qm.Where("(mr.initiator_user_ref_id = ? OR mr.receiver_user_ref_id = ?)", f.UserID.String(), f.UserID.String()),
		qm.Where("mr.is_dropped = true"),
		qm.Where("mr.is_approved = true"),
		qm.Where("mr.is_expired = false"),
	}

	// Optional filters
	if f.MatchID.Valid {
		qMods = append(qMods, qm.Where("mr.id = ?", f.MatchID.String))
	}

	if f.YourAction.Valid {
		qMods = append(qMods, qm.Where(
			fmt.Sprintf(`CASE WHEN mr.initiator_user_ref_id = '%s' THEN mr.initiator_action ELSE mr.receiver_action END = ?`, f.UserID.String()),
			f.YourAction.String,
		))
	}

	if f.PartnerAction.Valid {
		qMods = append(qMods, qm.Where(
			fmt.Sprintf(`CASE WHEN mr.initiator_user_ref_id = '%s' THEN mr.receiver_action ELSE mr.initiator_action END = ?`, f.UserID.String()),
			f.PartnerAction.String,
		))
	}

	if f.MutualProposal.Valid {
		if f.MutualProposal.Bool {
			// Mutual proposal = both users have proposed (regardless of date instance)
			qMods = append(qMods, qm.Where("mr.initiator_action = 'Proposed' AND mr.receiver_action = 'Proposed'"))
		} else {
			// Not mutual = at least one user hasn't proposed yet
			qMods = append(qMods, qm.Where("mr.initiator_action != 'Proposed' OR mr.receiver_action != 'Proposed'"))
		}
	}

	if f.IsExpired.Valid {
		qMods = append(qMods, qm.Where("mr.is_expired = ?", f.IsExpired.Bool))
	}

	// Unseen filter (seen_at IS NULL)
	if f.Unseen.Valid && f.Unseen.Bool {
		qMods = append(qMods, qm.Where(
			fmt.Sprintf(`CASE WHEN mr.initiator_user_ref_id = '%s' THEN mr.initiator_seen_at IS NULL ELSE mr.receiver_seen_at IS NULL END`, f.UserID.String()),
		))
	}

	// SeenOnly filter (seen_at IS NOT NULL)
	if f.SeenOnly.Valid && f.SeenOnly.Bool {
		qMods = append(qMods, qm.Where(
			fmt.Sprintf(`CASE WHEN mr.initiator_user_ref_id = '%s' THEN mr.initiator_seen_at IS NOT NULL ELSE mr.receiver_seen_at IS NOT NULL END`, f.UserID.String()),
		))
	}

	return qMods
}

// UserMatchesCount returns the total count of matches for pagination.
func (s *UserMatchActionsStore) UserMatchesCount(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterUserMatch,
) (int, error) {
	qMods := s.buildUserMatchBaseQuery(f)
	qMods = append([]qm.QueryMod{qm.Select("COUNT(*)")}, qMods...)

	var count int
	if err := pgmodel.NewQuery(qMods...).QueryRow(exec).Scan(&count); err != nil {
		return 0, fmt.Errorf("count user matches: %w", err)
	}

	return count, nil
}

// UserMatches returns dropped matches for a user from their perspective.
// The query determines whether the user is user_a or user_b and returns
// the appropriate partner info and action statuses.
func (s *UserMatchActionsStore) UserMatches(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterUserMatch,
) ([]matching.UserMatch, error) {
	var results []matching.UserMatch

	// Build select columns
	selectCols := []string{
		"mr.id AS id",
		fmt.Sprintf(`CASE WHEN mr.initiator_user_ref_id = '%s' THEN mr.receiver_user_ref_id ELSE mr.initiator_user_ref_id END AS partner_id`, f.UserID.String()),
		fmt.Sprintf(`CASE WHEN mr.initiator_user_ref_id = '%s' THEN ub.first_name ELSE ua.first_name END AS partner_name`, f.UserID.String()),
		fmt.Sprintf(`CASE WHEN mr.initiator_user_ref_id = '%s' THEN EXTRACT(YEAR FROM AGE(ub.birthday))::INTEGER ELSE EXTRACT(YEAR FROM AGE(ua.birthday))::INTEGER END AS partner_age`, f.UserID.String()),
		fmt.Sprintf(`CASE WHEN mr.initiator_user_ref_id = '%s' THEN ub.gender ELSE ua.gender END AS partner_gender`, f.UserID.String()),
		fmt.Sprintf(`CASE WHEN mr.initiator_user_ref_id = '%s' THEN ub.sexuality ELSE ua.sexuality END AS partner_sexuality`, f.UserID.String()),
		fmt.Sprintf(`CASE WHEN mr.initiator_user_ref_id = '%s' THEN mr.initiator_action ELSE mr.receiver_action END AS your_action`, f.UserID.String()),
		fmt.Sprintf(`CASE WHEN mr.initiator_user_ref_id = '%s' THEN mr.receiver_action ELSE mr.initiator_action END AS partner_action`, f.UserID.String()),
		"(mr.initiator_action = 'Proposed' AND mr.receiver_action = 'Proposed') AS mutual_proposal",
		"mr.current_date_instance_id AS date_instance_id",
		fmt.Sprintf(`CASE WHEN mr.initiator_user_ref_id = '%s' THEN mr.initiator_seen_at ELSE mr.receiver_seen_at END AS seen_at`, f.UserID.String()),
		"mr.expires_at AS expires_at",
		"mr.dropped_ts AS dropped_at",
		fmt.Sprintf(`CASE WHEN mr.initiator_user_ref_id = '%s' THEN ua.supabase_id ELSE ub.supabase_id END AS your_supabase_id`, f.UserID.String()),
		fmt.Sprintf(`CASE WHEN mr.initiator_user_ref_id = '%s' THEN ub.supabase_id ELSE ua.supabase_id END AS partner_supabase_id`, f.UserID.String()),
	}

	// Start with select, then add base query
	qMods := []qm.QueryMod{qm.Select(selectCols...)}
	qMods = append(qMods, s.buildUserMatchBaseQuery(f)...)

	// Ordering - default ASC (oldest first, expires soonest)
	orderDir := "ASC"
	if f.Sort.Valid && f.Sort.String == "-" {
		orderDir = "DESC"
	}

	if f.OrderBy.Valid {
		// Only allow sorting by actual DB columns
		allowedColumns := map[string]string{
			"dropped_at": "mr.dropped_ts",
			"expires_at": "mr.expires_at",
		}
		if col, ok := allowedColumns[f.OrderBy.String]; ok {
			qMods = append(qMods, qm.OrderBy(col+" "+orderDir))
		} else {
			qMods = append(qMods, qm.OrderBy("mr.dropped_ts ASC"))
		}
	} else {
		// Default: oldest first (dropped_ts ASC) - first to expire shown first
		qMods = append(qMods, qm.OrderBy("mr.dropped_ts ASC"))
	}

	// Pagination
	if f.Rows.Valid && f.Rows.Int > 0 {
		qMods = append(qMods, qm.Limit(f.Rows.Int))
		if f.Page.Valid && f.Page.Int > 1 {
			qMods = append(qMods, qm.Offset((f.Page.Int-1)*f.Rows.Int))
		}
	}

	if err := pgmodel.NewQuery(qMods...).Bind(ctx, exec, &results); err != nil {
		return nil, fmt.Errorf("query user matches: %w", err)
	}

	return results, nil
}

// UserMatchesPaginated returns user matches with pagination info.
func (s *UserMatchActionsStore) UserMatchesPaginated(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterUserMatch,
) (*matching.UserMatchPaginated, error) {
	// Get total count
	totalRows, err := s.UserMatchesCount(ctx, exec, f)
	if err != nil {
		return nil, err
	}

	// Get data
	matches, err := s.UserMatches(ctx, exec, f)
	if err != nil {
		return nil, err
	}

	// Log diagnosis when no results found
	if len(matches) == 0 && s.l != nil {
		s.logNoMatchDiagnosis(ctx, exec, f)
	}

	// Calculate pagination
	page := 1
	if f.Page.Valid && f.Page.Int > 0 {
		page = f.Page.Int
	}
	rows := 50 // default
	if f.Rows.Valid && f.Rows.Int > 0 {
		rows = f.Rows.Int
	}
	totalPages := (totalRows + rows - 1) / rows
	if totalPages == 0 {
		totalPages = 1
	}

	return &matching.UserMatchPaginated{
		Data:       matches,
		Page:       page,
		Rows:       rows,
		TotalRows:  totalRows,
		TotalPages: totalPages,
	}, nil
}

// UserMatch returns a single match for a user by ID.
func (s *UserMatchActionsStore) UserMatch(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterUserMatch,
) (*matching.UserMatch, error) {
	matches, err := s.UserMatches(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("user matches: %w", err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no match found for user %s with filter", f.UserID)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("expected 1 match, got %d", len(matches))
	}

	return &matches[0], nil
}

// Update updates a user's action or seen status on a match.
// This is a pure CRUD operation - business logic belongs in the Logic layer.
func (s *UserMatchActionsStore) Update(
	ctx context.Context,
	exec boil.ContextExecutor,
	params *matching.UpdateUserMatchAction,
) error {
	matchResult, err := pgmodel.FindMatchResult(ctx, exec, params.MatchResultID.String())
	if err != nil {
		return fmt.Errorf("find match result: %w", err)
	}

	// Determine if user is A or B
	isUserA := matchResult.InitiatorUserRefID == params.UserID.String()
	isUserB := matchResult.ReceiverUserRefID == params.UserID.String()
	if !isUserA && !isUserB {
		return fmt.Errorf("user %s is not part of match %s", params.UserID, params.MatchResultID)
	}

	// Update action category if provided
	if params.ActionCategoryID.Valid {
		if isUserA {
			matchResult.InitiatorAction = params.ActionCategoryID.String
		} else {
			matchResult.ReceiverAction = params.ActionCategoryID.String
		}
	}

	// Update seen_at if provided
	if params.SeenAt.Valid {
		if isUserA {
			matchResult.InitiatorSeenAt = params.SeenAt
		} else {
			matchResult.ReceiverSeenAt = params.SeenAt
		}
	}

	if _, err := matchResult.Update(ctx, exec, boil.Infer()); err != nil {
		return fmt.Errorf("update match result: %w", err)
	}

	return nil
}

// UpdateSeenBatch marks multiple matches as seen for a user using a single SQL query.
// This is more efficient than updating one at a time.
func (s *UserMatchActionsStore) UpdateSeenBatch(
	ctx context.Context,
	exec boil.ContextExecutor,
	userID uuid.UUID,
	matchIDs []uuid.UUID,
) error {
	if len(matchIDs) == 0 {
		return nil
	}

	// Build the IN clause
	placeholders := make([]string, len(matchIDs))
	args := make([]interface{}, 0, len(matchIDs)+2)

	for i, id := range matchIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args = append(args, id.String())
	}

	now := time.Now()
	userIDStr := userID.String()

	// Update initiator_seen_at for matches where user is user_a
	queryA := fmt.Sprintf(`
		UPDATE match_result
		SET initiator_seen_at = $%d
		WHERE id IN (%s)
		  AND initiator_user_ref_id = $%d
		  AND initiator_seen_at IS NULL`,
		len(matchIDs)+1,
		strings.Join(placeholders, ", "),
		len(matchIDs)+2,
	)
	argsA := append(args, now, userIDStr)

	if _, err := exec.ExecContext(ctx, queryA, argsA...); err != nil {
		return fmt.Errorf("update initiator_seen_at batch: %w", err)
	}

	// Update receiver_seen_at for matches where user is user_b
	queryB := fmt.Sprintf(`
		UPDATE match_result
		SET receiver_seen_at = $%d
		WHERE id IN (%s)
		  AND receiver_user_ref_id = $%d
		  AND receiver_seen_at IS NULL`,
		len(matchIDs)+1,
		strings.Join(placeholders, ", "),
		len(matchIDs)+2,
	)
	argsB := append(args, now, userIDStr)

	if _, err := exec.ExecContext(ctx, queryB, argsB...); err != nil {
		return fmt.Errorf("update receiver_seen_at batch: %w", err)
	}

	return nil
}

// looseCandidateMatch holds raw match_result data for diagnosis.
type looseCandidateMatch struct {
	ID          string
	IsDropped   bool
	IsApproved  bool
	IsExpired   bool
	InitiatorAction string
	ReceiverAction string
	InitiatorUserRefID  string
	ReceiverUserRefID  string
}

// logNoMatchDiagnosis runs a looser query and logs why filters excluded results.
func (s *UserMatchActionsStore) logNoMatchDiagnosis(
	ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterUserMatch,
) {
	// Build input filters map for logging
	inputFilters := map[string]any{
		"user_id": f.UserID.String(),
	}
	if f.MatchID.Valid {
		inputFilters["match_id"] = f.MatchID.String
	}
	if f.YourAction.Valid {
		inputFilters["your_action"] = f.YourAction.String
	}
	if f.PartnerAction.Valid {
		inputFilters["partner_action"] = f.PartnerAction.String
	}
	if f.MutualProposal.Valid {
		inputFilters["mutual_proposal"] = f.MutualProposal.Bool
	}
	if f.IsExpired.Valid {
		inputFilters["is_expired"] = f.IsExpired.Bool
	}
	if f.Unseen.Valid {
		inputFilters["unseen"] = f.Unseen.Bool
	}
	if f.SeenOnly.Valid {
		inputFilters["seen_only"] = f.SeenOnly.Bool
	}

	s.l.Info(ctx, "no-match-diagnosis: checking why no results", applog.F("input_filters", inputFilters))

	// Run looser query - just user_a OR user_b, no other filters
	qMods := []qm.QueryMod{
		qm.Select(
			"mr.id",
			"mr.is_dropped",
			"mr.is_approved",
			"mr.is_expired",
			"mr.initiator_action",
			"mr.receiver_action",
			"mr.initiator_user_ref_id",
			"mr.receiver_user_ref_id",
		),
		qm.From(pgmodel.TableNames.MatchResult + " mr"),
		qm.Where("(mr.initiator_user_ref_id = ? OR mr.receiver_user_ref_id = ?)", f.UserID.String(), f.UserID.String()),
	}

	var candidates []looseCandidateMatch
	if err := pgmodel.NewQuery(qMods...).Bind(ctx, exec, &candidates); err != nil {
		s.l.Warn(ctx, "no-match-diagnosis: failed to run loose query", applog.F("error", err.Error()))
		return
	}

	if len(candidates) == 0 {
		s.l.Info(ctx, "no-match-diagnosis: user has NO match_result rows at all (neither user_a nor user_b)")
		return
	}

	s.l.Info(ctx, "no-match-diagnosis: found candidates with loose query", applog.F("loose_count", len(candidates)))

	// Check each candidate against filters
	for _, c := range candidates {
		failedFilters := []string{}

		// Hard-coded filters in buildUserMatchBaseQuery
		if !c.IsDropped {
			failedFilters = append(failedFilters, "is_dropped=false (needs true)")
		}
		if !c.IsApproved {
			failedFilters = append(failedFilters, "is_approved=false (needs true)")
		}
		if c.IsExpired {
			failedFilters = append(failedFilters, "is_expired=true (needs false)")
		}

		// Determine if user is A or B
		isUserA := c.InitiatorUserRefID == f.UserID.String()
		yourAction := c.InitiatorAction
		partnerAction := c.ReceiverAction
		if !isUserA {
			yourAction = c.ReceiverAction
			partnerAction = c.InitiatorAction
		}

		// Optional filters
		if f.YourAction.Valid && yourAction != f.YourAction.String {
			failedFilters = append(failedFilters, fmt.Sprintf("your_action=%s (needs %s)", yourAction, f.YourAction.String))
		}
		if f.PartnerAction.Valid && partnerAction != f.PartnerAction.String {
			failedFilters = append(failedFilters, fmt.Sprintf("partner_action=%s (needs %s)", partnerAction, f.PartnerAction.String))
		}
		if f.MutualProposal.Valid {
			isMutual := c.InitiatorAction == "Proposed" && c.ReceiverAction == "Proposed"
			if f.MutualProposal.Bool && !isMutual {
				failedFilters = append(failedFilters, fmt.Sprintf("mutual_proposal=false (user_a=%s, user_b=%s)", c.InitiatorAction, c.ReceiverAction))
			} else if !f.MutualProposal.Bool && isMutual {
				failedFilters = append(failedFilters, "mutual_proposal=true (filter wants false)")
			}
		}
		if f.IsExpired.Valid && c.IsExpired != f.IsExpired.Bool {
			failedFilters = append(failedFilters, fmt.Sprintf("is_expired=%v (needs %v)", c.IsExpired, f.IsExpired.Bool))
		}

		if len(failedFilters) > 0 {
			s.l.Info(ctx, "no-match-diagnosis: candidate excluded",
				applog.F("match_id", c.ID),
				applog.F("is_dropped", c.IsDropped),
				applog.F("is_approved", c.IsApproved),
				applog.F("is_expired", c.IsExpired),
				applog.F("initiator_action", c.InitiatorAction),
				applog.F("receiver_action", c.ReceiverAction),
				applog.F("failed_filters", failedFilters),
			)
		}
	}
}
