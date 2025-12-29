package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/enums"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/require"
)

type WingsEcnTransaction struct {
	t       *testing.T
	db      boil.ContextExecutor
	Subject *pgmodel.WingsEcnTransaction

	FactoryUser      *User
	FactoryActionLog *WingsEcnActionLog
}

// New initializes a new WingsEcnTransaction instance for testing.
func (c *WingsEcnTransaction) New(t *testing.T, db boil.ContextExecutor) *WingsEcnTransaction {
	sub := &pgmodel.WingsEcnTransaction{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	return &WingsEcnTransaction{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

func (c *WingsEcnTransaction) IsValid() bool {
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
func (c *WingsEcnTransaction) EnsureFKDeps() {
	if !c.FactoryUser.IsValid() && c.Subject.UserRefID == "" {
		c.FactoryUser = (&factory.Entity[*User]{}).New(c.t, c.db)
	}
	if c.Subject.UserRefID == "" && c.FactoryUser.IsValid() {
		c.Subject.UserRefID = c.FactoryUser.Subject.ID
	}

	if c.Subject.ActionLogRefID == "" && !c.FactoryActionLog.IsValid() {
		c.FactoryActionLog = (&factory.Entity[*WingsEcnActionLog]{}).New(c.t, c.db)
	}
	if c.Subject.ActionLogRefID == "" && c.FactoryActionLog.IsValid() {
		c.Subject.ActionLogRefID = c.FactoryActionLog.Subject.ID
	}
}

// SetRequiredFields sets default required fields.
func (c *WingsEcnTransaction) SetRequiredFields() *WingsEcnTransaction {
	// ActionLogType is now a string enum
	if c.Subject.ActionLogType == "" {
		c.Subject.ActionLogType = string(enums.WingsEconomyActionLogDailyCheckIn)
	}
	// IsCredit, Claimed, Amount are required by schema but accept zero-values; leave unless pre-set.
	return c
}

// Save persists the WingsEcnTransaction instance to the database.
func (c *WingsEcnTransaction) Save() *WingsEcnTransaction {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting WingsEcnTransaction")
	return c
}
