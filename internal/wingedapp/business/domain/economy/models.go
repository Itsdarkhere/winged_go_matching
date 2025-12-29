package economy

// CheckinResponse is the response for daily check-in.
type CheckinResponse struct {
	Success            bool   `json:"success"`
	NewStreak          int    `json:"new_streak"`
	IsNewLongestStreak bool   `json:"is_new_longest_streak"`
	MilestoneReached   bool   `json:"milestone_reached"`
	MilestoneType      int    `json:"milestone_type,omitempty"`
	WingsAwarded       int    `json:"wings_awarded,omitempty"`
	AlreadyCheckedIn   bool   `json:"already_checked_in"`
	Message            string `json:"message"`
}

// CheckinStatusResponse is the response for getting check-in status.
type CheckinStatusResponse struct {
	CheckedInToday    bool `json:"checked_in_today"`
	StreakCurrentDays int  `json:"streak_current_days"`
	StreakLongestDays int  `json:"streak_longest_days"`
	NextMilestone     int  `json:"next_milestone"`
	DaysToMilestone   int  `json:"days_to_milestone"`
	MilestoneWings    int  `json:"milestone_wings"`
}

// RevenueCatWebhookRequest is the webhook payload from RevenueCat.
type RevenueCatWebhookRequest struct {
	APIVersion string          `json:"api_version"`
	Event      RevenueCatEvent `json:"event"`
}

// RevenueCatEvent is a single event in the RevenueCat webhook payload.
type RevenueCatEvent struct {
	ID                string   `json:"id"`
	Type              string   `json:"type"`
	EventTimestampMs  int64    `json:"event_timestamp_ms"`
	AppUserID         string   `json:"app_user_id"`
	OriginalAppUserID string   `json:"original_app_user_id"`
	Aliases           []string `json:"aliases"`
	ProductID         string   `json:"product_id"`
	PeriodType        string   `json:"period_type"`
	PurchasedAtMs     int64    `json:"purchased_at_ms"`
	ExpirationAtMs    int64    `json:"expiration_at_ms"`
	Environment       string   `json:"environment"`
	Store             string   `json:"store"`
	CountryCode       string   `json:"country_code"`
	Currency          string   `json:"currency"`
	Price             float64  `json:"price"`
	EntitlementIDs    []string `json:"entitlement_ids"`
	OfferCode         *string  `json:"offer_code"`
	IsFamilyShare     bool     `json:"is_family_share"`
}

// RevenueCatWebhookResponse is the response for the RevenueCat webhook.
type RevenueCatWebhookResponse struct {
	Success bool   `json:"success"`
	EventID string `json:"event_id"`
	Action  string `json:"action"`
	Reason  string `json:"reason,omitempty"`
	Error   string `json:"error,omitempty"`
}
