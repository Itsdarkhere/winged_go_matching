package matching

import (
	"errors"

	"wingedapp/pgtester/internal/wingedapp/business/sdk"
	schedulingLib "wingedapp/pgtester/internal/wingedapp/lib/scheduling"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"
)

// Sentinel errors for match overlap validation
var (
	ErrMatchNotFound       = errors.New("match not found")
	ErrUserNotInMatch      = errors.New("user not part of this match")
	ErrNotMutuallyProposed = errors.New("both users must have proposed")
)

// UserMatchFilter contains filter/sort/pagination options for user matches
type UserMatchFilter struct {
	UserID        uuid.UUID   // Required: the current user
	YourStatus    null.String // Filter by your action: Pending, Proposed, Passed
	PartnerStatus null.String // Filter by partner action
	MutualOnly    null.Bool   // Filter by mutual proposal (both proposed)
	SeenOnly      null.Bool   // Filter seen (true) or unseen (false)
	IsExpired     null.Bool   // Filter expired matches

	// Sorting
	OrderBy null.String // Field: dropped_at, expires_at
	Sort    null.String // + for ASC, - for DESC

	// Pagination
	Pagination *sdk.Pagination
}

// MatchesPaginated wraps a paginated list of matches
type MatchesPaginated struct {
	Data       []Match         `json:"data"`
	Pagination *sdk.Pagination `json:"pagination"`
}

type Match struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Age           int     `json:"age"`
	IntroURL      string  `json:"intro_url"`
	LovestoryURL  string  `json:"lovestory_url"`
	Gender        string  `json:"gender"`
	Sexuality     string  `json:"sexuality"`
	Images        []Image `json:"images"`
	DistanceInKM  float64 `json:"distance_in_kms"`
	MatchPercent  float64 `json:"match_percent"`
	YourStatus    string  `json:"your_status"`    // Your action: Pending, Proposed, Passed
	PartnerStatus string  `json:"partner_status"` // Partner's action: Pending, Proposed, Passed
	Seen          bool    `json:"seen"`
	ExpiresAt     *string `json:"expires_at,omitempty"` // 72h proposal expiration (RFC3339)
	DroppedAt     *string `json:"dropped_at,omitempty"` // When match was delivered (RFC3339)
}

type Image struct {
	URL     string `json:"url"`
	OrderNo int    `json:"order_no"`
}

// ProposeResult contains the result of proposing on a match.
type ProposeResult struct {
	Success        bool   `json:"success"`
	MutualProposal bool   `json:"mutual_proposal"`            // True if both users have proposed
	DateInstanceID string `json:"date_instance_id,omitempty"` // Populated on mutual proposal
}

// MatchOverlapsResult contains the scheduling overlap data for a match.
type MatchOverlapsResult struct {
	OverlapBlocks       []schedulingLib.TimeBlock `json:"overlap_blocks"`
	TotalOverlapMinutes int                       `json:"total_overlap_minutes"`
}
