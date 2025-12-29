package store

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/business/domain/youragent"
	"wingedapp/pgtester/internal/wingedapp/db/repo"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

func toPGInsertUserAIConvo(inserter *youragent.InsertUserAIConvo) *repo.InsertUserAIConvo {
	return &repo.InsertUserAIConvo{
		Response:          inserter.Response,
		Message:           inserter.Message,
		AdditionalContext: inserter.AdditionalContext,

		// refs
		PromptResponseID: inserter.PromptResponseID,
		UserID:           inserter.UserID,
		AiConvoType:      categoryYourAgent, // String enum value
	}
}

// UserAIConvos lists all the user AI conversations based on the filter
func (s *Store) UserAIConvos(ctx context.Context, db boil.ContextExecutor, f *youragent.UserAIConvoQueryFilter) ([]youragent.UserAIConvo, error) {
	u, err := s.repoBackendApp.UserAIConvos(ctx, db, toRepoUserAIConvoFilters(f))
	if err != nil {
		return nil, fmt.Errorf("query user ai convos: %w", err)
	}
	return toYourAgentUserAIConvos(u), nil
}

func (s *Store) UserAIConvo(ctx context.Context, db boil.ContextExecutor, f *youragent.UserAIConvoQueryFilter) (*youragent.UserAIConvo, error) {
	u, err := s.repoBackendApp.UserAIConvo(ctx, db, toRepoUserAIConvoFilters(f))
	if err != nil {
		return nil, fmt.Errorf("query user ai convo: %w", err)
	}
	return toYourAgentUserAIConvo(u), nil
}

func (s *Store) InsertUserAIConvo(ctx context.Context,
	db boil.ContextExecutor,
	inserter *youragent.InsertUserAIConvo,
) (*youragent.UserAIConvo, error) {
	inserted, err := s.repoBackendApp.InsertUserAIConvo(ctx, db, toPGInsertUserAIConvo(inserter))
	if err != nil {
		return nil, fmt.Errorf("insert user ai convo: %w", err)
	}

	userAIConvo, err := s.UserAIConvo(ctx, db, &youragent.UserAIConvoQueryFilter{
		ID: null.StringFrom(inserted.ID),
	})
	if err != nil {
		return nil, fmt.Errorf("fetching inserted user ai convo: %w", err)
	}

	return userAIConvo, nil
}

func toRepoUserAIConvoFilters(f *youragent.UserAIConvoQueryFilter) *repo.QueryFilterUserAIConvo {
	return &repo.QueryFilterUserAIConvo{
		ID:         f.ID,
		UserID:     f.UserID,
		Pagination: f.Pagination,
		OrderBy:    f.OrderBy,
		Sort:       f.Sort,
	}
}

func toYourAgentUserAIConvos(models []repo.UserAIConvo) []youragent.UserAIConvo {
	result := make([]youragent.UserAIConvo, 0, len(models))
	for _, model := range models {
		result = append(result, *toYourAgentUserAIConvo(&model))
	}
	return result
}

func toYourAgentUserAIConvo(model *repo.UserAIConvo) *youragent.UserAIConvo {
	return &youragent.UserAIConvo{
		ID:                model.ID,
		Role:              model.AiConvoType,
		UserID:            model.UserID,
		Message:           model.Message,
		AdditionalContext: model.AdditionalContext,
		Response:          model.Response,
		CreatedAt:         model.CreatedAt.Time,
	}
}
