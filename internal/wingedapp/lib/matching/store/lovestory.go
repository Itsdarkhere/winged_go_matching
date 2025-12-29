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

// LovestoryStore fetches lovestory data from ai_backend.lovestory table.
//
// Lovestory Generation Flow:
// 1. Admin approves a match in the admin tool
// 2. Backend calls POST /lovestory_by_user_ids with:
//   - romeo_user_id, juliet_user_id (required)
//   - romeo_voice_id, juliet_voice_id (optional - fetched from audio_files.voice_id where category='intro')
//
// 3. AI backend generates a "first date simulation" audio and returns:
//   - audio_storage_url: S3 URL for the generated lovestory audio
//   - romeo_voice_id, juliet_voice_id: the voice IDs used (can be null if user has no intro audio)
//   - script: the lovestory conversation script (JSON)
//
// 4. Response is stored in lovestory table:
//   - first_date_simulation_audio: direct S3 URL (no audio_files lookup needed for retrieval)
//   - first_date_simulation_script: the conversation script JSON
//   - user_a_ref_id, user_b_ref_id: the two users in the match
//
// Note: voice_id lookup happens BEFORE calling the lovestory API (to pass voice IDs to generation).
// The first_date_simulation_audio field stores the final S3 URL directly - no indirection through audio_files.
type LovestoryStore struct {
	l applog.Logger
}

// NewLovestoryStore creates a new LovestoryStore.
func NewLovestoryStore(l applog.Logger) *LovestoryStore {
	return &LovestoryStore{l: l}
}

// LovestoryURLsByMatchIDs fetches lovestory audio URLs for multiple match result IDs in a batch.
// Returns a map of match_result_id -> first_date_simulation_audio URL.
//
// The first_date_simulation_audio field contains a direct S3 URL (not a reference to audio_files).
// This URL is populated when the lovestory is generated via POST /lovestory_by_user_ids.
// Returns only matches that have a non-null audio URL (lovestory generation completed).
func (s *LovestoryStore) LovestoryURLsByMatchIDs(ctx context.Context, exec boil.ContextExecutor, matchResultIDs []uuid.UUID) (map[uuid.UUID]null.String, error) {
	result := make(map[uuid.UUID]null.String)

	if len(matchResultIDs) == 0 {
		return result, nil
	}

	// Convert UUIDs to strings for query
	matchIDStrings := make([]string, len(matchResultIDs))
	for i, id := range matchResultIDs {
		matchIDStrings[i] = id.String()
	}

	lovestories, err := aipgmodel.Lovestories(
		aipgmodel.LovestoryWhere.MatchResultRefID.IN(matchIDStrings),
		aipgmodel.LovestoryWhere.FirstDateSimulationAudio.IsNotNull(),
	).All(ctx, exec)

	if err != nil {
		return nil, fmt.Errorf("fetch lovestory audio: %w", err)
	}

	// Map results by match_result_ref_id
	for _, ls := range lovestories {
		matchUUID, err := uuid.Parse(ls.MatchResultRefID)
		if err != nil {
			continue
		}
		// Only store first found (in case of duplicates - shouldn't happen due to UNIQUE constraint)
		if _, exists := result[matchUUID]; !exists {
			result[matchUUID] = ls.FirstDateSimulationAudio
		}
	}

	return result, nil
}
