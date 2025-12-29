package repo

import (
	"context"
	"errors"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jackc/pgx/v4"
)

type UpsertUserElevenLabs struct {
	UserID       string `json:"email"`
	Conversation []byte `json:"conversation"`
}

// InsertUserElevenLabs inserts an eleven labs for a user.
func (s *Store) InsertUserElevenLabs(ctx context.Context, db boil.ContextExecutor, upsert *UpsertUserElevenLabs) error {
	if upsert.UserID == "" {
		return fmt.Errorf("user UserID is required for inserting userElevenLabs")
	}

	userElevenLabs := pgmodel.UserElevenLab{
		UserID:       upsert.UserID,
		Conversation: upsert.Conversation,
	}

	if err := userElevenLabs.Insert(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("insert userElevenLabs: %w", err)
	}

	return nil
}

type UserElevenLabsQueryFilter struct {
	ID     null.String `json:"id"`
	UserID null.String `json:"user_id"`
}

func userElevenLabsFilter(filter *UserElevenLabsQueryFilter) []qm.QueryMod {
	filters := make([]qm.QueryMod, 0)
	if filter.ID.Valid {
		filters = append(filters, pgmodel.UserElevenLabWhere.ID.EQ(filter.ID.String))
	}
	if filter.UserID.Valid {
		filters = append(filters, pgmodel.UserElevenLabWhere.ID.EQ(filter.UserID.String))
	}

	return filters
}

// UserElevenLabs lists all the userElevenLabs based on the filter.
func (s *Store) UserElevenLabs(ctx context.Context,
	exec boil.ContextExecutor,
	filter *UserElevenLabsQueryFilter,
) (pgmodel.UserElevenLabSlice, error) {
	userElevenLabs, err := pgmodel.UserElevenLabs(userElevenLabsFilter(filter)...).All(ctx, exec)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgmodel.UserElevenLabSlice{}, nil // no results found
		}
		return nil, fmt.Errorf("query userElevenLabs: %w", err)
	}

	return userElevenLabs, nil
}

// UserElevenLab returns a single userElevenLabs based on the filter.
func (s *Store) UserElevenLab(ctx context.Context,
	exec boil.ContextExecutor,
	filter *UserElevenLabsQueryFilter,
) (*pgmodel.UserElevenLab, error) {
	usersElevenLabs, err := s.UserElevenLabs(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("user eleven labs: %w", err)
	}

	if len(usersElevenLabs) == 0 {
		return &pgmodel.UserElevenLab{}, nil
	}

	if len(usersElevenLabs) != 1 {
		return nil, fmt.Errorf("user eleven labs count mismatch, have %d, want 1", len(usersElevenLabs))
	}

	return usersElevenLabs[0], nil
}
