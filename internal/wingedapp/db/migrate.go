package db

import (
	"context"
	"testing"
	dbMigrate "wingedapp/pgtester/internal/db/migrate"
	dbMock "wingedapp/pgtester/internal/db/mock"
	wingedappMigration "wingedapp/pgtester/internal/wingedapp/migration"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

var (
	stmts = wingedappMigration.Stmts
)

// MigratedConn initializes the PostgreSQL database and runs migrations for the Winged application.
func MigratedConn(t *testing.T) (*sqlx.DB, func()) {
	t.Helper()

	db, connCfg, cleanup := dbMock.Pg(t)
	// we override this to use public
	connCfg.Schema = "public"

	opts := &dbMigrate.Options{
		Stmts:   stmts,
		ConnCfg: connCfg,
	}

	err := dbMigrate.PG(context.TODO(), opts)
	require.NoError(t, err, "expected no error running migrations")

	return db, cleanup
}
