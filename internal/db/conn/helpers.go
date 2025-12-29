package conn

import (
	"context"
	"fmt"
	"reflect"
	"time"
	"wingedapp/pgtester/internal/util/strutil"

	"github.com/alitto/pond"
	"github.com/jmoiron/sqlx"
)

const (
	defaultTimeout = 5 * time.Second
)

const (
	sqlCount = "SELECT COUNT(*) as count FROM (%s) AS subquery"
)

func Count(ctx context.Context, db *sqlx.DB, sqlSTMT string) (int, error) {
	var (
		count        int
		sqlSTMTCount = fmt.Sprintf(sqlCount, sqlSTMT)
	)

	if err := db.GetContext(ctx, &count, sqlSTMTCount); err != nil {
		return count, fmt.Errorf("count with sql: %s, with err %w", sqlSTMTCount, err)
	}
	return count, nil
}

// QueryArray executes a SQL statement and returns the results as an array of strings.
func QueryArray(ctx context.Context, exec sqlx.QueryerContext, sqlSTMT string, excludeHeader ...bool) ([]any, error) {
	ctx, cancel := ctxTimeoutDefault(ctx)
	defer cancel()

	rows, err := exec.QueryxContext(ctx, sqlSTMT)
	if err != nil {
		return nil, fmt.Errorf("query sql: %w", err)
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("query sql columns: %w", err)
	}

	results := make([]any, 0)
	for rows.Next() {
		values := make([]any, len(cols))
		for i := range values {
			values[i] = new(any)
		}
		if err = rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		for i := range values {
			values[i] = Dereference(values[i])
			if b, ok := values[i].([]uint8); ok {
				values[i] = string(b)
			}
		}

		results = append(results, values)
	}

	if err = rows.Close(); err != nil {
		return nil, fmt.Errorf("close rows: %w", err)
	}

	if len(results) > 0 {
		columnNames := make([]any, len(cols))
		for i, col := range cols {
			columnNames[i] = col
		}
		if len(excludeHeader) == 0 || !excludeHeader[0] {
			results = append([]any{columnNames}, results...)
		}
	}

	return results, nil
}

func ctxTimeoutDefault(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, defaultTimeout)
}

// DropMySQLDB drops a MySQL database if it exists.
func DropMySQLDB(ctx context.Context, db *sqlx.DB, dbName string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
	if err != nil {
		return fmt.Errorf("failed to drop MySQL database %s: %w", dbName, err)
	}
	return nil
}

// CreateMySQLDB creates a MySQL database if it does not exist.
func CreateMySQLDB(ctx context.Context, db *sqlx.DB, dbName string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
	if err != nil {
		return fmt.Errorf("failed to create MySQL database %s: %w", dbName, err)
	}
	return nil
}

// MarkReadyForDeletion updates current session name as 'ready-for-deletion'
func MarkReadyForDeletion(ctx context.Context, db *sqlx.DB) error {
	res, err := db.ExecContext(ctx, fmt.Sprintf("SET application_name = '%s';", ReadyForDeletion))
	if err != nil {
		return fmt.Errorf("failed to rename backend: %w", err)
	}
	if _, err = res.RowsAffected(); err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	return nil
}

func TerminateReadyForDeletion(ctx context.Context, db *sqlx.DB) error {
	pgPIDS, err := PGPids(ctx, db, ReadyForDeletion)
	if err != nil {
		return fmt.Errorf("failed to get pids for app name %s: %w", ReadyForDeletion, err)
	}

	for _, pgPID := range pgPIDS {
		if err = TerminatePid(ctx, db, pgPID.Pid); err != nil {
			return fmt.Errorf("failed to terminate pid %d: %w", pgPID.Pid, err)
		}
	}

	return nil
}

func TerminatePid(ctx context.Context, db *sqlx.DB, pid int) error {
	if _, err := db.ExecContext(ctx, "SELECT pg_terminate_backend($1)", pid); err != nil {
		return fmt.Errorf("failed to terminate backend %d: %w", pid, err)
	}
	fmt.Println("==== deleted pid:", pid)
	return nil
}

func PGPidsCount(ctx context.Context, db *sqlx.DB) (int, error) {
	var count int
	if err := db.GetContext(ctx, &count, sqlStmtCountPids); err != nil {
		return count, fmt.Errorf("failed to select pg_stat_activity: %w", err)
	}
	return count, nil
}

func PGPids(ctx context.Context, db *sqlx.DB, applicationName string) ([]PID, error) {
	var r []PID
	if err := db.SelectContext(ctx, &r, fmt.Sprintf(sqlStmtSelectPidsWhereAppName), applicationName); err != nil {
		return r, fmt.Errorf("failed to select pg_stat_activity: %w", err)
	}
	return r, nil
}

func PGPid(ctx context.Context, db *sqlx.DB, dbName string) (int, error) {
	type resp struct {
		Pid int `db:"pg_backend_pid"`
	}
	var r resp
	if err := db.GetContext(ctx, &r, fmt.Sprintf("SELECT pg_backend_pid()")); err != nil {
		return 0, fmt.Errorf("failed to drop PostgreSQL database %s: %w", dbName, err)
	}
	return r.Pid, nil
}

// DropPGDB drops a PostgreSQL database with CASCADE.
func DropPGDB(ctx context.Context, db *sqlx.DB, dbName string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS \"%s\" WITH (FORCE)", dbName))
	if err != nil {
		return fmt.Errorf("failed to drop PostgreSQL database %s: %w", dbName, err)
	}
	return nil
}

// CreatePGDB creates a PostgreSQL database if it does not exist.
func CreatePGDB(ctx context.Context, db *sqlx.DB, dbName string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE \"%s\"", dbName))
	if err != nil {
		return fmt.Errorf("failed to create PostgreSQL database %s: %w", dbName, err)
	}
	return nil
}

// DropPGSchema drops a PostgreSQL schema with CASCADE.
func DropPGSchema(ctx context.Context, db *sqlx.DB, schemaName string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS \"%s\" CASCADE", schemaName))
	if err != nil {
		return fmt.Errorf("failed to drop PostgreSQL schema %s: %w", schemaName, err)
	}
	return nil
}

// CreatePGSchema creates a PostgreSQL schema if it does not exist.
func CreatePGSchema(ctx context.Context, db *sqlx.DB, schemaName string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS \"%s\"", schemaName))
	if err != nil {
		return fmt.Errorf("failed to create PostgreSQL schema %s: %w", schemaName, err)
	}
	return nil
}

// ListPGDBs lists all PostgreSQL databases matching the given schema name pattern.
func ListPGDBs(ctx context.Context, db *sqlx.DB, wildcard string) ([]string, error) {
	pgDBS := PGDBs{}

	sqlStmt := `SELECT datname as name FROM pg_database WHERE datname LIKE '%s'`
	sqlStmt = fmt.Sprintf(sqlStmt, wildcard) // careful here with SQL injection
	if err := db.SelectContext(ctx, &pgDBS, sqlStmt); err != nil {
		return nil, fmt.Errorf("list pgdbs: %w", err)
	}
	return pgDBS.StrArr(), nil
}

func DropDBS(ctx context.Context, db *sqlx.DB, currentDBConn, wildcard string) error {
	testDBs, err := ListPGDBs(ctx, db, wildcard)
	if err != nil {
		return fmt.Errorf("error listing test databases: %w", err)
	}

	dropDBFn := func(dbName string) func() {
		return func() {
			err := DropPGDB(ctx, db, dbName)
			if err != nil {
				fmt.Printf("error dropping test database %s: %v\n", dbName, err)
			} else {
				fmt.Println("==== dropped test db:", dbName)
			}
		}
	}

	// drop concurrently by the tens
	p := pond.New(10, 10)
	for _, testDB := range testDBs {
		if testDB == currentDBConn {
			continue
		}
		p.Submit(dropDBFn(testDB))
	}
	p.StopAndWait()

	fmt.Println("==== currentDBConn:", currentDBConn)
	fmt.Println("==== testDBs:", strutil.GetAsJson(testDBs))
	return nil
}

func Dereference(i interface{}) interface{} {
	if reflect.ValueOf(i).Kind() == reflect.Ptr {
		return reflect.Indirect(reflect.ValueOf(i)).Interface()
	}
	return i
}
