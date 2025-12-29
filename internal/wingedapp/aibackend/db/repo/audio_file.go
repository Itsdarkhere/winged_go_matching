package repo

import (
	"context"
	"errors"
	"fmt"
	wingedappModel "wingedapp/pgtester/internal/wingedapp/business/domain/registration"

	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jackc/pgx/v4"
)

func qModAudioFiles(f *wingedappModel.AudioFileQueryFilter) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)

	if f.ID != "" {
		qMods = append(qMods, aipgmodel.AudioFileWhere.ID.EQ(f.ID))
	}
	if f.ConversationID.Valid {
		qMods = append(qMods, aipgmodel.AudioFileWhere.ConversationID.EQ(f.ConversationID))
	}
	if f.UserID.Valid {
		qMods = append(qMods, aipgmodel.AudioFileWhere.UserID.EQ(f.UserID.String))
	}
	if len(f.Categories) > 0 {
		qMods = append(qMods, aipgmodel.AudioFileWhere.Category.IN(f.Categories))
	}
	if f.Limit.Valid && f.Limit.Int > 0 {
		qMods = append(qMods, qm.Limit(f.Limit.Int))
	}
	if f.HasStorage.Bool && f.HasStorage.Valid {
		qMods = append(qMods, aipgmodel.AudioFileWhere.StoragePath.NEQ(null.StringFrom("")))
	}

	return qMods
}

func (s *Store) AudioFiles(ctx context.Context, exec boil.ContextExecutor, filter *wingedappModel.AudioFileQueryFilter) (aipgmodel.AudioFileSlice, error) {
	if filter == nil {
		filter = &wingedappModel.AudioFileQueryFilter{}
	}

	audioFiles, err := aipgmodel.AudioFiles(qModAudioFiles(filter)...).All(ctx, exec)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return aipgmodel.AudioFileSlice{}, nil
		}
		return nil, fmt.Errorf("query audio files: %w", err)
	}

	return audioFiles, err
}

func (s *Store) AudioFile(ctx context.Context, exec boil.ContextExecutor, filter *wingedappModel.AudioFileQueryFilter) (*aipgmodel.AudioFile, error) {
	audioFile, err := s.AudioFiles(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("audio file: %w", err)
	}

	if len(audioFile) == 0 {
		return nil, fmt.Errorf("there is no audio file")
	}
	if len(audioFile) != 1 {
		return nil, fmt.Errorf("audio file count mismatch, have %d, want 1", len(audioFile))
	}

	return audioFile[0], nil
}
