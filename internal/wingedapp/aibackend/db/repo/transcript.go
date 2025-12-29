package repo

import (
	"context"
	"errors"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"
	wingedappModel "wingedapp/pgtester/internal/wingedapp/business/domain/registration"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jackc/pgx/v4"
)

func transcriptFilter(filter *wingedappModel.TranscriptQueryFilters) []qm.QueryMod {
	filters := make([]qm.QueryMod, 0)

	if filter.ID != "" {
		filters = append(filters, aipgmodel.TranscriptWhere.ID.EQ(filter.ID))
	}
	if filter.UserID.Valid {
		filters = append(filters, aipgmodel.TranscriptWhere.UserID.EQ(filter.UserID.String))
	}
	if filter.LatestTranscriptID.Valid {
		filters = append(filters, aipgmodel.TranscriptWhere.ID.EQ(filter.LatestTranscriptID.String))
	}
	if filter.Limit.Valid {
		filters = append(filters, qm.Limit(filter.Limit.Int))
	}
	if len(filter.OrderedBys) > 0 {
		for _, ob := range filter.OrderedBys {
			filters = append(filters, qm.OrderBy(ob))
		}
	}
	return filters
}

// Transcripts takes a list of transcripts
func (s *Store) Transcripts(ctx context.Context,
	exec boil.ContextExecutor,
	f *wingedappModel.TranscriptQueryFilters,
) (aipgmodel.TranscriptSlice, error) {
	if f == nil {
		f = &wingedappModel.TranscriptQueryFilters{}
	}
	transcripts, err := aipgmodel.Transcripts(transcriptFilter(f)...).All(ctx, exec)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return aipgmodel.TranscriptSlice{}, nil
		}
		return nil, fmt.Errorf("query profiles: %w", err)
	}
	return transcripts, nil
}

// Transcript returns a single transcript matching the filter.
func (s *Store) Transcript(
	ctx context.Context,
	exec boil.ContextExecutor,
	filter *wingedappModel.TranscriptQueryFilters,
) (*aipgmodel.Transcript, error) {
	transcripts, err := s.Transcripts(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("list of transcript: %w", err)
	}

	if len(transcripts) == 0 {
		return nil, nil // cleaner: no record found
	}

	if len(transcripts) != 1 {
		return nil, fmt.Errorf("profile count mismatch, have %d, want 1", len(transcripts))
	}

	return transcripts[0], nil
}
