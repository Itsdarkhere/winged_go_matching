package sysparam

import (
	"context"
)

// Settings represents application settings.
type Settings struct {

	/* transcripts */

	SettingUpAgentTranscriptMaxLoops         int `json:"setting_up_agent_transcript_max_loops"`
	SettingUpAgentTranscriptLoopIntervalSecs int `json:"setting_up_agent_transcript_loop_interval_secs"`

	/* audio files */

	SettingUpAgentAudioFilesMaxLoops         int `json:"setting_up_agent_audio_files_max_loops"`
	SettingUpAgentAudioFilesLoopIntervalSecs int `json:"setting_up_agent_audio_files_loop_interval_secs"`

	/* idle waiting time */

	SettingUpAgentIdleWaitingTime int `json:"setting_up_agent_idle_waiting_time_secs"`

	/* user invite codes */

	UserInviteCodeMaxUsage int `json:"user_invite_code_max_usage"`

	/* broken audio handling */

	BrokenAudioDurationThresholdSecs int `json:"broken_audio_duration_threshold_secs"`

	/* invite friend code expires in */

	InviteExpiryDays int `json:"invite_expiry_days"`

	/* wings economy */

	WingsEconIncrementThreshSendMessage int `json:"wings_econ_increment_thresh_send_message"`
	WingsEconIncrementRewardSendMessage int `json:"wings_econ_increment_reward_send_message"`

	WingsEconWingedPlusWeeklyPaymentAmount float64 `json:"wings_econ_winged_plus_weekly_payment_amount"`
	WingsEconWingedPlusWeeklyWingsAmount   int     `json:"wings_econ_winged_plus_weekly_wings_amount"`
}

type getter interface {
	Settings(ctx context.Context) (*Settings, error)
}

// Getter provides application settings.
type Getter struct {
	getter // getter
}

// Settings retrieves application settings.
func (g *Getter) Settings(ctx context.Context) (*Settings, error) {
	return g.getter.Settings(ctx)
}

// NewGetter creates a new Getter.
func NewGetter(g getter) *Getter {
	return &Getter{g}
}
