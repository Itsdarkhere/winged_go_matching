package mock

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"wingedapp/pgtester/internal/db/conn"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func viperWithPrefix(t *testing.T, p string) *viper.Viper {
	t.Helper()
	v := viper.New()
	v.SetEnvPrefix(p)
	v.AutomaticEnv()

	// Please loopp through here
	keys := v.AllKeys()
	for _, key := range keys {
		fmt.Println("=== key:", key, "value:", v.GetString(key))
	}

	return v
}

// guid generates a new UUID and replaces hyphens with underscores.
func guid() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "_")
}

func createStmtPG(guid string) string {
	return fmt.Sprintf(`CREATE DATABASE "%s"`, guid)
}

func createStmtMySQL(db string) string {
	return fmt.Sprintf("CREATE DATABASE %s", db)
}

func dropStmtMySQL(guid string, dbType conn.DbType) string {
	switch dbType {
	case conn.Postgres:
		return fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, guid)
	default:
		return fmt.Sprintf(`DROP DATABASE "%s"`, guid)
	}
}

func dropStmtPG(guid string, dbType conn.DbType) string {
	switch dbType {
	case conn.Postgres:
		return fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, guid)
	default:
		return fmt.Sprintf(`DROP DATABASE "%s"`, guid)
	}
}

func NewTestConfig(t *testing.T, viperPrefix string) *conn.Cfg {
	v := viperWithPrefix(t, viperPrefix)

	port, err := strconv.Atoi(v.GetString(conn.Port))
	require.NoError(t, err, "expected no error converting port to int")

	return &conn.Cfg{
		TimeoutDuration: 10,
		Database:        v.GetString(conn.Database),
		Host:            v.GetString(conn.Host),
		Port:            port,
		User:            v.GetString(conn.User),
		Pass:            v.GetString(conn.Pass),
		Schema:          v.GetString(conn.Schema),
	}
}

func SkipWingedPG(t *testing.T) {
	t.Helper()
	v := viperWithPrefix(t, "WINGED_PG")
	if v.GetString("HOST") == "" {
		t.Skip("WINGED_PG prefix not set, skipping test")
	}
}
