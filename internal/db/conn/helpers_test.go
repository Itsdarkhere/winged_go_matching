package conn_test

import (
	"context"
	"fmt"
	"testing"
	dbConn "wingedapp/pgtester/internal/db/conn"
	"wingedapp/pgtester/internal/util/strutil"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/stretchr/testify/require"
)

func Test_Count(t *testing.T) {
	t.Helper()

	const (
		q = `SELECT * FROM general_ai_context`
	)

	tSuite := testsuite.New(t)
	db, cleanup := tSuite.BackendAppDB()
	defer cleanup()

	c, err := dbConn.Count(context.TODO(), db, q)
	require.NoError(t, err, "expected no error executing count query")
	require.Greater(t, c, 0, "expected a positive number of rows")
}

func Test_QueryArray(t *testing.T) {
	t.Helper()

	const (
		q = `SELECT * FROM general_ai_context`
	)
	tSuite := testsuite.New(t)
	db, cleanup := tSuite.BackendAppDB()
	defer cleanup()

	rows, err := dbConn.QueryArray(context.TODO(), db, q)
	require.NoError(t, err, "expected no error executing count query")
	require.NotEmpty(t, rows, "expected non-empty rows")

	fmt.Println("==== rows:", strutil.GetAsJson(rows))
}
