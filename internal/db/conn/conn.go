package conn

import (
	"fmt"
	"strings"
	"time"
	"wingedapp/pgtester/internal/util/validationlib"

	"github.com/aarondl/null/v8"
	_ "github.com/lib/pq"
)

type DbType string

const (
	Postgres DbType = "postgres"
	MySQL           = "mysql"

	Host     string = "HOST"
	User            = "USER"
	Pass            = "PASS"
	Port            = "PORT"
	Database        = "DATABASE"
	Schema          = "SCHEMA"
)

type Cfg struct {
	TimeoutDuration time.Duration
	Database        string `json:"database" validate:"required"`
	Host            string `json:"host" validate:"required"`
	Port            int    `json:"port" validate:"required"`
	User            string `json:"user" validate:"required"`
	Pass            string `json:"pass" validate:"required"`
	Schema          string `json:"schema"`
	SearchPath      string `json:"search_path"`

	MaxIdleConns    int      `json:"max_idle_conns"`
	MaxOpenConns    int      `json:"max_open_conns"`
	MaxConnLifetime null.Int `json:"max_conn_lifetime"`
}

func (c *Cfg) Clone() *Cfg {
	return &Cfg{
		TimeoutDuration: c.TimeoutDuration,
		Database:        c.Database,
		Host:            c.Host,
		Port:            c.Port,
		User:            c.User,
		Pass:            c.Pass,
		Schema:          c.Schema,
		SearchPath:      c.SearchPath,
		MaxIdleConns:    c.MaxIdleConns,
		MaxOpenConns:    c.MaxOpenConns,
		MaxConnLifetime: c.MaxConnLifetime,
	}
}

func (c *Cfg) Validate(includeDatabase bool) error {
	err := validationlib.Validate(c)
	if err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	if includeDatabase && strings.TrimSpace(c.Database) == "" {
		return fmt.Errorf("missing database")
	}
	return nil
}

// MySQLSQLBoilerTOML generates a SQLBoiler TOML configuration for MySQL
// based on the current connection configuration.
func (c *Cfg) MySQLSQLBoilerTOML(outputPath, pkgName string) string {
	if outputPath == "" {
		outputPath = "./internal/imapp/db/pgmodel"
	}
	if pkgName == "" {
		pkgName = "pgmodel"
	}

	return fmt.Sprintf(sqlboilerTomlMySQL, outputPath, pkgName, c.Host, c.Port, c.User, c.Pass, c.Database)
}
