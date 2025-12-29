package mock

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"testing"
	dbConn "wingedapp/pgtester/internal/db/conn"
)

func MySQL(t *testing.T, cfg *dbConn.Cfg) (*sqlx.DB, *dbConn.Cfg, func()) {
	require.NotNil(t, cfg, "expected config to be not nil")

	var db *sqlx.DB

	h := &handler{
		t:      t,
		cfg:    cfg,
		driver: dbConn.MySQL,
		db:     db,
	}

	db = h.mySQLConn()
	db, cleanup := h.mySQLMockDb(db)

	return db, cfg, cleanup
}

func (h *handler) mySQLMockDb(db *sqlx.DB) (*sqlx.DB, func()) {
	t := h.t
	t.Helper()

	h.guid = fmt.Sprintf("a_test_db_%s", guid())
	h.cfg.Database = h.guid

	_, err := db.Exec(createStmtMySQL(h.guid))
	require.NoError(t, err, "expected no error creating mock database")

	testDb, err := h.cfg.MySQLConn(context.TODO(), false)
	require.NoError(t, err, "expected no error connecting to guid MYSQL")
	require.NotNil(t, testDb, "expected testDb to be not nil")

	h.db = db
	h.testDb = testDb

	return h.testDb, h.cleanupFn()
}

func (h *handler) mySQLConn() *sqlx.DB {
	db, err := h.cfg.MySQLConn(context.TODO(), false)
	require.NoError(h.t, err, "expected no error connecting to initial MySQL")
	require.NotNil(h.t, db, "expected db to be not nil")

	return db
}
