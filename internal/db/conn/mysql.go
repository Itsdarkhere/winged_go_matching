package conn

import (
	"context"
	"fmt"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func (c *Cfg) MySQLConn(ctx context.Context, excludeDb bool) (*sqlx.DB, error) {
	if !excludeDb && c.Database == "" {
		return nil, fmt.Errorf("missing database in config")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	db, err := sqlx.ConnectContext(ctx, MySQL, c.mysqlTmpl(excludeDb))
	if err != nil {
		return nil, fmt.Errorf("connect ctx: %w", err)
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping ctx: %w", err)
	}

	db.SetConnMaxLifetime(time.Minute * 60)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(10)
	return db, nil
}

func (c *Cfg) mysqlTmpl(excludeDatabase bool) string {
	portStr := strconv.Itoa(c.Port)

	var (
		connTmpl     = "%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&tls=false"
		connTmplNoDB = "%s:%s@tcp(%s:%s)/?charset=utf8&parseTime=true&tls=false"
	)

	params := []any{c.User, c.Pass, c.Host, portStr}
	tmpl := connTmplNoDB

	if !excludeDatabase {
		tmpl = connTmpl
		params = append(params, c.Database)
	}

	return fmt.Sprintf(tmpl, params...)
}
