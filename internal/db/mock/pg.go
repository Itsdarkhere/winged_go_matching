package mock

import (
	"context"
	"fmt"
	"testing"
	dbConn "wingedapp/pgtester/internal/db/conn"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// pgConn initializes a PostgreSQL connection using the provided configuration.
func (h *handler) pgConn() *sqlx.DB {
	db, err := h.cfg.PGConn(context.TODO(), false)
	require.NoError(h.t, err, "expected no error connecting to initial PostgreSQL")
	require.NotNil(h.t, db, "expected db to be not nil")

	return db
}

// pgMockDb creates a mock PostgreSQL database for testing purposes.
func (h *handler) pgMockDb(db *sqlx.DB) (*sqlx.DB, func()) {
	t := h.t
	t.Helper()

	h.guid = fmt.Sprintf("a_test_db_%s", guid())
	h.cfg.Database = h.guid

	_, err := db.Exec(createStmtPG(h.guid))
	require.NoError(t, err, "expected no error creating mock database")

	testDb, err := h.cfg.PGConn(context.TODO(), false)
	require.NoError(t, err, "expected no error connecting to guid PostgreSQL")
	require.NotNil(t, testDb, "expected testDb to be not nil")

	h.db = db
	h.testDb = testDb

	return h.testDb, h.cleanupFn()
}

func Pg(t *testing.T) (*sqlx.DB, *dbConn.Cfg, func()) {
	SkipWingedPG(t)

	var db *sqlx.DB

	cfg := NewTestConfig(t, "WINGED_PG")
	h := &handler{
		t:            t,
		cfg:          cfg,
		driver:       dbConn.Postgres,
		db:           db,
		disablePGSSL: true,
	}

	db = h.pgConn()
	db, cleanup := h.pgMockDb(db)
	return db, cfg, cleanup
}
