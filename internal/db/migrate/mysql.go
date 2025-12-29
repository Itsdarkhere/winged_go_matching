package migrate

import (
	"context"
	"fmt"
	"strconv"
	"wingedapp/pgtester/internal/db/conn"
)

// MySQL initializes the MySQL database, and runs migrations.
func MySQL(ctx context.Context, opts *Options) error {
	if opts != nil {
		opts.connTmpl = mySQLTMPL(opts.ConnCfg, false)
	}

	if err := migrate_(ctx, opts); err != nil {
		return fmt.Errorf("expected no error running migrations: %w", err)
	}

	return nil
}

func mySQLTMPL(cfg *conn.Cfg, excludeDb bool) string {
	var (
		prefix = "mysql"
		user   = cfg.User
		pass   = cfg.Pass
		host   = cfg.Host
		port   = strconv.Itoa(cfg.Port)
		dbname = cfg.Database
	)

	const (
		tmpl     = "%s://%s:%s@tcp(%s:%s)/%s?multiStatements=true"
		tmplNoDB = "%s://%s:%s@tcp(%s:%s)/?multiStatements=true"
	)

	if excludeDb {
		return fmt.Sprintf(tmplNoDB, prefix, user, pass, host, port)
	}

	return fmt.Sprintf(tmpl, prefix, user, pass, host, port, dbname)
}
