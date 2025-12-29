package store

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
)

// AudioStore fetches intro audio URLs from ai_backend.
type AudioStore struct {
	l applog.Logger
}

// NewAudioStore creates a new AudioStore.
func NewAudioStore(l applog.Logger) *AudioStore {
	return &AudioStore{l: l}
}

// AudioCategoryIntro is the category for intro audio files.
const AudioCategoryIntro = "intro"

// IntroAudioURL returns the intro audio storage_url for a user (by supabase_id).
// Returns null.String{} if no audio found.
func (s *AudioStore) IntroAudioURL(ctx context.Context, exec boil.ContextExecutor, supabaseUserID uuid.UUID) null.String {
	if supabaseUserID == uuid.Nil {
		return null.String{}
	}

	// Find the user's selected intro audio file
	// Category = "intro" and has storage_url
	audioFiles, err := aipgmodel.AudioFiles(
		aipgmodel.AudioFileWhere.UserID.EQ(supabaseUserID.String()),
		aipgmodel.AudioFileWhere.Category.EQ(null.StringFrom(AudioCategoryIntro)),
		aipgmodel.AudioFileWhere.StorageURL.IsNotNull(),
	).All(ctx, exec)

	if err != nil {
		s.l.Warn(ctx, "failed to fetch intro audio", applog.F("user_id", supabaseUserID), applog.F("error", err.Error()))
		return null.String{}
	}

	if len(audioFiles) == 0 {
		return null.String{}
	}

	// Return the first one's storage_url
	return audioFiles[0].StorageURL
}

// IntroAudioURLs fetches intro audio URLs for multiple users in a batch.
// Returns a map of supabase_user_id -> storage_url.
func (s *AudioStore) IntroAudioURLs(ctx context.Context, exec boil.ContextExecutor, supabaseUserIDs []uuid.UUID) (map[uuid.UUID]null.String, error) {
	result := make(map[uuid.UUID]null.String)

	if len(supabaseUserIDs) == 0 {
		return result, nil
	}

	// Convert UUIDs to strings for query
	userIDStrings := make([]string, len(supabaseUserIDs))
	for i, id := range supabaseUserIDs {
		userIDStrings[i] = id.String()
	}

	audioFiles, err := aipgmodel.AudioFiles(
		aipgmodel.AudioFileWhere.UserID.IN(userIDStrings),
		aipgmodel.AudioFileWhere.Category.EQ(null.StringFrom(AudioCategoryIntro)),
		aipgmodel.AudioFileWhere.StorageURL.IsNotNull(),
	).All(ctx, exec)

	if err != nil {
		return nil, fmt.Errorf("fetch intro audio files: %w", err)
	}

	// Map results
	for _, af := range audioFiles {
		userUUID, err := uuid.Parse(af.UserID)
		if err != nil {
			continue
		}
		// Only store first found (in case of duplicates)
		if _, exists := result[userUUID]; !exists {
			result[userUUID] = af.StorageURL
		}
	}

	return result, nil
}
