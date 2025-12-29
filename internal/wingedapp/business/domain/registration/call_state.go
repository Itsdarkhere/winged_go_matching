package registration

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
)

// updateUserLastCheckedCallStatus updates the user's last checked call status timestamp.
func (b *Business) updateUserLastCheckedCallStatus(ctx context.Context, userID string, t *time.Time) error {
	tx, err := b.transBE.TX()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer b.transBE.Rollback(tx)

	_, err = b.storer.UpdateUser(ctx, tx, b.dbAI(), &UpdateUser{
		ID:                    userID,
		LastCheckedCallStatus: null.TimeFrom(*t),
	})
	if err != nil {
		return fmt.Errorf("update user last checked call status: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// userWithCallState retrieves a user along with their call state information.
func (b *Business) userWithCallState(ctx context.Context, userID string) (*User, error) {
	user, err := b.storer.User(ctx, b.dbBE(), b.dbAI(), &QueryFilterUser{
		ID:               null.StringFrom(userID),
		EnrichCallStates: true,
	})
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	return user, nil
}

// userCallFailed checks if the user's call has failed based on transcript data.
func userCallFailed(user *User) bool {
	hasTranscript := user.HasTranscript
	userLatestTrans := user.LatestTranscriptID
	userTrans := user.Transcripts

	if !hasTranscript {
		return false // no transcript, no concrete failure
	}
	if len(userTrans) == 0 {
		return false // no transcripts, no concrete failure
	}

	// concrete failure: latest transcript Code is different from the latest saved transcript Code
	return userLatestTrans.String != userTrans[0].ID
}

// hasSuccessfulPostCallTranscript checks if there are successful post-call transcripts.
func hasSuccessfulPostCallTranscript(ut []UserTranscript, userTalkingMinFailDuration int) (bool, *UserTranscript) {
	for _, t := range ut {
		if t.Status == TranscriptDone && t.CallSuccessful == TranscriptSuccess {
			noBrokenAudios := !userTransHasBrokenAudios(&t, userTalkingMinFailDuration)

			if noBrokenAudios {
				return true, &t
			}
		}
	}
	return false, nil
}

// userTransHasBrokenAudios checks if the user's audio files are broken based on transcript data.
// Broken meaning the user talking duration is less than the minimum fail duration, which
// 11labs requires.
func userTransHasBrokenAudios(ut *UserTranscript, userTalkingMinFailDuration int) bool {
	// false if not provided
	if ut == nil {
		return false
	}

	userTalkingDuration := 0
	for _, tData := range ut.TranscriptData {
		// increment only user talking duration
		if tData.Role == transcriptDataRoleUser {
			userTalkingDuration = userTalkingDuration + tData.TimeInCallSecs
		}
	}

	return userTalkingDuration < userTalkingMinFailDuration
}

// hasCompletedAudioFiles checks if there are successful post-call audio files.
func hasCompletedAudioFiles(successUserTrans *UserTranscript, a []UserAudio) bool {

	if successUserTrans == nil {

		return false // no successful transcript, no completed audio files
	}

	var hasExciting, hasGeneric, hasVulnerable bool
	for _, af := range a {
		if af.ConversationID != successUserTrans.ConversationID {

			continue // skip non-matching user audio files
		}

		switch af.Category {
		case AudioFileExciting:

			hasExciting = true
		case AudioFileGeneric:

			hasGeneric = true
		case AudioFileVulnerable:

			hasVulnerable = true
		default:

		}
	}

	return hasExciting && hasGeneric && hasVulnerable
}

// UpdateUserLatestCallState sets the anchor point for the
// next call status check by saving the latest checked time,
// and latest transcript Code if provided.
func (b *Business) UpdateUserLatestCallState(ctx context.Context, supabaseID, userID string) error {
	var hasNoTranscript bool
	latestTranscript, err := b.storer.Transcript(ctx, b.dbAI(), &TranscriptQueryFilters{
		UserID:     null.StringFrom(supabaseID),
		Limit:      null.IntFrom(1),
		OrderedBys: []string{"-created_at"},
	})
	if err != nil {
		hasNoTranscript = errors.Is(err, ErrUserTranscriptNotFound)
		if !hasNoTranscript {
			return fmt.Errorf("latestTranscript: %w", err)
		}
		// continue with nil latestTranscript
	}

	u := UpdateUser{
		ID:            userID,
		HasTranscript: null.BoolFrom(!hasNoTranscript),
	}
	if !hasNoTranscript {
		u.LatestTranscriptID = null.StringFrom(latestTranscript.ID)
		u.LatestTranscriptTS = null.TimeFrom(latestTranscript.CreatedAt)
	}

	// update user record
	tx, err := b.transBE.TX()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	_, err = b.storer.UpdateUser(ctx, tx, b.dbAI(), &u)
	if err != nil {
		return fmt.Errorf("update user latest call state: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// UserCallState checks the user's call state.
func (b *Business) UserCallState(ctx context.Context, user *User) (*UserCallState, error) {
	now := time.Now()
	user, err := b.userWithCallState(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get user with call state: %w", err)
	}

	// update last checked ts if applicable
	firstTimeCheck := !user.LastCheckedCallStatus.Valid

	threshold := time.Duration(b.cfg.LastCallThresholdIntervalSecs) * time.Second
	lastChecked := time.Since(user.LastCheckedCallStatus.Time)
	exceededThresh := lastChecked > threshold

	if firstTimeCheck || exceededThresh {
		if err = b.updateUserLastCheckedCallStatus(ctx, user.ID, &now); err != nil {
			return nil, fmt.Errorf("update user call state timestamp: %w", err)
		}
	}

	settings, err := b.settingGetter.Settings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get app settings: %w", err)
	}

	// check successful call state
	var callState CallState
	hasSuccessfulTrans, successUserTrans := hasSuccessfulPostCallTranscript(user.Transcripts, settings.BrokenAudioDurationThresholdSecs)
	hasNewFailedTrans := userCallFailed(user)
	hasCompleteAudioFiles := hasCompletedAudioFiles(successUserTrans, user.AudioFiles)
	hasBrokenAudios := userTransHasBrokenAudios(successUserTrans, settings.BrokenAudioDurationThresholdSecs)

	switch {
	case hasSuccessfulTrans && !hasBrokenAudios:
		callState = CallSuccessful
		break
	case hasNewFailedTrans:
		callState = CallFailed
		break
	case exceededThresh:
		callState = CallRetry
		break
	default:
		callState = CallWaitForSuccess
	}

	ucs := &UserCallState{
		CallState:               callState,
		HasCompletedAudioFiles:  hasCompleteAudioFiles,
		HasSuccessfulTranscript: hasSuccessfulTrans,
		HasBrokenAudioFiles:     hasBrokenAudios,
	}

	// update latest call state info
	if err := b.UpdateUserLatestCallState(ctx, user.SupabaseID.String, user.ID); err != nil {
		return nil, fmt.Errorf("update user latest call state: %w", err)
	}

	return ucs, nil
}
