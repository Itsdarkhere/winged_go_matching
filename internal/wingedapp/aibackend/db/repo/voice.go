package repo

import (
	"context"
	"errors"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jackc/pgx/v4"
)

type VoiceQueryFilter struct {
	ID     null.String `json:"id"`
	UserID null.String `json:"user_id"`
}

func voicesFiter(f *VoiceQueryFilter) []qm.QueryMod {
	qMod := make([]qm.QueryMod, 0)
	if f.ID.Valid {
		qMod = append(qMod, aipgmodel.TranscriptWhere.ID.EQ(f.ID.String))
	}
	if f.UserID.Valid {
		qMod = append(qMod, aipgmodel.TranscriptWhere.UserID.EQ(f.UserID.String))
	}
	return qMod
}

func (s *Store) Voices(ctx context.Context,
	exec boil.ContextExecutor,
	f *VoiceQueryFilter,
) (aipgmodel.TranscriptSlice, error) {
	transcripts, err := aipgmodel.Transcripts(voicesFiter(f)...).All(ctx, exec)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return aipgmodel.TranscriptSlice{}, nil
		}
		return nil, fmt.Errorf("query profiles: %w", err)
	}
	return transcripts, nil
}
