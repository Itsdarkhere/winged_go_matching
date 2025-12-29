package registration

import (
	"encoding/json"
	"time"

	"github.com/aarondl/null/v8"
)

// ElevenLabsCallState determines the call state for 11 labs.
type ElevenLabsCallState struct {
	LastAttempt       null.Time `json:"last_attempt"`
	Completed         bool      `json:"completed"`
	PastConversations []string  `json:"past-conversations"`
}

// OauthUser represents a user registered via OAuth.
type OauthUser struct {
	SupabaseID       string      `json:"supabase_id" validate:"required"`
	Email            string      `json:"email" validate:"required"`
	RegistrationCode null.String `json:"registration_code" validate:"required"`
}

type InsertUserElevenLabs struct {
	UserID       string `json:"user_id"`
	Conversation []byte `json:"conversation"`
}

type User2 struct {
	ID string `json:"id" boil:"id"`
}

// User is our main user model.
type User struct {
	ID                     string      `json:"id" boil:"id"`
	SupabaseID             null.String `json:"supabase_id" boil:"supabase_id"`
	FirstName              null.String `json:"first_name" boil:"first_name"`
	LastName               null.String `json:"last_name" boil:"last_name"`
	Birthday               null.Time   `json:"birthday" boil:"birthday"`
	Email                  string      `json:"email" boil:"email"`
	Gender                 null.String `json:"gender" boil:"gender"`
	Height                 null.Int    `json:"height" boil:"height"`
	RegistrationCode       null.String `json:"registration_code" boil:"registration_code"`
	MobileConfirmed        bool        `json:"mobile_confirmed" boil:"mobile_confirmed"`
	MobileNumber           null.String `json:"mobile_number" boil:"mobile_number"`
	MobileCode             null.String `json:"mobile_code" boil:"mobile_code"`
	RegisteredSuccessfully null.Bool   `json:"registered_successfully" boil:"registered_successfully"`
	RegistrationCodeSentAt null.Time   `json:"registration_code_sent_at" boil:"registration_code_sent_at"`
	LastCheckedCallStatus  null.Time   `json:"last_checked_call_status" boil:"last_checked_call_status"`
	AgentDeployed          bool        `json:"agent_deployed" boil:"agent_deployed"`
	SelectedIntroID        null.String `json:"selected_intro_id" boil:"selected_intro_id"`
	Sexuality              null.String `json:"sexuality" boil:"sexuality"`
	SexualityIsVisible     null.Bool   `json:"sexuality_is_visible" boil:"sexuality_is_visible"`
	UserTypeID             null.String `json:"user_type_id" boil:"user_type_id"`
	UserType               null.String `json:"user_type" boil:"user_type"`
	UserInviteCodeRefID    null.String `json:"user_invite_code_ref_id" boil:"user_invite_code_ref_id"`

	/* transcript timestamps */
	LatestTranscriptID null.String `json:"latest_transcript_id" boil:"latest_transcript_id"`
	LatestTranscriptTS null.Time   `json:"latest_transcript_ts" boil:"latest_transcript_ts"`
	HasTranscript      bool        `json:"has_transcript" boil:"has_transcript"`

	/* enriched cols */
	Photos      []UserPhoto      `json:"photos" boil:"-"`
	Transcripts []UserTranscript `json:"transcripts" boil:"-"`
	AudioFiles  []UserAudio      `json:"audio_files" boil:"-"`

	DatingPreferences []string `json:"dating_preferences" boil:"-"`
}

type UserCall struct {
	ID           string          `json:"id"`
	Conversation json.RawMessage `json:"conversation"`
	AudioBytes   []byte          `json:"audio_bytes"`
}

type UserPhoto struct {
	ID       string `json:"id"`
	Bytes    []byte `json:"bytes"`
	UserID   string `json:"user_id"`
	Bucket   string `json:"bucket"`
	Key      string `json:"key"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
	OrderNo  int    `json:"order_no"`
}

type UpsertUser struct {
	ID                     null.String `json:"id"`
	SupabaseID             null.String `json:"supabase_id"`
	Email                  null.String `json:"email"`
	FirstName              null.String `json:"first_name"`
	LastName               null.String `json:"last_name"`
	Birthday               null.Time   `json:"birthday"`
	Gender                 null.String `json:"gender"`
	Height                 null.Int    `json:"height"`
	Number                 null.String `json:"number"`
	RegistrationCode       null.String `json:"registration_code"`
	RegistrationCodeSentAt null.Time   `json:"registration_code_sent_at"`
}

// UpdateUser represents the data needed to update/update a user.
type UpdateUser struct {
	ID                     string      `json:"id"`
	Email                  null.String `json:"email"`
	Sha256Hash             null.String `json:"sha256_hash"`
	FirstName              null.String `json:"first_name"`
	LastName               null.String `json:"last_name"`
	Birthday               null.Time   `json:"birthday"`
	Gender                 null.String `json:"gender"`
	Height                 null.Int    `json:"height"`
	Number                 null.String `json:"number"`
	RegisteredSuccessfully null.Bool   `json:"registered_successfully"`
	MobileCode             null.String `json:"mobile_code"`
	MobileConfirmed        null.Bool   `json:"mobile_confirmed"`
	RegistrationCode       null.String `json:"registration_code"`
	RegistrationCodeSentAt null.Time   `json:"registration_code_sent_at"`
	LastCheckedCallStatus  null.Time   `json:"last_checked_call_status"`
	AgentDeployed          null.Bool   `json:"agent_deployed"`
	SelectedVoiceID        null.String `json:"selected_voice_id"`
	UserInviteCodeID       null.String `json:"user_invite_code_id"`
	SexualityCategoryID    null.String `json:"sexuality"`
	SexualityIsVisible     null.Bool   `json:"sexuality_is_visible"`

	DatingPreferences UpdateUserDatingPreferences `json:"user_dating_preferences"`

	HasTranscript      null.Bool   `json:"has_transcript"`
	LatestTranscriptID null.String `json:"latest_transcript_id"`
	LatestTranscriptTS null.Time   `json:"latest_transcript_ts"`
}

type UpdateUserDatingPreferences struct {
	Male      null.Bool
	Female    null.Bool
	NonBinary null.Bool
}

type UserInviteCode struct {
	ID         string    `json:"id"`
	Code       string    `json:"code"`
	Category   string    `json:"category"`
	ForNumber  string    `json:"for_number"`
	UsageCount int       `json:"usage_count"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsed   null.Time `json:"last_used"`
}

type InsertUserBlockedContact struct {
	UserID string `json:"user_id"`
	Number string `json:"number"`
}

type UserBlockedContactQueryFilter struct {
	ID            null.String `json:"id"`
	UserID        null.String `json:"user_id"`
	BlockedNumber null.String `json:"blocked_number"`
}

type UserBlockedContact struct {
	ID            string
	UserID        string
	BlockedNumber null.String
}

type FileDetails struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

type UserPhotoQueryFilter struct {
	ID           null.String `json:"id"`
	UserID       null.String `json:"user_id"`
	Key          null.String `json:"key"`
	Bucket       null.String `json:"bucket"`
	Order        null.Int    `json:"order"`
	EnrichPhotos bool        `json:"enrich_photos"`
}

type Profile struct {
	ID                         string       `json:"id"`
	UserID                     null.String  `json:"user_id"`
	CreatedAt                  time.Time    `json:"created_at"`
	UpdatedAt                  null.Time    `json:"updated_at"`
	ConversationID             null.String  `json:"conversation_id"`
	WellbeingHabits            null.String  `json:"wellbeing_habits"`
	Interests                  null.String  `json:"interests"`
	SenseOfHumor               null.String  `json:"sense_of_humor"`
	SelfCareHabits             null.String  `json:"self_care_habits"`
	MoneyManagement            null.String  `json:"money_management"`
	SelfReflectionCapabilities null.String  `json:"self_reflection_capabilities"`
	MoralFrameworks            null.String  `json:"moral_frameworks"`
	LifeGoals                  null.String  `json:"life_goals"`
	PartnershipValues          null.String  `json:"partnership_values"`
	SpiritualityGrowthMindset  null.String  `json:"spirituality_growth_mindset"`
	CulturalValues             null.String  `json:"cultural_values"`
	FamilyPlanning             null.String  `json:"family_planning"`
	ExtroversionSocialEnergy   null.Float64 `json:"extroversion_social_energy"`
	RoutineVSSpontaneity       null.Float64 `json:"routine_vs_spontaneity"`
	Agreeableness              null.Float64 `json:"agreeableness"`
	Conscientiousness          null.Float64 `json:"conscientiousness"`
	DominanceLevel             null.Float64 `json:"dominance_level"`
	EmotionalExpressiveness    null.Float64 `json:"emotional_expressiveness"`
	SexDrive                   null.Float64 `json:"sex_drive"`
	GeographicalMobility       null.Float64 `json:"geographical_mobility"`
	ConflictResolutionStyle    null.String  `json:"conflict_resolution_style"`
	SexualityPreferences       null.String  `json:"sexuality_preferences"`
	Religion                   null.String  `json:"religion"`
	SelfPortrait               null.String  `json:"self_portrait"`
	MutualCommitment           null.String  `json:"mutual_commitment"`
	IdealDate                  null.String  `json:"ideal_date"`
	RedGreenFlags              null.String  `json:"red_green_flags"`
	Neuroticism                null.Float32 `json:"neuroticism"`
}

type ProfileQueryFilter struct {
	ID     string      `json:"id"`
	UserID null.String `json:"user_id"`
}

type UserAudio struct {
	ID             string `json:"id"`
	Category       string `json:"category"`
	UserID         string `json:"user_id"`
	ConversationID string `json:"conversation_id"`
	URL            string `json:"url"`
	StoragePath    string `json:"storage_path"`
}

type AudioFileQueryFilter struct {
	ID             string      `json:"id"`
	UserID         null.String `json:"user_id"`
	AgentID        string      `json:"agent_id"`
	ConversationID null.String `json:"conversation_id"`
	Categories     []string    `json:"categories"`
	Limit          null.Int    `json:"limit"`
	Offset         null.Int    `json:"offset"`
	HasStorage     null.Bool   `json:"has_storage"`
}

type CallState string

const (
	CallSuccessful     CallState = "call_successful"
	CallWaitForSuccess CallState = "call_wait_for_success"
	CallRetry          CallState = "call_retry"
	CallFailed         CallState = "call_failed"
)

type UserCallState struct {
	CallState               CallState `json:"call_state"`
	HasCompletedAudioFiles  bool      `json:"has_completed_audio_files"`
	HasSuccessfulTranscript bool      `json:"has_successful_transcript"`
	HasBrokenAudioFiles     bool      `json:"has_broken_audio_files"`
}

type UserTranscript struct {
	ID             string       `json:"id"`
	UserID         string       `json:"user_id"`
	ConversationID string       `json:"conversation_id"`
	Status         string       `json:"status"`
	CallSuccessful string       `json:"call_successful"`
	CreatedAt      time.Time    `json:"created_at"`
	TranscriptData []Transcript `json:"transcript_data"`
}

type Transcript struct {
	Role           string `json:"role"`
	Message        string `json:"message"`
	TimeInCallSecs int    `json:"time_in_call_secs"`
}

type UserTranscripts []UserTranscript

// Successful filters and returns only successful transcripts.
func (t UserTranscripts) Successful() UserTranscripts {
	var ts UserTranscripts
	if len(t) == 0 {
		return ts
	}

	for _, tr := range t {
		if tr.Status == "success" {
			ts = append(ts, tr)
		}
	}

	return ts
}

type TranscriptQueryFilters struct {
	ID                       string      `json:"id"`
	UserID                   null.String `json:"user_id"`
	LatestTranscriptID       null.String `json:"latest_transcript_id"`
	CategoriesCallSuccessful []string    `json:"categories_call_successful"`
	Limit                    null.Int    `json:"limit"`
	OrderedBys               []string    `json:"ordered_bys"`
}

type VoiceQueryFilter struct {
	ID     null.String `json:"id"`
	UserID null.String `json:"user_id"`
}

type Voice struct {
	ID    string `json:"id"`
	Bytes []byte `json:"bytes"`
}

type UpdateUserInviteCode struct {
	ID         string    `json:"id"`
	UsageCount null.Int  `json:"usage_count"`
	LastUsed   null.Time `json:"last_used"`
}
