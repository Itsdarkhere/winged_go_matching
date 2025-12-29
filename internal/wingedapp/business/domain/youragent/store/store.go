package store

import (
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
)

type Store struct {
	logrus         applog.Logger
	repoBackendApp *repo.Store // (global) backend_app
}

// NewStore initializes the Store with log, winged repo, and ai backend repo.
func NewStore(log applog.Logger, winged *repo.Store) (*Store, error) {
	if log == nil {
		return nil, fmt.Errorf("logrus entry cannot be nil")
	}
	if winged == nil {
		return nil, fmt.Errorf("winged repo cannot be nil")
	}

	return &Store{
		logrus:         log,
		repoBackendApp: winged,
	}, nil
}
