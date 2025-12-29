package store

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
)

type MessageStore struct {
	logger applog.Logger
	repo   *repo.Store
}

// Exists checks if a message with the given UUID exists
func (s *MessageStore) Exists(
	ctx context.Context,
	exec boil.ContextExecutor,
	uuid string,
) (bool, error) {
	if _, err := s.repo.UserAIConvo(ctx, exec, &repo.QueryFilterUserAIConvo{
		ID: null.StringFrom(uuid),
	}); err != nil {
		return false, fmt.Errorf("checking message exists: %w", err)
	}

	return true, nil
}
