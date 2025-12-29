package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"wingedapp/pgtester/internal/db/conn"
	"wingedapp/pgtester/internal/db/testpg"
	"wingedapp/pgtester/internal/util/validationlib"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/jmoiron/sqlx"
)

const (
	defaultConnTimeout = 10 * time.Second
)

//type TransactorImapp struct {
//	*Transactor
//}
//
//// NewTransactorImapp creates a new NewTransactorImapp instance,
//// embedding the provided Transactor.
//// This is just so we can differentiate our struct types in DI containers.
//func NewTransactorImapp(t *Transactor) *TransactorImapp {
//	return &TransactorImapp{Transactor: t}
//}

// Transactor implements the Transactor interface for managing database transactions.
type Transactor struct {
	db     *sqlx.DB
	DBName string
}

type Config struct {
	Host         string `json:"host" validate:"required"`
	User         string `json:"user" validate:"required"`
	Pass         string `json:"pass" validate:"required"`
	Database     string `json:"database" validate:"required"`
	Port         int    `json:"port" validate:"required"`
	Schema       string `json:"schema" validate:"required"`
	MaxIdleConns int    `json:"maxIdleConns" validate:"required"`
	MaxOpenConns int    `json:"maxOpenConns" validate:"required"`
}

func (c *Config) Validate() error {
	return validationlib.Validate(c)
}

func NewTransactor(cfg *Config) (*Transactor, error) {
	// In test container mode, return a transactor with nil connection.
	// The testsuite will set the actual connection via SetDBConn later.
	// This prevents connection exhaustion when many tests run in parallel.
	if testpg.IsTestContainerMode() {
		return &Transactor{db: nil}, nil
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultConnTimeout)
	defer cancel()

	connCfg := conn.Cfg{
		Database:     cfg.Database,
		Host:         cfg.Host,
		Port:         cfg.Port,
		User:         cfg.User,
		Pass:         cfg.Pass,
		Schema:       cfg.Schema,
		MaxIdleConns: cfg.MaxIdleConns,
		MaxOpenConns: cfg.MaxOpenConns,
	}

	db, err := connCfg.PGConn(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("pg conn: %w", err)
	}

	t := &Transactor{
		db: db,
	}
	fmt.Println("==== db inside:", db)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return t, nil
}

func (t *Transactor) DB() boil.ContextExecutor {
	return t.db
}

func (t *Transactor) TX() (boil.ContextTransactor, error) {
	tx, err := t.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("beginx: %w", err)
	}

	return tx, nil
}

// Rollback rolls back the transaction if it exists.
func (t *Transactor) Rollback(exec boil.ContextTransactor) {
	if exec == nil {
		return
	}

	if err := exec.Rollback(); err != nil {
		if !errors.Is(err, sql.ErrTxDone) { // only log legit errs
			fmt.Printf("rollback: %v\n", err)
		}
	}
}

func (t *Transactor) SetDBConn(db *sqlx.DB) {
	t.db = db
}

// IsConnected returns true if the transactor has a database connection.
func (t *Transactor) IsConnected() bool {
	return t.db != nil
}

func (t *Transactor) Close() error {
	if t.db != nil {
		if err := t.db.Close(); err != nil {
			return fmt.Errorf("close db: %w", err)
		}
	}
	return nil
}
