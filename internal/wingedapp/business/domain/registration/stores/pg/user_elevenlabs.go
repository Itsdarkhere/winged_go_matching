package pg

import (
	"context"
	"wingedapp/pgtester/internal/wingedapp/business/domain/registration"
	"wingedapp/pgtester/internal/wingedapp/db/repo"

	"github.com/aarondl/sqlboiler/v4/boil"
)

func pgUpsertUserElevenLabs(inserter *registration.InsertUserElevenLabs) *repo.UpsertUserElevenLabs {
	if inserter == nil {
		return nil
	}

	return &repo.UpsertUserElevenLabs{
		UserID:       inserter.UserID,
		Conversation: inserter.Conversation,
	}
}

func (s *Store) InsertUserElevenLabs(ctx context.Context, db boil.ContextExecutor, inserter *registration.InsertUserElevenLabs) error {
	return s.repoBackendApp.InsertUserElevenLabs(ctx, db, pgUpsertUserElevenLabs(inserter))
}
