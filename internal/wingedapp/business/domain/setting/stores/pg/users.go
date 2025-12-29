package pg

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/business/domain/setting"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"

	"github.com/aarondl/sqlboiler/v4/boil"
)

func toBizUsers(repoUsers pgmodel.UserSlice) []setting.User {
	bizUsers := make([]setting.User, 0)
	for _, user := range repoUsers {
		bizUsers = append(bizUsers, toBizUser(user))
	}
	return bizUsers
}

// qFilterRepoUsers converts a business layer QueryFilterUser to a repository layer QueryFilterUser.
// for now we only support filtering by ExtNumber.
func qFilterRepoUsers(f *setting.QueryFilterUser) *repo.QueryFilterUser {
	return &repo.QueryFilterUser{
		Number: f.Number,
	}
}

func (s *Store) Users(ctx context.Context,
	exec boil.ContextExecutor,
	filter *setting.QueryFilterUser,
) ([]setting.User, error) {
	repoUsers, err := s.repoBackendApp.Users(ctx, exec, qFilterRepoUsers(filter))
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}
	return toBizUsers(repoUsers), nil
}
