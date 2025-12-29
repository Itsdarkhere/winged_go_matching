package pg

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/repo"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

// SysParam removes a blocked contact for a user based on the provided filter.
func (s *Store) SysParam(ctx context.Context, exec boil.ContextExecutor, key string) (string, error) {
	sp, err := s.repoBackendApp.SysParam(ctx, exec, &repo.SysParamQueryFilter{Key: null.StringFrom(key)})
	if err != nil {
		return "", fmt.Errorf("get sys param: %w", err)
	}
	return sp.Val, nil
}
