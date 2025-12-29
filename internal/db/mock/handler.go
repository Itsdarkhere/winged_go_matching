package mock

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"testing"
	"wingedapp/pgtester/internal/db/conn"
)

type handler struct {
	t      *testing.T
	cfg    *conn.Cfg
	driver conn.DbType
	db     *sqlx.DB
	testDb *sqlx.DB
	guid   string

	disablePGSSL bool
}

// cleanupFn removes the mock PostgreSQL database created for testing.
func (h *handler) cleanupFn() func() {
	return func() {
		h.t.Helper()

		err := h.testDb.Close()
		require.NoError(h.t, err, "expected no error closing 'test' mock database")

		_, err = h.db.Exec(dropStmt(h.guid, h.driver))
		require.NoError(h.t, err, "expected no error dropping mock database")

		err = h.db.Close()
		require.NoError(h.t, err, "expected no error closing 'main' database connection")
	}
}

// dropStmt generates a SQL statement to drop a database based on the provided GUID and driver type.
func dropStmt(guid string, driver conn.DbType) string {
	switch driver {
	case conn.Postgres:
		return fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, guid)
	case conn.MySQL:
		return fmt.Sprintf("DROP DATABASE IF EXISTS %s", guid)
	default:
		panic("unsupported driver " + string(driver))
	}
}
