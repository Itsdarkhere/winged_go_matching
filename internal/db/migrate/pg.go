package migrate

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strconv"
	"wingedapp/pgtester/internal/db/conn"

	_ "github.com/lib/pq"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

type Options struct {
	Stmts    embed.FS  `json:"stmts" validate:"required"`
	ConnCfg  *conn.Cfg `json:"conn_cfg" validate:"required"`
	connTmpl string
}

// PG initializes the PostgreSQL database, and runs migrations.
func PG(ctx context.Context, opts *Options) error {
	if opts == nil {
		return errors.New("no options provided")
	}

	connTmpl, err := conn.PGTmpl(&conn.PGConnParams{
		User:           opts.ConnCfg.User,
		Pass:           opts.ConnCfg.Pass,
		Host:           opts.ConnCfg.Host,
		Database:       opts.ConnCfg.Database,
		Schema:         opts.ConnCfg.Schema,
		Port:           opts.ConnCfg.Port,
		SearchPathBase: opts.ConnCfg.SearchPath,
	}, false)
	if err != nil {
		return fmt.Errorf("expected no error generating conn tmpl: %w", err)
	}

	// Perhaps I need to also create the schema if it doesn't exist?
	// But that would require connecting without a database first.
	// But do I do that at the top? hmmm or will it be easier if its just public?
	opts.connTmpl = connTmpl
	if err := migrate_(ctx, opts); err != nil {
		return fmt.Errorf("expected no error running migrations: %w", err)
	}

	return nil
}

func pgTMPL(cfg *conn.Cfg, excludeDb bool) string {
	var (
		driver = "postgres"
		user   = cfg.User
		pass   = cfg.Pass
		host   = cfg.Host
		port   = strconv.Itoa(cfg.Port)
		dbname = cfg.Database
	)

	// TODO: could be multiple schemas separated by comma
	if cfg.Schema == "" {
		cfg.Schema = "public"
	} else {
		cfg.Schema += ", public, extensions"
	}

	const (
		tmpl     = "%s://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s"
		tmplNoDB = "%s://%s:%s@%s:%s/?sslmode=disable&search_path=%s"
	)

	if excludeDb {
		return fmt.Sprintf(tmplNoDB, driver, user, pass, host, port, cfg.Schema)
	}

	return fmt.Sprintf(tmpl, driver, user, pass, host, port, dbname, cfg.Schema)
}
