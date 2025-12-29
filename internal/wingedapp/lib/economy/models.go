package economy

import (
	"fmt"
	"time"
	"wingedapp/pgtester/internal/util/validationlib"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"
)

// InsertActionLog represents an entry to be inserted into the accounting system.
type InsertActionLog struct {
	UserID      string     `json:"user_id" validate:"required"`
	RefID       string     `json:"ref_id"` // Optional for some actions (e.g., ActionReferralComplete looks it up)
	Type        ActionType `json:"category" validate:"required"`
	JSONDetails null.JSON  `json:"json_details"`
}

// QueryFilterActionLog represents filters for querying action logs.
type QueryFilterActionLog struct {
	ID       null.String
	Category null.String
	UserID   null.String
	RefID    null.String
	IsActive null.Int
}

type QueryFilterTransactions struct {
	ID          null.String
	ActionLogID null.String
	IsActive    null.Int
}

// actionsRequiringRefID lists action types that require RefID at validation time.
// Actions not in this list may have RefID looked up by the handler.
var actionsRequiringRefID = map[ActionType]bool{
	ActionWingedXWeeklyPayment:        true,
	ActionWingedXMonthlyPayment:       true,
	ActionWingedPlusWeeklyPayment:     true,
	ActionWingedPlusMonthlyPayment:    true,
	ActionWingedPlusThreeMonthPayment: true,
	ActionWingedPlusSixMonthPayment:   true,
	ActionAttendDate:                  true,
	ActionSendMessage:                 true,
	// ActionReferralComplete: false - RefID is looked up by processReferralBonus
}

func (i *InsertActionLog) validateParams() error {
	if err := validationlib.Validate(i); err != nil {
		return err
	}
	// Check RefID requirement based on action type
	if actionsRequiringRefID[i.Type] && i.RefID == "" {
		return fmt.Errorf("ref_id is required for action type %s", i.Type)
	}
	return nil
}

type ActionLog struct {
	ID          string     `json:"id" boil:"id"`
	UserID      string     `json:"user_id" boil:"user_id"`
	RefID       string     `json:"ref_id" boil:"ref_id"`
	Type        ActionType `json:"type" boil:"type"`
	JSONDetails null.JSON  `json:"json_details" boil:"json_details"` // map for resp readability
	IsActive    bool       `json:"is_active" boil:"is_active"`
}

type Category struct {
	ID   string
	Name string
}

type CanPerformActionParams struct {
	UserID     string
	ActionType ActionType
}

type UserTotals struct {
	ID                string    `boil:"id"`
	Wings             int       `boil:"wings"`
	SentMessages      int       `boil:"sent_messages"`
	PremiumExpiresIn  null.Time `boil:"premium_expires_in"`
	StreakLastDate    null.Time `boil:"streak_last_date"`
	StreakCurrentDays int       `boil:"streak_current_days"`
	StreakLongestDays int       `boil:"streak_longest_days"`
}

type UpdateUserTotals struct {
	ID                string
	PremiumExpiresIn  null.Time
	Wings             null.Int
	SentMessages      null.Int
	StreakLastDate    null.Time
	StreakCurrentDays null.Int
	StreakLongestDays null.Int
}

type SubscriptionPlan struct {
	ID    string        `boil:"id"`
	Name  string        `boil:"name"`
	Price types.Decimal `boil:"price"`
	Wings int           `boil:"wings"`
}

type InsertTransaction struct {
	UserID       string
	ActionTypeID string    // category type of action
	ActionRefID  string    // duplicate in ActionLog.RefID (consistency)
	WingsAmount  int       // wings amount
	IsCredit     bool      // credit or debit
	Claimed      bool      // if already claimed
	ExpiresAt    null.Time // nullable: NULL = never expires, set for earned wings
	ExtraInfo    null.JSON // extra meta-info
}

type Transaction struct {
	ID          string `boil:"id"`
	UserID      string `boil:"user_id"`
	ActionLogID string `boil:"action_ref_id"`
	Amount      int    `boil:"amount"`
	IsCredit    bool   `boil:"is_credit"`
	IsActive    bool   `boil:"is_active"`
}

type SubscriptionPayment struct {
	Type string
	Name string
}

// CheckinTransaction represents a daily check-in transaction
type CheckinTransaction struct {
	ID          string    `boil:"id" json:"id"`
	UserID      string    `boil:"user_id" json:"user_id"`
	Amount      int       `boil:"amount" json:"amount"`
	Claimed     bool      `boil:"claimed" json:"claimed"`
	CreatedDate time.Time `boil:"created_date" json:"created_date"`
}

// QueryFilterCheckinTransaction for filtering check-in transactions
type QueryFilterCheckinTransaction struct {
	UserID        null.String
	ActionLogType null.String // filter by action log type
	IsActive      null.Int    // filter by is_active
	Claimed       null.Bool
	OrderBy       null.String // column to order by: created_date
	Sort          null.String // "+" for ASC, "-" for DESC
	Limit         null.Int    // limit results
}

// UpdateCheckinTransaction for updating check-in transactions
type UpdateCheckinTransaction struct {
	IDs     []string  // transaction IDs to update
	Claimed null.Bool // set claimed status
}

// InsertCheckinTransaction for inserting new check-in transaction
type InsertCheckinTransaction struct {
	UserID      string
	ActionLogID string
	Amount      int
	ExpiresAt   null.Time // nullable: set to 30 days for earned wings
}

// CheckinStatus for API response - streak-based per spec
type CheckinStatus struct {
	CheckedInToday    bool `json:"checked_in_today"`
	StreakCurrentDays int  `json:"streak_current_days"`
	StreakLongestDays int  `json:"streak_longest_days"`
	NextMilestone     int  `json:"next_milestone"`    // 7 or 30
	DaysToMilestone   int  `json:"days_to_milestone"` // days until next milestone
	MilestoneWings    int  `json:"milestone_wings"`   // wings at next milestone
}

// CheckinResult returned after performing check-in
type CheckinResult struct {
	NewStreak         int  `json:"new_streak"`
	MilestoneReached  bool `json:"milestone_reached"`
	MilestoneType     int  `json:"milestone_type,omitempty"` // 7 or 30 if reached
	WingsAwarded      int  `json:"wings_awarded"`
	IsNewLongestStreak bool `json:"is_new_longest_streak"`
}

// User represents a user for economy operations
type User struct {
	ID                  string      `boil:"id"`
	UserInviteCodeRefID null.String `boil:"user_invite_code_ref_id"`
}

// QueryFilterUser for filtering users in economy operations
type QueryFilterUser struct {
	ID         null.String
	Sha256Hash null.String
}

// InviteCode represents an invite code for referral tracking
type InviteCode struct {
	ID                 string      `boil:"id"`
	ReferrerNumberHash null.String `boil:"referrer_number_hash"`
}

// QueryFilterInviteCode for filtering invite codes in economy operations
type QueryFilterInviteCode struct {
	ID null.String
}

// QueryFilterExpiryTransaction for filtering transactions in expiry operations
type QueryFilterExpiryTransaction struct {
	ExpiresAtBefore  null.Time // filter: expires_at < value
	ExpiresAtNotNull null.Bool // filter: expires_at IS NOT NULL
	IsExpired        null.Bool // filter by is_expired
	IsActive         null.Int  // filter by is_active
	IsCredit         null.Bool // filter by is_credit
	GroupByUser      null.Bool // if true, group by user_ref_id with SUM(amount)
}

// UpdateExpiryTransaction for updating transactions in expiry operations
type UpdateExpiryTransaction struct {
	IsExpired null.Bool // set is_expired status
}

// ExpiredUserAmount represents the total expired amount for a user.
type ExpiredUserAmount struct {
	UserID string `boil:"user_id"`
	Amount int    `boil:"amount"`
}

// ProductIDToActionType maps RevenueCat product IDs to our action types.
var ProductIDToActionType = map[string]ActionType{
	"com.app.wingedplus_weekly":  ActionWingedPlusWeeklyPayment,
	"com.app.wingedplus_monthly": ActionWingedPlusMonthlyPayment,
	"com.app.wingedplus_3month":  ActionWingedPlusThreeMonthPayment,
	"com.app.wingedplus_6month":  ActionWingedPlusSixMonthPayment,
	// WingedX products
	"com.app.wingedx_weekly":  ActionWingedXWeeklyPayment,
	"com.app.wingedx_monthly": ActionWingedXMonthlyPayment,
}
