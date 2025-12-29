package registration

import (
	"context"
	"fmt"
)

func (b *Business) AudioFile(ctx context.Context, id string) (*UserAudio, error) {
	filter := AudioFileQueryFilter{
		ID: id,
	}

	audioFile, err := b.storer.AudioFile(ctx, b.dbAI(), &filter)
	if err != nil {
		return nil, fmt.Errorf("get audio file: %w", err)
	}

	return audioFile, nil
}

func (b *Business) AudioFiles(ctx context.Context, filter *AudioFileQueryFilter) ([]UserAudio, error) {
	audioFiles, err := b.storer.AudioFiles(ctx, b.dbAI(), filter)
	if err != nil {
		return nil, fmt.Errorf("get audio files: %w", err)
	}

	return audioFiles, nil
}
