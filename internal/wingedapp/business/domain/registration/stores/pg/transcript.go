package pg

import (
	"context"
	"encoding/json"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"
	"wingedapp/pgtester/internal/wingedapp/business/domain/registration"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// Transcripts retrieves a list of user audio files based on the provided filter.
func (s *Store) Transcripts(ctx context.Context, execAI boil.ContextExecutor,
	filter *registration.TranscriptQueryFilters,
) ([]registration.UserTranscript, error) {
	pgUserTranscripts, err := s.repoAIBackend.Transcripts(ctx, execAI, filter)
	if err != nil {
		return nil, fmt.Errorf("list transcripts: %w", err)
	}

	return newUserTranscriptFromSlice(pgUserTranscripts), nil
}

// Transcript gets details of a specific user audio file based on the provided filter.
func (s *Store) Transcript(ctx context.Context,
	exec boil.ContextExecutor,
	filters *registration.TranscriptQueryFilters,
) (*registration.UserTranscript, error) {
	// format filters
	for i, filter := range filters.OrderedBys {
		if filter == "-created_at" {
			filters.OrderedBys[i] = "created_at DESC"
		}
	}

	// fetch transcripts from repo
	transcripts, err := s.repoAIBackend.Transcripts(ctx, exec, filters)
	if err != nil {
		return nil, fmt.Errorf("list transcript: %w", err)
	}
	if len(transcripts) == 0 {
		return nil, registration.ErrUserTranscriptNotFound
	}
	if len(transcripts) != 1 {
		return nil, fmt.Errorf("transcript count mismatch, have %d, want 1", len(transcripts))
	}

	newUserTranscript := newUserTranscriptFromSlice(transcripts)

	return &newUserTranscript[0], nil
}

func newUserTranscriptFromSlice(pgUserTranscripts aipgmodel.TranscriptSlice) []registration.UserTranscript {
	if pgUserTranscripts == nil {
		return nil
	}

	transcripts := make([]registration.UserTranscript, 0, len(pgUserTranscripts))
	for _, pgUserTranscript := range pgUserTranscripts {
		var transcriptData []registration.Transcript
		// attempt to unmarshal the JSON data into the Transcript struct
		_ = json.Unmarshal(pgUserTranscript.TranscriptData.JSON, &transcriptData)

		transcripts = append(transcripts, registration.UserTranscript{
			ID:             pgUserTranscript.ID,
			UserID:         pgUserTranscript.UserID,
			ConversationID: pgUserTranscript.ConversationID.String,
			Status:         pgUserTranscript.Status.String,
			CallSuccessful: pgUserTranscript.CallSuccessful.String,
			TranscriptData: transcriptData,
			CreatedAt:      pgUserTranscript.CreatedAt,
		})
	}

	return transcripts
}
