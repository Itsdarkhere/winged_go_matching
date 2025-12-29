package registration

import (
	"context"
	"fmt"
)

func (b *Business) Transcript(ctx context.Context, id string) (*UserTranscript, error) {
	filter := TranscriptQueryFilters{
		ID: id,
	}

	transcript, err := b.storer.Transcript(ctx, b.dbAI(), &filter)
	if err != nil {
		return nil, fmt.Errorf("get audio file: %w", err)
	}

	return transcript, nil
}

// Transcripts returns all transcripts, need to add filter later,
// or remove this method all together.
func (b *Business) Transcripts(ctx context.Context) ([]UserTranscript, error) {
	transcripts, err := b.storer.Transcripts(ctx, b.dbAI(), &TranscriptQueryFilters{})
	if err != nil {
		return nil, fmt.Errorf("get audio files: %w", err)
	}

	return transcripts, nil
}
