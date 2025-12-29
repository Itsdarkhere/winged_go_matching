package conn_test

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/conn"
	"wingedapp/pgtester/internal/db/testpg"

	"github.com/stretchr/testify/require"
)

func Test_ConnectPG(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	containerCfg, err := testpg.GetContainerConfig(ctx)
	require.NoError(t, err, "expected no error getting test container config")

	c := &conn.Cfg{
		TimeoutDuration: 0,
		Database:        containerCfg.DBName,
		Host:            containerCfg.DBHost,
		Port:            containerCfg.DBPort,
		User:            containerCfg.DBUser,
		Pass:            containerCfg.DBPass,
		Schema:          containerCfg.DBSchema,
	}

	db, err := c.PGConn(ctx, false)
	require.NoError(t, err, "expected no error connecting to PostgreSQL")
	require.NotNil(t, db, "expected db to be not nil")
}
