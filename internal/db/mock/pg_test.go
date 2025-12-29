package mock_test

import (
	"testing"
	dbMock "wingedapp/pgtester/internal/db/mock"

	"github.com/stretchr/testify/require"
)

func TestPgMock(t *testing.T) {
	t.Skip()
	t.Helper()
	db, _, cleanup := dbMock.Pg(t)
	require.NotNil(t, db, "expected mock database to be initialized")
	require.NotNil(t, cleanup, "expected mock database to be initialized")
	defer cleanup()
}
