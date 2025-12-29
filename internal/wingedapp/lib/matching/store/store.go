package store

import (
	aiRepo "wingedapp/pgtester/internal/wingedapp/aibackend/db/repo"
	beRepo "wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
)

type MatchingStores struct {
	UserStore             *UserStore
	UserDatingPrefsStore  *UserDatingPrefsStore
	MatchSetStore         *MatchSetStore
	MatchConfigStore      *ConfigStore
	MatchResultStore      *MatchResultStore
	ProfileStore          *ProfileStore
	SupabaseStore         *SupabaseStore
	UserMatchActionsStore *UserMatchActionsStore
	AudioStore            *AudioStore
	LovestoryStore        *LovestoryStore
	DateInstanceStore     *DateInstanceStore
}

// NewMatchingStores creates a new instance of MatchingStores with the provided logger.
func NewMatchingStores(l applog.Logger) *MatchingStores {
	r := &beRepo.Store{} // no need for this to come from params
	a := &aiRepo.Store{} // ai backend repo

	return &MatchingStores{
		UserStore:             &UserStore{l, r},
		UserDatingPrefsStore:  &UserDatingPrefsStore{l, r},
		MatchSetStore:         &MatchSetStore{l, r},
		MatchConfigStore:      &ConfigStore{l, r},
		MatchResultStore:      &MatchResultStore{l, r},
		ProfileStore:          &ProfileStore{l, a},
		SupabaseStore:         NewSupabaseStore(l),
		UserMatchActionsStore: NewUserMatchActionsStore(l),
		AudioStore:            NewAudioStore(l),
		LovestoryStore:        NewLovestoryStore(l),
		DateInstanceStore:     &DateInstanceStore{l, r},
	}
}
