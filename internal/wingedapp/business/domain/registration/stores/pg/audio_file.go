package pg

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"
	"wingedapp/pgtester/internal/wingedapp/business/domain/registration"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// AudioFiles retrieves a list of user audio files based on the provided filter.
func (s *Store) AudioFiles(ctx context.Context,
	exec boil.ContextExecutor,
	filter *registration.AudioFileQueryFilter,
) ([]registration.UserAudio, error) {
	pgUserAudioFiles, err := s.repoAIBackend.AudioFiles(ctx,
		exec,
		filter,
	)
	if err != nil {
		return nil, fmt.Errorf("list audio files: %w", err)
	}

	return newUserAudioFileFromSlice(pgUserAudioFiles), nil
}

// AudioFile gets details of a specific user audio file based on the provided filter.
func (s *Store) AudioFile(ctx context.Context,
	exec boil.ContextExecutor,
	filter *registration.AudioFileQueryFilter,
) (*registration.UserAudio, error) {
	audioFiles, err := s.repoAIBackend.AudioFiles(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("list user audio files: %w", err)
	}
	if len(audioFiles) == 0 {
		return nil, registration.ErrUserAudioFileNotFound
	}
	if len(audioFiles) != 1 {
		return nil, fmt.Errorf("user audio file count mismatch, have %d, want 1", len(audioFiles))
	}

	newUserAudiFile := newUserAudioFileFromSlice(audioFiles)

	return &newUserAudiFile[0], nil
}

func newUserAudioFileFromSlice(pgUserAudioFiles aipgmodel.AudioFileSlice) []registration.UserAudio {
	if pgUserAudioFiles == nil {
		return nil
	}

	audioFiles := make([]registration.UserAudio, 0, len(pgUserAudioFiles))
	for _, pgUserAudioFile := range pgUserAudioFiles {
		audioFiles = append(audioFiles, registration.UserAudio{
			ID:             pgUserAudioFile.ID,
			Category:       pgUserAudioFile.Category.String,
			UserID:         pgUserAudioFile.UserID,
			ConversationID: pgUserAudioFile.ConversationID.String,
			StoragePath:    pgUserAudioFile.StoragePath.String,
		})
	}

	return audioFiles
}
