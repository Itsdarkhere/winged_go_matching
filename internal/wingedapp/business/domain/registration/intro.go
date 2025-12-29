package registration

import (
	"context"
	"fmt"

	"github.com/aarondl/null/v8"
)

// UpdateUserIntro sets a user's selected voice clone introduction.
func (b *Business) UpdateUserIntro(ctx context.Context, userID, supabaseUserID, audioFileID string) error {
	existingAudioFile, err := b.storer.AudioFile(ctx, b.dbAI(), &AudioFileQueryFilter{
		ID:     audioFileID,
		UserID: null.StringFrom(supabaseUserID),
	})
	if err != nil {
		return fmt.Errorf("existingAudioFile: %w", err)
	}

	tx, err := b.transBE.TX()
	if err != nil {
		return fmt.Errorf("transback backend transaction: %w", err)
	}
	defer b.transBE.Rollback(tx)

	if _, err = b.storer.UpdateUser(ctx, tx, b.dbAI(), &UpdateUser{
		ID:              userID,
		SelectedVoiceID: null.StringFrom(existingAudioFile.ID),
	}); err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
