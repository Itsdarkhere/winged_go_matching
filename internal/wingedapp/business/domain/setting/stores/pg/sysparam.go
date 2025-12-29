package pg

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/business/domain/setting"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// toBizSysParams converts a slice of SysParams from the repository model to the business model.
func toBizSysParams(spRepo pgmodel.SysParamSlice) []setting.SysParam {
	spBiz := make([]setting.SysParam, 0)
	for _, s := range spRepo {
		spBiz = append(spBiz, setting.SysParam{
			Key: s.Key,
			Val: s.Val,
		})
	}
	return spBiz
}

func (s *Store) SysParams(ctx context.Context,
	exec boil.ContextExecutor,
) ([]setting.SysParam, error) {
	spRepo, err := s.repoBackendApp.SysParams(ctx, exec, &repo.SysParamQueryFilter{})
	if err != nil {
		return nil, fmt.Errorf("repo sysparams: %w", err)
	}
	return toBizSysParams(spRepo), nil
}
