package repo

import (
	"context"
	"errors"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jackc/pgx/v4"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

type InsertUserBlockedContact struct {
	UserID string `json:"email"`
	Number string `json:"contact"`
}

// InsertUserBlockedContact inserts a blocked contact for a user.
func (s *Store) InsertUserBlockedContact(ctx context.Context, db boil.ContextExecutor, inserter *InsertUserBlockedContact) error {
	if inserter.UserID == "" {
		return fmt.Errorf("user UserID is required for inserting userBlockedContact")
	}

	userBlockedContact := pgmodel.UserBlockedContact{
		UserID:        inserter.UserID,
		BlockedNumber: null.StringFrom(inserter.Number),
	}

	if err := userBlockedContact.Insert(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("insert userBlockedContact: %w", err)
	}
	return nil
}

type UserBlockedContactQueryFilter struct {
	ID            null.String `json:"id"`
	UserID        null.String `json:"user_id"`
	BlockedNumber null.String `json:"blocked_number"`
}

func blockedUsersFilter(filter *UserBlockedContactQueryFilter) []qm.QueryMod {
	filters := make([]qm.QueryMod, 0)
	if filter.ID.Valid {
		filters = append(filters, pgmodel.UserBlockedContactWhere.ID.EQ(filter.ID.String))
	}
	if filter.UserID.Valid {
		filters = append(filters, pgmodel.UserBlockedContactWhere.UserID.EQ(filter.UserID.String))
	}
	if filter.BlockedNumber.Valid {
		filters = append(filters, pgmodel.UserBlockedContactWhere.BlockedNumber.EQ(filter.BlockedNumber))
	}

	return filters
}

// UserBlockedContacts lists all the blocked contacts based on the filter
func (s *Store) UserBlockedContacts(ctx context.Context,
	exec boil.ContextExecutor,
	filter *UserBlockedContactQueryFilter,
) (pgmodel.UserBlockedContactSlice, error) {
	blockedUsers, err := pgmodel.UserBlockedContacts(blockedUsersFilter(filter)...).All(ctx, exec)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgmodel.UserBlockedContactSlice{}, nil
		}
		return nil, fmt.Errorf("query blocked users: %w", err)
	}
	return blockedUsers, nil
}

// UserBlockedContactDetails returns blocked contact based on the filter
func (s *Store) UserBlockedContactDetails(ctx context.Context,
	exec boil.ContextExecutor,
	filter *UserBlockedContactQueryFilter,
) (*pgmodel.UserBlockedContact, error) {
	blockedContact, err := s.UserBlockedContacts(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("blocked contact: %w", err)
	}

	if len(blockedContact) == 0 {
		return &pgmodel.UserBlockedContact{}, nil
	}

	if len(blockedContact) != 1 {
		return nil, fmt.Errorf("blocked contact count mismatch, have %d, want 1", len(blockedContact))
	}

	return blockedContact[0], nil
}

// DeleteUserBlockedContact deletes a blocked contact for a user.
func (s *Store) DeleteUserBlockedContact(
	ctx context.Context,
	db boil.ContextExecutor,
	id string,
) error {
	blockedContact := pgmodel.UserBlockedContact{ID: id}
	count, err := blockedContact.Delete(ctx, db)
	if err != nil {
		return fmt.Errorf("delete blocked user: %w", err)
	}
	fmt.Println("=== count:", count)
	if count == 0 {
		return fmt.Errorf("delete blocked user: no rows affected")
	}

	return nil
}

// UserUnblockAll deletes multiple blocked contacts for a user.
func (s *Store) UserUnblockAll(ctx context.Context, db boil.ContextExecutor, ids []string) error {
	if len(ids) == 0 {
		return errors.New("no ids provided for bulk delete")
	}

	count, err := pgmodel.UserBlockedContacts(
		pgmodel.UserBlockedContactWhere.ID.IN(ids),
	).DeleteAll(ctx, db)
	if err != nil {
		return fmt.Errorf("bulk delete blocked contacts: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("bulk delete blocked contacts: no rows affected")
	}

	return nil
}

// UserBlockedContact returns one blocked contact based on the filter
func (s *Store) UserBlockedContact(ctx context.Context,
	exec boil.ContextExecutor,
	filter *UserBlockedContactQueryFilter,
) (*pgmodel.UserBlockedContact, error) {
	blockedUsers, err := s.UserBlockedContacts(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("blocked user: %w", err)
	}

	if len(blockedUsers) == 0 {
		return nil, fmt.Errorf("blocked user: none found")
	}

	if len(blockedUsers) != 1 {
		return nil, fmt.Errorf("blocked user count mismatch, have %d, want 1", len(blockedUsers))
	}

	return blockedUsers[0], nil
}
