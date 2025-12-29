package store

import (
	"context"
	"fmt"

	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"
	supabaseRepo "wingedapp/pgtester/internal/wingedapp/supabase/db/repo"

	"github.com/aarondl/sqlboiler/v4/boil"
)

type SupabaseStore struct {
	l    applog.Logger
	repo *supabaseRepo.Store
}

func NewSupabaseStore(l applog.Logger) *SupabaseStore {
	return &SupabaseStore{
		l:    l,
		repo: supabaseRepo.NewStore(),
	}
}

// Insert inserts a user into supabase auth.users and returns the user ID.
func (s *SupabaseStore) Insert(
	ctx context.Context,
	exec boil.ContextExecutor,
	params *matching.InsertSupabaseUser,
) (string, error) {
	if params == nil {
		return "", fmt.Errorf("params cannot be nil")
	}

	userID, err := s.repo.Insert(ctx, exec, &supabaseRepo.InsertUserParams{
		Email: params.Email,
	})
	if err != nil {
		return "", fmt.Errorf("insert supabase auth user: %w", err)
	}

	return userID, nil
}

// DeleteByEmails deletes users by email addresses.
// Returns the number of users deleted.
func (s *SupabaseStore) DeleteByEmails(
	ctx context.Context,
	exec boil.ContextExecutor,
	emails []string,
) (int64, error) {
	deleted, err := s.repo.DeleteByEmails(ctx, exec, emails)
	if err != nil {
		return 0, fmt.Errorf("delete supabase users: %w", err)
	}

	s.l.Debug(ctx, "deleted supabase auth users by email", applog.F("deleted_count", deleted), applog.F("emails", emails))

	return deleted, nil
}
