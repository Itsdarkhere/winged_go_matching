package store

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/economy"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

type UserStore struct {
	logger applog.Logger
	repo   *repo.Store
}

func NewUserStore(logger applog.Logger, repo *repo.Store) *UserStore {
	return &UserStore{logger: logger, repo: repo}
}

func (s *UserStore) Exists(
	ctx context.Context,
	exec boil.ContextExecutor,
	uuid string,
) (bool, error) {
	if _, err := s.repo.User(ctx, exec, &repo.QueryFilterUser{
		ID: null.StringFrom(uuid),
	}); err != nil {
		return false, fmt.Errorf("checking user exists: %w", err)
	}

	return true, nil
}

func (s *UserStore) User(
	ctx context.Context,
	exec boil.ContextExecutor,
	filter *economy.QueryFilterUser,
) (*economy.User, error) {
	pgUser, err := s.repo.User(ctx, exec, &repo.QueryFilterUser{
		ID:         filter.ID,
		Sha256Hash: filter.Sha256Hash,
	})
	if err != nil {
		return nil, fmt.Errorf("user query: %w", err)
	}
	if pgUser == nil || pgUser.ID == "" {
		return nil, nil
	}

	return &economy.User{
		ID:                  pgUser.ID,
		UserInviteCodeRefID: pgUser.UserInviteCodeRefID,
	}, nil
}
