package store

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/economy"

	"github.com/aarondl/sqlboiler/v4/boil"
)

type InviteCodeStore struct {
	logger applog.Logger
	repo   *repo.Store
}

func NewInviteCodeStore(logger applog.Logger, repo *repo.Store) *InviteCodeStore {
	return &InviteCodeStore{logger: logger, repo: repo}
}

func (s *InviteCodeStore) InviteCode(ctx context.Context,
	exec boil.ContextExecutor,
	filter *economy.QueryFilterInviteCode,
) (*economy.InviteCode, error) {
	inviteCode, err := s.repo.UserInviteCode(ctx, exec, &repo.UserInviteCodeQueryFilter{
		ID: filter.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("invite code query: %w", err)
	}
	if inviteCode == nil || inviteCode.ID == "" {
		return nil, nil
	}

	return &economy.InviteCode{
		ID:                 inviteCode.ID,
		ReferrerNumberHash: inviteCode.ReferrerNumberHash,
	}, nil
}
