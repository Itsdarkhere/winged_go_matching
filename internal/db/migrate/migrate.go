package migrate

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/util/validationlib"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// migrate_ runs the database migrations using the provided options.
func migrate_(ctx context.Context, opts *Options) error {
	if err := validationlib.Validate(opts); err != nil {
		return fmt.Errorf("validate options: %w", err)
	}

	d, err := iofs.New(opts.Stmts, ".")
	if err != nil {
		return fmt.Errorf("expected no error creating iofs instance: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, opts.connTmpl)
	if err != nil {
		return fmt.Errorf("expected no error creating migrate instance: %w, with connTmpl: %s", err, opts.connTmpl)
	}

	if err = m.Up(); err != nil {
		return fmt.Errorf("expected no error running migrations: %w, with connTmpl: %s", err, opts.connTmpl)
	}

	if _, err := m.Close(); err != nil {
		return fmt.Errorf("expected no error closing migrate instance: %w", err)
	}

	return nil
}
