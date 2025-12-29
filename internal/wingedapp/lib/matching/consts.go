package matching

var (
	male      = "Male"
	female    = "Female"
	nonBinary = "Non-Binary"

	matchingIsHeteroMatch = "is_hetero_match"
	matchingAgeGap        = "age_gap"
)

// Category constants for date instance creation
const (
	// Category types
	DateInstanceStatusCategoryType = "Date Instance Status"
	MatchLifecycleStatusType       = "Match Lifecycle Status"

	// Date Instance Status values
	DateInstanceStatusProposed = "Proposed"

	// Match Lifecycle Status values
	MatchLifecycleStatusConfirmed                   = "Confirmed"
	MatchLifecycleStatusScheduling                  = "Scheduling"
	MatchLifecycleStatusDateSet                     = "Date Set"
	MatchLifecycleStatusDateCompletePendingFeedback = "Date Complete Pending Feedback"
	MatchLifecycleStatusDecisionPendingWindow       = "Decision Pending Window"
	MatchLifecycleStatusQueued                      = "Queued"
	MatchLifecycleStatusClosed                      = "Closed"

	// Match Status values (expiration tracking)
	MatchStatusActive  = "Active"
	MatchStatusExpired = "Expired"
)
