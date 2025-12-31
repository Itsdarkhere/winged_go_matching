package matching

import (
	"context"
	"encoding/json"
	"time"
	"wingedapp/pgtester/internal/util/errutil"
	"wingedapp/pgtester/internal/util/validationlib"
	"wingedapp/pgtester/internal/wingedapp/business/sdk"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/google/uuid"
)

type UserDatingPreference struct {
	ID               string    `json:"id" boil:"id"`
	UserID           uuid.UUID `json:"user_id" boil:"user_id"`
	DatingPreference string    `json:"dating_preference" boil:"dating_preference"`
}

// User the userByID model, that contains matching parameters.
type User struct {
	ID                uuid.UUID              `boil:"id" json:"id" validate:"required"`
	Email             string                 `boil:"email" json:"email"`
	Age               null.Int               `boil:"age" json:"age"`
	DatingPreferences []UserDatingPreference `boil:"dating_preferences" json:"dating_preferences,omitempty"`
	Gender            null.String            `boil:"gender" json:"gender"`
	Height            null.Float64           `boil:"height" json:"height"`
	Latitude          null.Float64           `boil:"latitude" json:"latitude"`
	Longitude         null.Float64           `boil:"longitude" json:"longitude"`
	FirstName         null.String            `boil:"firstname" json:"firstname"`
	LastName          null.String            `boil:"lastname" json:"lastname"`
	UserType          null.String            `boil:"user_type" json:"user_type"`       // Admin, Regular User
	IsTestUser        null.Bool              `boil:"is_test_user" json:"is_test_user"` // Test user flag
}

func (user *User) Validate() error {
	return validationlib.Validate(user)
}

const (
	ageWindowQualifier       QualifierType = "age_window_qualifier"
	datePrefsQualifier       QualifierType = "date_prefs_qualifier"
	heightQualifier          QualifierType = "height_qualifier"
	distanceQualifier        QualifierType = "distance_qualifier"
	qualitativeQualifierName QualifierType = "qualitative_qualifier"
)

// getMaleUser returns the male user
func getMaleUser(userA *User, userB *User) (*User, error) {
	if userA.Gender.Valid && userA.Gender.String == male {
		return userA, nil
	}
	if userB.Gender.Valid && userB.Gender.String == male {
		return userB, nil
	}

	return nil, ErrNoMale
}

// getFemaleUser returns the female user
func getFemaleUser(userA *User, userB *User) (*User, error) {
	if userA.Gender.Valid && userA.Gender.String == female {
		return userA, nil
	}
	if userB.Gender.Valid && userB.Gender.String == female {
		return userB, nil
	}

	return nil, ErrNoFemale
}

// UserPair is a set of users to be matched
type UserPair struct {
	UserA *User
	UserB *User
}

func newQualifier(qt QualifierType) *Qualifier {
	return &Qualifier{
		Name:      qt,
		Telemetry: make(map[string]any),
	}
}

// UserPairs is the current set of users.
type UserPairs []UserPair

// UserPairsUniqPerm returns a full permutation of userByID sets
// from users.
func UserPairsUniqPerm(users []User) UserPairs {
	// build all possible pairs
	uSets := make(UserPairs, 0)
	for i := 0; i < len(users); i++ {
		for j := i + 1; j < len(users); j++ {
			// If uuids not equal then push a new one.
			if users[i].ID != users[j].ID {
				uSets = append(uSets, UserPair{
					UserA: &users[i],
					UserB: &users[j],
				})
			}
		}
	}

	return uSets.dedupedPermutation()
}

// dedupedPermutation returns sorted (by uuid) unique pairs.
func (u UserPairs) dedupedPermutation() UserPairs {
	hash := map[string]UserPair{}

	for i := 0; i < len(u); i++ {
		a := u[i].UserA
		b := u[i].UserB

		// compare UUIDs as strings, and build sorted hash
		first := a
		second := b
		if first.ID.String() < second.ID.String() {
			first = b
			second = a
		}
		key := first.ID.String() + "-" + second.ID.String()
		hash[key] = UserPair{
			UserA: first,
			UserB: second,
		}
	}

	// fill deduped slice from unique sorted hash
	deduped := make(UserPairs, 0, len(hash))
	for _, v := range hash {
		deduped = append(deduped, v)
	}

	return deduped
}

type MatchSet struct {
	ID                   uuid.UUID `boil:"id" json:"id"`
	Name                 string    `boil:"name" json:"name"`
	NumberOfParticipants int       `boil:"number_of_participants" json:"number_of_participants"`
	MatchConfiguration   null.JSON `boil:"match_configuration" json:"match_configuration,omitempty"`
	TimeStart            null.Time `boil:"time_start" json:"time_start,omitempty"`
	TimeEnd              null.Time `boil:"time_end" json:"time_end,omitempty"`
	CreatedAt            null.Time `boil:"created_at" json:"created_at,omitempty"`
	UpdatedAt            null.Time `boil:"updated_at" json:"updated_at,omitempty"`
}

type MatchResultPaginated struct {
	Data       []MatchResult   `json:"data"`
	Pagination *sdk.Pagination `json:"pagination"`
}

// MatchResult is the result of matching two users.
// The result of this struct will be stored into a JSONB column in the DB.
// Alpha
type MatchResult struct {
	ID                   uuid.UUID   `boil:"id" json:"id"`
	MatchSetID           uuid.UUID   `boil:"match_set_id" json:"match_set_id"`
	InitiatorUserID      uuid.UUID   `boil:"initiator_user_id" json:"initiator_user_id"`
	ReceiverUserID       uuid.UUID   `boil:"receiver_user_id" json:"receiver_user_id"`
	QualifierResults     null.JSON   `boil:"qualifier_results" json:"qualifier_results,omitempty"`
	MatchedQualitatively null.Bool   `boil:"matched_qualitatively" json:"matched_qualitatively"`
	IsPossibleMatch      bool        `boil:"is_possible_match" json:"is_possible_match"`
	IsApproved           bool        `boil:"is_approved" json:"is_approved"`
	IsExpired            bool        `boil:"is_expired" json:"is_expired"`
	UserLifeCycleStatus  null.String `boil:"user_lifecycle_status" json:"user_lifecycle_status,omitempty"`

	/* enriched admin fields */
	InitiatorUserDetails *User          `json:"initiator_user_details,omitempty"`
	ReceiverUserDetails  *User          `json:"receiver_user_details,omitempty"`
	InitiatorUserProfile *PersonProfile `json:"initiator_user_profile,omitempty"`
	ReceiverUserProfile  *PersonProfile `json:"receiver_user_profile,omitempty"`
}

type QualifierType string

type Qualifier struct {
	Name             QualifierType     `json:"name"`
	QualifierResults *QualifierResults `json:"-"` // omit
	Telemetry        map[string]any    `json:"telemetry"`
	ErrorMsg         string            `json:"error_msg"`
}

func (q *Qualifier) SetError(err error) *Qualifier {
	q.ErrorMsg = err.Error()
	return q
}

type qualifiers []qualifier

func (q *QualifierResults) HasErrors() bool {
	return q.Error() != nil
}

// Error aggregates all qualifier errors into one.
func (q *QualifierResults) Error() error {
	var errList errutil.List
	qs := []*Qualifier{
		q.AgeWindow,
		q.DatingPreferences,
		q.Height,
		q.Distance,
		q.Qualitative,
	}

	for _, q_ := range qs {
		if q_ == nil {
			continue // ignore nil qualifiers
		}

		if q_.ErrorMsg != "" {
			errList.Add(q_.ErrorMsg)
		}
	}

	return errList.Error()
}

// ExecuteAll runs all qualifier logic collection
// that populates the Results.
func (qs qualifiers) ExecuteAll(
	ctx context.Context,
	config *Config,
	results *QualifierResults,
	parameters *QualifierParameters,
) *QualifierResults {
	parameters.QualifierResults = results // assign to parameters
	for _, q := range qs {
		q(ctx, parameters)
	}
	return results
}

type qualifier func(
	ctx context.Context,
	parameters *QualifierParameters,
	/*
	   ctx context.Context,
	   config *Config,
	   results *QualifierResults,
	   parameters *QualifierParameters,
	*/
) *Qualifier

func (q *QualifierResults) AsJSON() ([]byte, error) {
	return json.Marshal(q)
}

// QualifierResults is the different qualifiers we
// need to run 2 people against.
type QualifierResults struct {
	AgeWindow         *Qualifier `json:"age_window"`
	DatingPreferences *Qualifier `json:"dating_preferences"`
	Height            *Qualifier `json:"height"`
	Distance          *Qualifier `json:"distance"`
	Qualitative       *Qualifier `json:"qualitative"`

	// add more qualifiers as needed
}

type InsertMatchSet struct {
	Name                 string
	NumberOfParticipants int
	MatchingParameters   json.RawMessage
}

type UpdateMatchSet struct {
}

type UpdateUser struct {
	ID uuid.UUID
}

type QueryFilterUserDatingPrefs struct {
	UserID null.String
}

type QueryFilterUser struct {
	ID       null.String
	IsActive null.Bool

	// User type filters for batch matching exclusions
	IsTestUser null.Bool   // Filter by test user status
	UserType   null.String // Filter by user_type (Admin, Regular User)

	EnrichDatingPrefs bool
	EnrichAudioURLS   bool
	// EnrichProfiles   bool
	// there are details missing like listen
}

// MatchSetPaginated wraps a paginated list of MatchSets with pagination metadata.
type MatchSetPaginated struct {
	Data       []MatchSet      `json:"data"`
	Pagination *sdk.Pagination `json:"pagination"`
}

// QueryFilterMatchSet contains filter options for querying match sets.
type QueryFilterMatchSet struct {
	ID   null.String
	Name null.String // filter by name (partial match)

	// Participant count filters
	NumberOfParticipantsMin null.Int // filter number_of_participants >= value
	NumberOfParticipantsMax null.Int // filter number_of_participants <= value

	// Time range filters
	TimeStartAfter  null.Time // filter time_start >= value
	TimeStartBefore null.Time // filter time_start <= value
	TimeEndAfter    null.Time // filter time_end >= value
	TimeEndBefore   null.Time // filter time_end <= value
	CreatedAfter    null.Time // filter created_at >= value
	CreatedBefore   null.Time // filter created_at <= value

	// Sorting
	OrderBy null.String // column to order by
	Sort    null.String // "+" for ASC, "-" for DESC

	// Pagination
	Pagination *sdk.Pagination
}

type QueryFilterMatchResult struct {
	ID         null.String
	MatchSetID null.String

	// User filters - UserID matches EITHER initiator OR receiver (OR condition)
	UserID           null.String // matches user as either initiator OR receiver
	InitiatorUserID  null.String // matches user as initiator specifically
	ReceiverUserID   null.String // matches user as receiver specifically

	// Status filters (now string enum values instead of category UUIDs)
	MatchLifecycleStatus null.String // String enum - lifecycle status
	InitiatorAction          null.String // String enum - user A's action (Pending/Proposed/Passed)
	ReceiverAction          null.String // String enum - user B's action (Pending/Proposed/Passed)

	// Boolean filters
	MatchedQualitatively null.Bool
	IsVerified           null.Bool
	IsExpired            null.Bool
	IsApproved           null.Bool
	IsDropped            null.Bool
	IsPossibleMatch      null.Bool

	/* search helpers */
	OrderBy    null.String `json:"order_by"`
	Sort       null.String `json:"sort"`
	Pagination *sdk.Pagination

	/* enrichers */
	EnrichUsers bool // admin enriched
}

type InsertMatchResult struct {
	MatchSetID      uuid.UUID `boil:"match_set_id"`
	InitiatorUserID uuid.UUID `boil:"initiator_user_id"`
	ReceiverUserID  uuid.UUID `boil:"receiver_user_id"`
	StatusID        string    `boil:"status_id"`
}

type UpdateMatchResult struct {
	ID                   uuid.UUID
	MatchedQualitatively null.Bool
	QualifierResults     null.JSON
	IsVerified           null.Bool
	IsApproved           null.Bool
	IsDropped            null.Bool
	DroppedTS            null.Time
	IsPossibleMatch      null.Bool
	IsExpired            null.Bool
}

type InsertMatchParticipant struct {
	MatchSetID uuid.UUID
	UserPair   *UserPair
}

type Config struct {
	ID                        uuid.UUID         `boil:"id" json:"id"`
	AgeRangeStart             int               `boil:"age_range_start" json:"age_range_start"`
	AgeRangeEnd               int               `boil:"age_range_end" json:"age_range_end"`
	AgeRangeWomanOlderBy      int               `boil:"age_range_woman_older_by" json:"age_range_woman_older_by"`
	AgeRangeManOlderBy        int               `boil:"age_range_man_older_by" json:"age_range_man_older_by"`
	HeightMaleGreaterByCM     float64           `boil:"height_male_greater_by_cm" json:"height_male_greater_by_cm"`
	LocationRadiusKM          float64           `boil:"location_radius_km" json:"location_radius_km"`
	LocationAdaptiveExpansion types.Int64Array  `boil:"location_adaptive_expansion" json:"location_adaptive_expansion"`
	DropHours                 types.StringArray `boil:"drop_hours" json:"drop_hours"`
	DropHoursUTC              types.StringArray `boil:"drop_hours_utc" json:"drop_hours_utc"`
	StaleChatNudge            int               `boil:"stale_chat_nudge" json:"stale_chat_nudge"`
	StaleChatAgentSetup       int               `boil:"stale_chat_agent_setup" json:"stale_chat_agent_setup"`
	MatchExpirationHours      int               `boil:"match_expiration_hours" json:"match_expiration_hours"`
	MatchBlockDeclined        int               `boil:"match_block_declined" json:"match_block_declined"`
	MatchBlockIgnored         int               `boil:"match_block_ignored" json:"match_block_ignored"`
	MatchBlockClosed          int               `boil:"match_block_closed" json:"match_block_closed"`
	ScoreRangeStart           float64           `boil:"score_range_start" json:"score_range_start"`
	ScoreRangeEnd             float64           `boil:"score_range_end" json:"score_range_end"`
}

type QualifierParameters struct {
	config           *Config
	QualifierResults *QualifierResults
	UserA            *User
	UserB            *User
}

// Users returns the two users to be matched.
func (cp *QualifierParameters) Users() (*User, *User) {
	return cp.UserA, cp.UserB
}

type CompatibilityScore struct {
	Score       float64 `json:"score"`
	Explanation string  `json:"explanation"`
}

type MatchCompatibilityResult struct {
	PersonalityCompatibilityScore CompatibilityScore `json:"personality_compatibility_score"`
	LifestyleCompatibilityScore   CompatibilityScore `json:"lifestyle_compatibility_score"`
	ValuesCompatibilityScore      CompatibilityScore `json:"values_compatibility_score"`
	TotalScore                    float64            `json:"total_score"`
}

// Profile represents a user profile for qualitative matching.
type Profile struct {
}

type ProfileData struct {
	Romeo  *PersonProfile `json:"romeo"`
	Juliet *PersonProfile `json:"juliet"`
}

// Validate validates the ProfileData struct.
func (p *ProfileData) Validate() error {
	var errs errutil.List

	toValidate := []any{
		p.Romeo.Qualitative,
		p.Romeo.Quantitative,
		p.Romeo.Categorical,
		p.Juliet.Qualitative,
		p.Juliet.Quantitative,
		p.Juliet.Categorical,
	}

	for _, e := range toValidate {
		if err := validationlib.Validate(e); err != nil {
			errs.AddErr(err)
		}
	}

	return errs.Error()
}

type PersonProfile struct {
	Qualitative  QualitativeSection  `json:"Qualitative" validate:"required"`
	Quantitative QuantitativeSection `json:"Quantitative" validate:"required"`
	Categorical  CategoricalSection  `json:"Categorical" validate:"required"`
}

type QualitativeSection struct {
	SelfPortrait               string `boil:"self_portrait" json:"Self Portrait" validate:"required"`
	Interests                  string `boil:"interests" json:"Interests" validate:"required"`
	WellbeingHabits            string `boil:"wellbeing_habits" json:"Wellbeing Habits" validate:"required"`
	SelfCareHabits             string `boil:"self_care_habits" json:"Self Care Habits" validate:"required"`
	MoneyManagement            string `boil:"money_management" json:"Money Management" validate:"required"`
	SelfReflectionCapabilities string `boil:"self_reflection_capabilities" json:"Self Reflection Capabilities" validate:"required"`
	MoralFrameworks            string `boil:"moral_frameworks" json:"Moral Frameworks" validate:"required"`
	LifeGoals                  string `boil:"life_goals" json:"Life Goals" validate:"required"`
	PartnershipValues          string `boil:"partnership_values" json:"Partnership Values" validate:"required"`
	MutualCommitment           string `boil:"mutual_commitment" json:"Mutual Commitment" validate:"required"`
	SpiritualityGrowthMindset  string `boil:"spirituality_growth_mindset" json:"Spirituality Growth Mindset" validate:"required"`
	CulturalValues             string `boil:"cultural_values" json:"Cultural Values" validate:"required"`
	FamilyPlanning             string `boil:"family_planning" json:"Family Planning" validate:"required"`
	IdealDate                  string `boil:"ideal_date" json:"Ideal Date" validate:"required"`
	RedGreenFlags              string `boil:"red_green_flags" json:"Red/Green Flags" validate:"required"`
}

type QuantitativeSection struct {
	ExtroversionSocialEnergy float64 `boil:"extroversion_social_energy" json:"Extroversion Social Energy" validate:"required"`
	RoutineVsSpontaneity     float64 `boil:"routine_vs_spontaneity" json:"Routine vs Spontaneity" validate:"required"`
	Agreeableness            float64 `boil:"agreeableness" json:"Agreeableness" validate:"required"`
	Conscientiousness        float64 `boil:"conscientiousness" json:"Conscientiousness" validate:"required"`
	Neuroticism              float64 `boil:"neuroticism" json:"Neuroticism" validate:"required"`
	DominanceLevel           float64 `boil:"dominance_level" json:"Dominance Level" validate:"required"`
	EmotionalExpressiveness  float64 `boil:"emotional_expressiveness" json:"Emotional Expressiveness" validate:"required"`
	SexDrive                 float64 `boil:"sex_drive" json:"Sex Drive" validate:"required"`
	GeographicalMobility     float64 `boil:"geographical_mobility" json:"Geographical Mobility" validate:"required"`
}

type CategoricalSection struct {
	ConflictResolutionStyle string `boil:"conflict_resolution_style" json:"Conflict Resolution Style" validate:"required"`
	SexualityPreferences    string `boil:"sexuality_preferences" json:"Sexuality Preferences" validate:"required"`
	Religion                string `boil:"religion" json:"Religion" validate:"required"`
}

type Category struct {
	ID   string
	Name string
}

// ============================================================================
// USER-FACING MATCHING MODELS
// These models support the user-facing matching endpoints (propose, pass)
// ============================================================================

// MatchUserAction represents the user's action on a match (Pending/Proposed/Passed)
const (
	MatchUserActionCategoryType = "Match User Action"
	MatchUserActionPending      = "Pending"
	MatchUserActionProposed     = "Proposed"
	MatchUserActionPassed       = "Passed"
)

// UserMatch represents a match from a single user's perspective.
// This is the user-facing view of a match_result.
type UserMatch struct {
	ID               uuid.UUID    `boil:"id" json:"id"`
	PartnerID        uuid.UUID    `boil:"partner_id" json:"partner_id"`
	PartnerName      null.String  `boil:"partner_name" json:"partner_name"`
	PartnerAge       null.Int     `boil:"partner_age" json:"partner_age"`
	PartnerGender    null.String  `boil:"partner_gender" json:"partner_gender"`
	PartnerSexuality null.String  `boil:"partner_sexuality" json:"partner_sexuality"`
	DistanceKM       null.Float64 `boil:"distance_km" json:"distance_km,omitempty"`
	MatchScore       null.Float64 `boil:"match_score" json:"match_score,omitempty"`
	YourAction       string       `boil:"your_action" json:"your_action"`           // Pending/Proposed/Passed
	PartnerAction    string       `boil:"partner_action" json:"partner_action"`     // Pending/Proposed/Passed
	MutualProposal   bool         `boil:"mutual_proposal" json:"mutual_proposal"`   // Both proposed
	DateInstanceID   null.String  `boil:"date_instance_id" json:"date_instance_id"` // Linked date instance (if mutual proposal)
	SeenAt           null.Time    `boil:"seen_at" json:"seen_at,omitempty"`
	ExpiresAt        null.Time    `boil:"expires_at" json:"expires_at,omitempty"`
	DroppedAt        null.Time    `boil:"dropped_at" json:"dropped_at,omitempty"`

	// Internal fields for enrichment (not exposed to JSON)
	YourSupabaseID    null.String `boil:"your_supabase_id" json:"-"`
	PartnerSupabaseID null.String `boil:"partner_supabase_id" json:"-"`

	// Enriched fields (populated after query if EnrichAudio/EnrichPhotos is set)
	YourIntroURL    null.String `json:"your_intro_url,omitempty"`                     // Current user's intro audio URL
	PartnerIntroURL null.String `json:"partner_intro_url,omitempty"`                  // Partner's intro audio URL
	LovestoryURL    null.String `boil:"lovestory_url" json:"lovestory_url,omitempty"` // First date simulation audio URL from ai_backend
	PartnerPhotos   []UserPhoto `json:"partner_photos,omitempty"`
}

// UserPhoto represents a user's profile photo
type UserPhoto struct {
	URL     string `boil:"url" json:"url"`
	OrderNo int    `boil:"order_no" json:"order_no"`
}

// QueryFilterUserMatch filters for user-facing match queries
type QueryFilterUserMatch struct {
	UserID         uuid.UUID   // Required: the current user
	MatchID        null.String // Filter by specific match
	YourAction     null.String // Filter by your action (Pending/Proposed/Passed)
	PartnerAction  null.String // Filter by partner action
	MutualProposal null.Bool   // Filter by mutual proposal status (both users proposed)
	IsExpired      null.Bool   // Filter expired matches
	Unseen         null.Bool   // Filter by unseen status (user_*_seen_at IS NULL)
	SeenOnly       null.Bool   // Filter by seen status (user_*_seen_at IS NOT NULL)

	// Sorting
	OrderBy null.String // Field: dropped_at, expires_at
	Sort    null.String // + for ASC, - for DESC

	// Pagination
	Page null.Int // Page number (1-indexed)
	Rows null.Int // Items per page

	// Enrichment
	EnrichPhotos    bool // Fetch partner photos
	EnrichAudio     bool // Fetch intro audio URLs from ai_backend
	EnrichLovestory bool // Fetch lovestory audio URLs from ai_backend
}

// UserMatchPaginated wraps user matches with pagination info
type UserMatchPaginated struct {
	Data       []UserMatch `json:"data"`
	Page       int         `json:"page"`
	Rows       int         `json:"rows"`
	TotalRows  int         `json:"total_rows"`
	TotalPages int         `json:"total_pages"`
}

// ProposeMatchParams contains parameters for proposing on a match
type ProposeMatchParams struct {
	MatchResultID uuid.UUID
	UserID        uuid.UUID
}

// ProposeMatchResult contains the result of a propose action
type ProposeMatchResult struct {
	Success        bool   `json:"success"`
	MutualProposal bool   `json:"mutual_proposal"`            // True if both users have proposed
	DateInstanceID string `json:"date_instance_id,omitempty"` // Populated on mutual proposal
}

// PassMatchParams contains parameters for passing on a match
type PassMatchParams struct {
	MatchResultID uuid.UUID
	UserID        uuid.UUID
}

// MarkSeenParams contains parameters for marking matches as seen
type MarkSeenParams struct {
	UserID   uuid.UUID
	MatchIDs []uuid.UUID
}

// UpdateUserMatchAction contains parameters for updating a user's action on a match.
// Used by the store layer for CRUD operations.
type UpdateUserMatchAction struct {
	MatchResultID    uuid.UUID
	UserID           uuid.UUID
	ActionCategoryID null.String // Set to update the user's action (Pending/Proposed/Passed)
	SeenAt           null.Time   // Set to mark as seen
}

// InsertDateInstance contains parameters for creating a date instance.
type InsertDateInstance struct {
	MatchResultRefID  string
	Status            string // Required: Date Instance Status enum value
	DecisionWindowEnd time.Time
	DateTypeCore      null.String // Date type core enum value
	ScheduledTimeUTC  null.Time
	DurationMinutes   null.Int
}

// InsertDateInstanceLog contains parameters for logging a date instance event.
type InsertDateInstanceLog struct {
	DateInstanceRefID string
	UserRefID         null.String
	EventType         string
	OldValue          null.JSON
	NewValue          null.JSON
	Details           null.String
}

// UpdateMatchForDateInstance links a match to its date instance.
type UpdateMatchForDateInstance struct {
	MatchResultID         string
	CurrentDateInstanceID string
	MatchLifecycleStatus  string // Match lifecycle status enum value
}

// QueryFilterMatchConfig contains filter options for querying match configurations.
// Note: match_config is a singleton table - typically only one row exists.
type QueryFilterMatchConfig struct {
	ID null.String

	// Age range filters
	AgeRangeStartMin null.Int // filter age_range_start >= value
	AgeRangeStartMax null.Int // filter age_range_start <= value
	AgeRangeEndMin   null.Int // filter age_range_end >= value
	AgeRangeEndMax   null.Int // filter age_range_end <= value

	// Height filter
	HeightMaleGreaterByCMMin null.Float64 // filter height_male_greater_by_cm >= value
	HeightMaleGreaterByCMMax null.Float64 // filter height_male_greater_by_cm <= value

	// Location radius filter
	LocationRadiusKMMin null.Float64 // filter location_radius_km >= value
	LocationRadiusKMMax null.Float64 // filter location_radius_km <= value

	// Drop hours filters
	DropHour     null.String // filter by specific drop hour in array (e.g., "19:00")
	DropHourUTC  null.String // filter by specific timezone in array (e.g., "GMT+3")
	HasDropHours null.Bool   // filter configs that have drop hours configured

	// Stale chat filters
	StaleChatNudgeMin      null.Int // filter stale_chat_nudge >= value
	StaleChatNudgeMax      null.Int // filter stale_chat_nudge <= value
	StaleChatAgentSetupMin null.Int // filter stale_chat_agent_setup >= value
	StaleChatAgentSetupMax null.Int // filter stale_chat_agent_setup <= value

	// Match expiration/block filters
	MatchExpirationHoursMin null.Int // filter match_expiration_hours >= value
	MatchExpirationHoursMax null.Int // filter match_expiration_hours <= value

	// Score range filters
	ScoreRangeStartMin null.Float64 // filter score_range_start >= value
	ScoreRangeStartMax null.Float64 // filter score_range_start <= value
	ScoreRangeEndMin   null.Float64 // filter score_range_end >= value
	ScoreRangeEndMax   null.Float64 // filter score_range_end <= value

	// Sorting
	OrderBy null.String // column to order by
	Sort    null.String // "+" for ASC, "-" for DESC
}

// UpdateMatchConfig contains optional fields for updating match configuration.
// All fields are nullable since this is a PATCH operation - only provided fields are updated.
type UpdateMatchConfig struct {
	// Age range settings
	AgeRangeStart        null.Int `json:"age_range_start"`
	AgeRangeEnd          null.Int `json:"age_range_end"`
	AgeRangeWomanOlderBy null.Int `json:"age_range_woman_older_by"`
	AgeRangeManOlderBy   null.Int `json:"age_range_man_older_by"`

	// Height settings
	HeightMaleGreaterByCM null.Float64 `json:"height_male_greater_by_cm"`

	// Location settings
	LocationRadiusKM          null.Float64      `json:"location_radius_km"`
	LocationAdaptiveExpansion *types.Int64Array `json:"location_adaptive_expansion"` // pointer to distinguish between null and empty

	// Schedule settings
	DropHours    *types.StringArray `json:"drop_hours"`     // pointer to distinguish between null and empty
	DropHoursUTC *types.StringArray `json:"drop_hours_utc"` // pointer to distinguish between null and empty

	// Stale chat settings
	StaleChatNudge      null.Int `json:"stale_chat_nudge"`
	StaleChatAgentSetup null.Int `json:"stale_chat_agent_setup"`

	// Match expiration/block settings
	MatchExpirationHours null.Int `json:"match_expiration_hours"`
	MatchBlockDeclined   null.Int `json:"match_block_declined"`
	MatchBlockIgnored    null.Int `json:"match_block_ignored"`
	MatchBlockClosed     null.Int `json:"match_block_closed"`

	// Score range settings
	ScoreRangeStart null.Float64 `json:"score_range_start"`
	ScoreRangeEnd   null.Float64 `json:"score_range_end"`
}

// Validate validates the UpdateMatchConfig with business rules.
func (u *UpdateMatchConfig) Validate() error {
	if u == nil {
		return nil
	}

	// Validate adaptive expansion array is strictly incrementing
	if u.LocationAdaptiveExpansion != nil && len(*u.LocationAdaptiveExpansion) > 1 {
		arr := *u.LocationAdaptiveExpansion
		for i := 1; i < len(arr); i++ {
			if arr[i] <= arr[i-1] {
				return ErrAdaptiveExpansionNotIncrementing
			}
		}
	}

	// Validate age range consistency (if both provided)
	if u.AgeRangeStart.Valid && u.AgeRangeEnd.Valid {
		if u.AgeRangeStart.Int > u.AgeRangeEnd.Int {
			return ErrAgeRangeStartGreaterThanEnd
		}
	}

	// Validate score range consistency (if both provided)
	if u.ScoreRangeStart.Valid && u.ScoreRangeEnd.Valid {
		if u.ScoreRangeStart.Float64 > u.ScoreRangeEnd.Float64 {
			return ErrScoreRangeStartGreaterThanEnd
		}
	}

	// Validate drop hours format (HH:MM)
	if u.DropHours != nil {
		for _, h := range *u.DropHours {
			if !isValidTimeFormat(h) {
				return ErrDropHoursInvalidFormat
			}
		}
	}

	// Validate drop hours UTC format (e.g., "GMT+3", "UTC-5")
	if u.DropHoursUTC != nil {
		for _, tz := range *u.DropHoursUTC {
			if !isValidTimezoneFormat(tz) {
				return ErrDropHoursUTCInvalidFormat
			}
		}
	}

	// Validate non-negative values
	if u.StaleChatNudge.Valid && u.StaleChatNudge.Int < 0 {
		return ErrNegativeValue
	}
	if u.StaleChatAgentSetup.Valid && u.StaleChatAgentSetup.Int < 0 {
		return ErrNegativeValue
	}
	if u.MatchExpirationHours.Valid && u.MatchExpirationHours.Int < 0 {
		return ErrNegativeValue
	}
	if u.MatchBlockDeclined.Valid && u.MatchBlockDeclined.Int < 0 {
		return ErrNegativeValue
	}
	if u.MatchBlockIgnored.Valid && u.MatchBlockIgnored.Int < 0 {
		return ErrNegativeValue
	}
	if u.MatchBlockClosed.Valid && u.MatchBlockClosed.Int < 0 {
		return ErrNegativeValue
	}
	if u.LocationRadiusKM.Valid && u.LocationRadiusKM.Float64 < 0 {
		return ErrNegativeValue
	}
	if u.HeightMaleGreaterByCM.Valid && u.HeightMaleGreaterByCM.Float64 < 0 {
		return ErrNegativeValue
	}

	return nil
}

// isValidTimeFormat checks if a string matches HH:MM format
func isValidTimeFormat(t string) bool {
	if len(t) < 4 || len(t) > 5 {
		return false
	}
	// Match pattern: H:MM or HH:MM where H is 0-23, MM is 00-59
	parts := make([]byte, 0, 5)
	colonIdx := -1
	for i, c := range t {
		if c == ':' {
			colonIdx = i
		}
		parts = append(parts, byte(c))
	}
	if colonIdx < 1 || colonIdx > 2 {
		return false
	}

	// Parse hour
	hourStr := t[:colonIdx]
	minStr := t[colonIdx+1:]

	if len(minStr) != 2 {
		return false
	}

	hour := 0
	for _, c := range hourStr {
		if c < '0' || c > '9' {
			return false
		}
		hour = hour*10 + int(c-'0')
	}
	if hour > 23 {
		return false
	}

	min := 0
	for _, c := range minStr {
		if c < '0' || c > '9' {
			return false
		}
		min = min*10 + int(c-'0')
	}
	if min > 59 {
		return false
	}

	return true
}

// isValidTimezoneFormat checks if a string matches GMT+N or UTC+N format
func isValidTimezoneFormat(tz string) bool {
	if len(tz) < 5 || len(tz) > 6 {
		return false
	}

	prefix := tz[:3]
	if prefix != "GMT" && prefix != "UTC" {
		return false
	}

	sign := tz[3]
	if sign != '+' && sign != '-' {
		return false
	}

	offsetStr := tz[4:]
	if len(offsetStr) == 0 || len(offsetStr) > 2 {
		return false
	}

	offset := 0
	for _, c := range offsetStr {
		if c < '0' || c > '9' {
			return false
		}
		offset = offset*10 + int(c-'0')
	}

	return offset <= 14 // valid timezone offsets are -12 to +14
}
