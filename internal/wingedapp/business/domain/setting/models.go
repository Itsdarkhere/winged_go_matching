package setting

import (
	"fmt"
	"time"
	"wingedapp/pgtester/internal/util/errutil"
	"wingedapp/pgtester/internal/wingedapp/business/sdk"

	"github.com/aarondl/null/v8"
)

type InviteUser struct {
	ReferrerNumber string
	ExtNumber      string
	InviteCode     string
}

type User struct {
	ID string `json:"id"`
}

type UserInviteCode struct {
	ID             string    `json:"id"`
	Code           string    `json:"code"`
	Category       string    `json:"category"`
	ForNumber      string    `json:"for_number"`
	ForNumberHash  string    `json:"for_number_hash"`
	UsageCount     int       `json:"usage_count"`
	ReferralSource string    `json:"referral_source"`
	CreatedAt      time.Time `json:"created_at"`
	LastUsed       null.Time `json:"last_used"`
}

type AnonymizedContacts struct {
	Data       []AnonymizedContact `json:"data"`
	Pagination *sdk.Pagination     `json:"pagination"`
}

type UpsertAnonymizedContact struct {
	OwnerHash   string `json:"owner_hash" validate:"sha256"`
	ContactHash string `json:"contact_hash" validate:"sha256"`
}

type AnonymizedContact struct {
	ContactHash     string `json:"contact_hash"`
	FriendsOnWinged int    `json:"friends_on_winged"`
	Status          string `json:"status"`
}

// QueryFilterAnonymizedContact represents the filters for querying anonymized contacts.
type QueryFilterAnonymizedContact struct {
	ContactHashes []string
	OrderBy       null.String `json:"order_by"`
	Sort          null.String `json:"sort"`
	Pagination    *sdk.Pagination
}

type SysParam struct {
	Key string
	Val string
}

// Category is a lightweight category reference (id + name).
type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PersonalInformation represents user's personal information.
type PersonalInformation struct {
	ID                            string    `json:"id"`
	Email                         string    `json:"email"`
	Name                          string    `json:"name"`
	Birthday                      time.Time `json:"birthday"`
	Number                        string    `json:"number"`
	Height                        int       `json:"height"`
	Gender                        string    `json:"gender"`
	AgentDating                   bool      `json:"agent_dating"`
	DatingPreferences             []string  `json:"dating_preferences"`
	DatingPreferenceAgeRangeStart int       `json:"dating_preference_age_range_start"`
	DatingPreferenceAgeRangeEnd   int       `json:"dating_preference_age_range_end"`
	Location                      string    `json:"location"`

	DietaryRestrictions []Category `json:"dietary_restrictions"`
	DateTypePreferences []Category `json:"date_type_preferences"`
	MobilityConstraints []Category `json:"mobility_constraints"`
}

// UpdatePersonalInformation represents the data required to update user's personal information.
type UpdatePersonalInformation struct {
	UserID                        string      `json:"user_id"`
	Name                          null.String `json:"name"`
	Birthday                      null.Time   `json:"birthday"`
	Number                        null.String `json:"number"`
	Height                        null.Int    `json:"height"`
	AgentDating                   null.Bool   `json:"agent_dating"`
	DatingPreferences             []string    `json:"dating_preferences"`
	DatingPreferenceAgeRangeStart null.Int    `json:"dating_preference_age_range_start"`
	DatingPreferenceAgeRangeEnd   null.Int    `json:"dating_preference_age_range_end"`
	Location                      null.String `json:"location"`

	DietaryRestrictionIDs []string `json:"dietary_restriction_ids"`
	DateTypePreferenceIDs []string `json:"date_type_preference_ids"`
	MobilityConstraintIDs []string `json:"mobility_constraint_ids"`

	updateDatingPrefs         bool // internal use only
	updateDietaryRestrictions bool // internal use only
	updateDateTypePreferences bool // internal use only
	updateMobilityConstraints bool // internal use only
}

const (
	datingPrefMale      = "Male"
	datingPrefFemale    = "Female"
	datingPrefNonBinary = "Non-Binary"
)

// validateDatingPrefs checks if the provided dating preferences are valid.
func validateDatingPrefs(datePrefs []string) errutil.List {
	errs := errutil.List{}

	validDatingPreferences := map[string]bool{
		datingPrefMale:      true,
		datingPrefFemale:    true,
		datingPrefNonBinary: true,
	}
	for _, pref := range datePrefs {
		if _, ok := validDatingPreferences[pref]; !ok {
			errs.Add(fmt.Sprintf("invalid dating preference: %s", pref))
			continue
		}
	}

	return errs
}

type PersonalInformationQueryFilter struct {
	ID string
}

type QueryFilterUser struct {
	ID         null.String `json:"id"`
	Number     null.String `json:"mobile_number"`
	MobileCode null.String `json:"mobile_code"`
	Email      null.String `json:"email"`

	EnrichPhotos     bool `json:"enrich_photos"`      // add user photos
	EnrichCallStates bool `json:"enrich_call_states"` // add call states
}

type QueryFilterUserInviteCode struct {
	InviteCode null.String `json:"invite_code"`
	Number     null.String `json:"number"`

	Pagination *sdk.Pagination
	OrderBy    null.String `json:"order_by"`
	Sort       null.String `json:"sort"`
}

type InsertUserInviteCode struct {
	InviteCode         string `json:"code"`
	ExtNumber          string `json:"ext_number"`
	ExtNumberHash      string `json:"ext_number_hash"`
	ReferrerNumberHash string `json:"referrer_number_hash"`
}
