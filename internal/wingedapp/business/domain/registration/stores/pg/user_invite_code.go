package pg

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/business/domain/registration"
	"wingedapp/pgtester/internal/wingedapp/db/repo"

	"github.com/aarondl/sqlboiler/v4/boil"
)

func toBizUserInviteCode(pgInviteCode *repo.UserInviteCode) registration.UserInviteCode {
	return registration.UserInviteCode{
		ID:         pgInviteCode.ID,
		ForNumber:  pgInviteCode.ForNumber.String,
		Category:   pgInviteCode.InviteCodeType,
		CreatedAt:  pgInviteCode.CreatedAt,
		Code:       pgInviteCode.InviteCode,
		UsageCount: pgInviteCode.UsageCount,
		LastUsed:   pgInviteCode.LastUsed,
	}
}

func toBizUserInviteCodes(repoUserInviteCodes []repo.UserInviteCode) []registration.UserInviteCode {
	userInviteCodes := make([]registration.UserInviteCode, 0)
	for _, pgInviteCode := range repoUserInviteCodes {
		userInviteCodes = append(userInviteCodes, toBizUserInviteCode(&pgInviteCode))
	}
	return userInviteCodes
}

func toRepoQueryFilterUserInviteCode(filter *registration.UserInviteCodeQueryFilter) *repo.UserInviteCodeQueryFilter {
	return &repo.UserInviteCodeQueryFilter{
		Code: filter.Code,
	}
}

func (s *Store) UserInviteCodes(ctx context.Context,
	exec boil.ContextExecutor,
	filter *registration.UserInviteCodeQueryFilter,
) ([]registration.UserInviteCode, error) {
	inviteCodes, err := s.repoBackendApp.UserInviteCodes(ctx, exec, toRepoQueryFilterUserInviteCode(filter))
	if err != nil {
		return nil, fmt.Errorf("invite codes: %w", err)
	}
	return toBizUserInviteCodes(inviteCodes), nil
}

func (s *Store) DeleteUserInviteCode(ctx context.Context, exec boil.ContextExecutor, id string) (int64, error) {
	return s.repoBackendApp.DeleteInviteCode(ctx, exec, id)
}

func toPGUpdateUserInviteCode(updater *registration.UpdateUserInviteCode) *repo.UpdateUserInviteCode {
	return &repo.UpdateUserInviteCode{
		ID:         updater.ID,
		UsageCount: updater.UsageCount,
		LastUsed:   updater.LastUsed,
	}
}

// UpdateUserInviteCode updates an existing user invite code in the database.
func (s *Store) UpdateUserInviteCode(ctx context.Context,
	exec boil.ContextExecutor,
	updater *registration.UpdateUserInviteCode,
) error {
	if err := s.repoBackendApp.UpdateUserInviteCode(ctx, exec, toPGUpdateUserInviteCode(updater)); err != nil {
		return fmt.Errorf("updater user: %w", err)
	}
	return nil
}

// UserInviteCode retrieves a single invite code from the database.
func (s *Store) UserInviteCode(ctx context.Context,
	exec boil.ContextExecutor,
	filter *registration.UserInviteCodeQueryFilter,
) (*registration.UserInviteCode, error) {
	userInviteCodes, err := s.UserInviteCodes(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("invite: %w", err)
	}

	if len(userInviteCodes) == 0 {
		return nil, registration.ErrRegistrationCodeNotFound
	}

	if len(userInviteCodes) != 1 {
		return nil, fmt.Errorf("user invite codes count mismatch, have %d, want 1", len(userInviteCodes))
	}

	return &userInviteCodes[0], nil
}
