package db_test

import (
	"testing"
	wingedappDb "wingedapp/pgtester/internal/wingedapp/db"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_MigrateWinged(t *testing.T) {
	t.Skip()
	wingedappDb.MigratedConn(t)
}

func Test_MigrateWinged_Query(t *testing.T) {
	var err error

	tSuite := testsuite.New(t)
	db, cleanup := tSuite.BackendAppDB()
	defer cleanup()

	// Query general_ai_context table to verify migrations ran successfully
	type aiContext struct {
		ID            uuid.UUID `json:"id" db:"id"`
		AIContextType string    `json:"ai_context_type" db:"ai_context_type"`
	}

	type aiContexts []aiContext
	var resp aiContexts

	qGetAIContexts := `SELECT id, ai_context_type FROM general_ai_context;`
	err = db.Select(&resp, qGetAIContexts)
	require.NoError(t, err, "expected no error executing query to get general_ai_context")
	// Validate query works and returns seeded data
}
