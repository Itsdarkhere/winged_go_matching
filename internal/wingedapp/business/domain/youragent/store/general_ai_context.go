package store

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/business/domain/youragent"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"

	"github.com/aarondl/sqlboiler/v4/boil"
)

func toRepoGeneralAIContextFilters(f *youragent.GeneralAIContextQueryFilter) *repo.GeneralAIContextQueryFilter {
	return &repo.GeneralAIContextQueryFilter{
		ID: f.ID,
	}
}

func toYourAgentGeneralAIContext(model *pgmodel.GeneralAIContext) *youragent.GeneralAIContext {
	return &youragent.GeneralAIContext{
		ID:      model.ID,
		Context: model.Context,
	}
}

func toYourAgentGeneralAIContexts(models pgmodel.GeneralAIContextSlice) []youragent.GeneralAIContext {
	result := make([]youragent.GeneralAIContext, 0, len(models))
	for _, model := range models {
		result = append(result, *toYourAgentGeneralAIContext(model))
	}
	return result
}

func (s *Store) GeneralAIContexts(ctx context.Context,
	db boil.ContextExecutor,
	f *youragent.GeneralAIContextQueryFilter,
) ([]youragent.GeneralAIContext, error) {
	u, err := s.repoBackendApp.GeneralAIContexts(ctx, db, toRepoGeneralAIContextFilters(f))
	if err != nil {
		return nil, fmt.Errorf("query general AI context: %w", err)
	}
	return toYourAgentGeneralAIContexts(u), nil
}
