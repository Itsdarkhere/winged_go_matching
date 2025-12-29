package conn

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
	"wingedapp/pgtester/internal/util/validationlib"

	"github.com/gofiber/fiber/v3/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func (c *Cfg) PGConn(ctx context.Context, excludeDb bool) (*sqlx.DB, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	connTmpl, err := PGTmpl(
		&PGConnParams{
			User:           c.User,
			Pass:           c.Pass,
			Host:           c.Host,
			Database:       c.Database,
			Schema:         c.Schema,
			Port:           c.Port,
			SearchPathBase: c.SearchPath,
		},
		excludeDb,
	)
	if err != nil {
		return nil, fmt.Errorf("pg tmpl: %w", err)
	}

	db, err := sqlx.ConnectContext(ctx, string(Postgres), connTmpl)
	if err != nil {
		log.Info("failed with connTmpl", connTmpl)
		return nil, fmt.Errorf("connect ctx: %w", err)
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping ctx: %w", err)
	}

	// open connections
	maxOpenConns := 50
	if c.MaxOpenConns > 0 {
		maxOpenConns = c.MaxOpenConns
	}
	maxIdleConns := 25
	if c.MaxIdleConns > 0 {
		maxIdleConns = c.MaxIdleConns
	}
	maxConnLifetime := 30
	if c.MaxConnLifetime.Valid {
		maxConnLifetime = c.MaxConnLifetime.Int
	}

	db.SetConnMaxLifetime(time.Minute * time.Duration(maxConnLifetime))
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	return db, nil
}

func PGTmpl(p *PGConnParams, excludeDatabase bool) (string, error) {
	if err := p.Validate(); err != nil {
		return "", fmt.Errorf("validate: %w", err)
	}

	portStr := strconv.Itoa(p.Port)

	var (
		connTmpl     = "postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s"
		connTmplNoDB = "postgres://%s:%s@%s:%s/?sslmode=disable&search_path=%s"
	)

	params := []any{p.User, url.QueryEscape(p.Pass), p.Host, portStr}
	tmpl := connTmplNoDB

	if !excludeDatabase {
		tmpl = connTmpl
		params = append(params, p.Database)
	}

	searchPath := p.Schema
	if p.SearchPathBase != "" {
		searchPath += ", " + p.SearchPathBase
	}
	params = append(params, searchPath)

	return fmt.Sprintf(tmpl, params...), nil
}

type PGConnParams struct {
	User           string `json:"user" validate:"required"`
	Pass           string `json:"pass" validate:"required"`
	Host           string `json:"host" validate:"required"`
	Database       string `json:"database" validate:"required"`
	Schema         string `json:"schema" validate:"required"`
	Port           int    `json:"port" validate:"required"`
	SearchPathBase string `json:"search_path_base"` // for Supabase, usually "public"

	MaxOpenConns int `json:"max_open_conns"`
	MaxIdleConns int `json:"max_idle_conns"`
}

func (p *PGConnParams) Validate() error {
	return validationlib.Validate(p)
}
