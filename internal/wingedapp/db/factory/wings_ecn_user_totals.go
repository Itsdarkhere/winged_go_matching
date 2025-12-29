package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/require"
)

type WingsEcnUserTotal struct {
	t       *testing.T
	db      boil.ContextExecutor
	Subject *pgmodel.WingsEcnUserTotal

	FactoryUser *User
}

// New initializes a new WingsEcnUserTotal instance for testing.
func (c *WingsEcnUserTotal) New(t *testing.T, db boil.ContextExecutor) *WingsEcnUserTotal {
	sub := &pgmodel.WingsEcnUserTotal{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	return &WingsEcnUserTotal{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

func (c *WingsEcnUserTotal) IsValid() bool {
	if c == nil {
		return false
	}
	if c.Subject == nil {
		return false
	}
	if c.Subject.ID == "" {
		return false
	}
	return true
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *WingsEcnUserTotal) EnsureFKDeps() {
	if !c.FactoryUser.IsValid() && c.Subject.UserRefID == "" {
		c.FactoryUser = (&factory.Entity[*User]{}).New(c.t, c.db)
	}
	if c.Subject.UserRefID == "" {
		c.Subject.UserRefID = c.FactoryUser.Subject.ID
	}
}

// SetRequiredFields sets default required fields.
func (c *WingsEcnUserTotal) SetRequiredFields() *WingsEcnUserTotal {
	if c.Subject.UserRefID == "" {
		c.Subject.UserRefID = c.FactoryUser.Subject.ID
	}
	return c
}

// Save persists the WingsEcnUserTotal instance to the database.
func (c *WingsEcnUserTotal) Save() *WingsEcnUserTotal {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting WingsEcnUserTotal")
	return c
}
