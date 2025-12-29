package contact

import (
	"context"
)

// List represents the contact list domain.
type List struct {
}

// BlockedUsers retrieves the list of blocked users.
func (l *List) BlockedUsers(ctx context.Context) error {
	return nil
}

// BlockUser adds a user to the block list.
func (l *List) BlockUser(ctx context.Context) error {
	return nil
}

// UnblockUser removes a user from the block list.
func (l *List) UnblockUser(ctx context.Context) error {
	return nil
}

// UnblockAllUsers removes all users from the block list.
func (l *List) UnblockAllUsers(ctx context.Context) error {
	return nil
}
