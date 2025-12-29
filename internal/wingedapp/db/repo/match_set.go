package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/types"
)

type InsertMatchSet struct {
	Name                  string
	NumberOfParticipants  int
	MatchingConfiguration json.RawMessage // current config this set is run against
}

type QueryFilterMatchSet struct {
	Key null.String
}

func (s *Store) InsertMatchSet(
	ctx context.Context,
	exec boil.ContextExecutor,
	inserter *InsertMatchSet,
) (*pgmodel.MatchSet, error) {
	if inserter.NumberOfParticipants <= 0 {
		return nil, fmt.Errorf("number_of_participants must be greater than 0")
	}
	if len(inserter.MatchingConfiguration) == 0 {
		return nil, fmt.Errorf("matching_parameters is required")
	}

	matchSet := pgmodel.MatchSet{
		Name:                 inserter.Name,
		NumberOfParticipants: inserter.NumberOfParticipants,
		MatchConfiguration:   types.JSON(inserter.MatchingConfiguration),
	}

	if err := matchSet.Insert(ctx, exec, boil.Infer()); err != nil {
		return nil, fmt.Errorf("insert match set: %w", err)
	}

	return &matchSet, nil
}
