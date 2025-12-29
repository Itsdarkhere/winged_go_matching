package pg

import (
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
)

type Store struct {
	logrus         applog.Logger
	repoBackendApp *repo.Store // global winged repo
}

func NewStore(log applog.Logger) (*Store, error) {
	if log == nil {
		return nil, fmt.Errorf("logrus entry cannot be nil")
	}

	s := &Store{
		logrus: log,
	}

	return s, nil
}
