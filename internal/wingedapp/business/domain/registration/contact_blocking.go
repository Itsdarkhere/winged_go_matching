package registration

import (
	"context"
	"fmt"

	"github.com/aarondl/null/v8"
)

// BlockContact blocks a contact by inserting a new record into the store.
func (b *Business) BlockContact(ctx context.Context, userID string, contactNumber string) error {
	inserter := &InsertUserBlockedContact{
		UserID: userID,
		Number: contactNumber,
	}

	if err := b.storer.InsertUserBlockedContact(ctx, b.transBE.DB(), inserter); err != nil {
		return fmt.Errorf("insert user blocked contact: %w", err)
	}

	return nil
}

func (b *Business) BlockedContacts(ctx context.Context, userId string) ([]UserBlockedContact, error) {
	filter := UserBlockedContactQueryFilter{
		UserID: null.StringFrom(userId),
	}

	userBlockedContacts, err := b.storer.UserBlockedContacts(ctx, b.transBE.DB(), &filter)
	if err != nil {
		return nil, fmt.Errorf("list blocked contacts: %w", err)
	}

	return userBlockedContacts, nil
}

func (b *Business) UserBlockedContactDetails(ctx context.Context, userId string, contactNumber string) (*UserBlockedContact, error) {
	filter := UserBlockedContactQueryFilter{
		UserID:        null.StringFrom(userId),
		BlockedNumber: null.StringFrom(contactNumber),
	}

	userBlockedContact, err := b.storer.UserBlockedContact(ctx, b.transBE.DB(), &filter)
	if err != nil {
		return nil, fmt.Errorf("get blocked contact: %w", err)
	}
	return userBlockedContact, nil

}

func (b *Business) UserUnblockContact(ctx context.Context, userId string, contactNumber string) error {
	f := UserBlockedContactQueryFilter{
		UserID:        null.StringFrom(userId),
		BlockedNumber: null.StringFrom(contactNumber),
	}

	return b.storer.UserUnblockContact(ctx, b.transBE.DB(), &f)

}

func (b *Business) UserUnblockAll(ctx context.Context, userId string) error {
	tx, err := b.transBE.TX()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer b.transBE.Rollback(tx)

	filter := UserBlockedContactQueryFilter{
		UserID: null.StringFrom(userId),
	}
	blockedContacts, err := b.storer.UserBlockedContacts(ctx, tx, &filter)
	if err != nil {
		return fmt.Errorf("failed to fetch blocked contacts: %w", err)
	}

	var blockedIDs []string
	for _, blockedContact := range blockedContacts {
		blockedIDs = append(blockedIDs, blockedContact.ID)
	}

	if err = b.storer.UserUnblockAll(ctx, tx, blockedIDs); err != nil {
		return fmt.Errorf("failed to unblock contacts: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
