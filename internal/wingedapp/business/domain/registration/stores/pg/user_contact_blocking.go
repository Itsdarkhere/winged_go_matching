package pg

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/business/domain/registration"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"

	"github.com/aarondl/sqlboiler/v4/boil"
)

func pgInsertUserBlockedContact(inserter *registration.InsertUserBlockedContact) *repo.InsertUserBlockedContact {
	if inserter == nil {
		return nil
	}

	return &repo.InsertUserBlockedContact{
		UserID: inserter.UserID,
		Number: inserter.Number,
	}
}

func (s *Store) InsertUserBlockedContact(ctx context.Context, db boil.ContextExecutor, inserter *registration.InsertUserBlockedContact) error {
	return s.repoBackendApp.InsertUserBlockedContact(ctx, db, pgInsertUserBlockedContact(inserter))
}

func pgUserBlockedContactQueryFilter(inserter *registration.UserBlockedContactQueryFilter) *repo.UserBlockedContactQueryFilter {
	return &repo.UserBlockedContactQueryFilter{
		ID:     inserter.ID,
		UserID: inserter.UserID,
	}
}

// UserBlockedContacts retrieves a list of blocked contacts based on the provided filter.
func (s *Store) UserBlockedContacts(ctx context.Context,
	exec boil.ContextExecutor,
	filter *registration.UserBlockedContactQueryFilter,
) ([]registration.UserBlockedContact, error) {
	pgUserBlockedContacts, err := s.repoBackendApp.UserBlockedContacts(ctx,
		exec,
		pgUserBlockedContactQueryFilter(filter),
	)
	if err != nil {
		return nil, fmt.Errorf("list blocked contacts: %w", err)
	}

	return newUserBlockedContactsFromSlice(pgUserBlockedContacts), nil
}

// UserBlockedContact retrieves details of a specific blocked contact based on the provided filter.
func (s *Store) UserBlockedContact(ctx context.Context,
	exec boil.ContextExecutor,
	filter *registration.UserBlockedContactQueryFilter,
) (*registration.UserBlockedContact, error) {
	pgUserBlockedContact, err := s.repoBackendApp.UserBlockedContactDetails(ctx,
		exec,
		pgUserBlockedContactQueryFilter(filter),
	)
	if err != nil {
		return nil, fmt.Errorf("list blocked contacts: %w", err)
	}

	if pgUserBlockedContact == nil {
		return nil, nil
	}

	blockedContact := &registration.UserBlockedContact{
		ID:            pgUserBlockedContact.ID,
		UserID:        pgUserBlockedContact.UserID,
		BlockedNumber: pgUserBlockedContact.BlockedNumber,
	}
	return blockedContact, nil
}

// UserUnblockContact removes a blocked contact for a user based on the provided filter.
func (s *Store) UserUnblockContact(ctx context.Context, db boil.ContextExecutor, filter *registration.UserBlockedContactQueryFilter) error {
	userBlockedContact, err := s.repoBackendApp.UserBlockedContact(ctx, db, &repo.UserBlockedContactQueryFilter{
		UserID:        filter.UserID,
		BlockedNumber: filter.BlockedNumber,
	})
	if err != nil {
		return fmt.Errorf("get blocked contact: %w", err)
	}
	return s.repoBackendApp.DeleteUserBlockedContact(ctx, db, userBlockedContact.ID)
}

// UserUnblockAll deletes multiple blocked contacts for a user.
func (s *Store) UserUnblockAll(ctx context.Context, db boil.ContextExecutor, ids []string) error {
	return s.repoBackendApp.UserUnblockAll(ctx, db, ids)
}

func newUserBlockedContactsFromSlice(slice pgmodel.UserBlockedContactSlice) []registration.UserBlockedContact {
	users := make([]registration.UserBlockedContact, 0)

	for _, user := range slice {
		users = append(users, registration.UserBlockedContact{
			ID:            user.ID,
			UserID:        user.UserID,
			BlockedNumber: user.BlockedNumber,
		})
	}

	return users
}
