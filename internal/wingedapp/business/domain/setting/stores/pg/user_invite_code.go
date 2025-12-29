package pg

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/business/domain/setting"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

func toBizUserInviteCode(pgInviteCode *repo.UserInviteCode) setting.UserInviteCode {
	return setting.UserInviteCode{
		ID:             pgInviteCode.ID,
		ForNumber:      pgInviteCode.ForNumber.String,
		ForNumberHash:  pgInviteCode.ForNumberHash.String,
		Category:       pgInviteCode.InviteCodeType,
		CreatedAt:      pgInviteCode.CreatedAt,
		Code:           pgInviteCode.InviteCode,
		UsageCount:     pgInviteCode.UsageCount,
		ReferralSource: pgInviteCode.ReferralSource,
		LastUsed:       pgInviteCode.LastUsed,
	}
}

func toBizUserInviteCodes(repoUserInviteCodes []repo.UserInviteCode) []setting.UserInviteCode {
	bizUserInviteCodes := make([]setting.UserInviteCode, 0)
	for _, r := range repoUserInviteCodes {
		bizUserInviteCodes = append(bizUserInviteCodes, toBizUserInviteCode(&r))
	}
	return bizUserInviteCodes
}

func toRepoQueryFilterUserInviteCode(f *setting.QueryFilterUserInviteCode) *repo.UserInviteCodeQueryFilter {
	return &repo.UserInviteCodeQueryFilter{
		Code:      f.InviteCode,
		ForNumber: f.Number,
	}
}

func toBizUser(p *pgmodel.User) setting.User {
	return setting.User{
		ID: p.ID,
	}
}

func (s *Store) InsertUserInviteCode(ctx context.Context,
	exec boil.ContextExecutor,
	inserter *setting.InsertUserInviteCode,
) error {
	repoInserter := &repo.InsertUserInviteCode{
		InviteCode:         inserter.InviteCode,
		ReferralSource:     categoryUserInviteCodeReferral,
		InviteCodeType:     categoryUserInviteCodeReferral, // String enum value
		ForNumber:          null.StringFrom(inserter.ExtNumber),
		ForNumberHash:      null.StringFrom(inserter.ExtNumberHash),
		ReferrerNumberHash: null.StringFrom(inserter.ReferrerNumberHash),
	}

	if _, err := s.repoBackendApp.InsertUserInviteCode(ctx, exec, repoInserter); err != nil {
		return fmt.Errorf("insert user invite code: %w", err)
	}

	return nil
}

func (s *Store) UserInviteCodes(ctx context.Context,
	exec boil.ContextExecutor,
	f *setting.QueryFilterUserInviteCode,
) ([]setting.UserInviteCode, error) {
	i, err := s.repoBackendApp.UserInviteCodes(ctx, exec, toRepoQueryFilterUserInviteCode(f))
	if err != nil {
		return nil, fmt.Errorf("invite codes: %w", err)
	}
	return toBizUserInviteCodes(i), nil
}

func (s *Store) DeleteUserInviteCode(ctx context.Context, exec boil.ContextExecutor, id string) (int64, error) {
	return s.repoBackendApp.DeleteInviteCode(ctx, exec, id)
}
