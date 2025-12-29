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

type WingsEcnActionLog struct {
	t           *testing.T
	db          boil.ContextExecutor
	Subject     *pgmodel.WingsEcnActionLog
	FactoryUser *User
}

// New initializes a new WingsEcnActionLog instance for testing.
func (c *WingsEcnActionLog) New(t *testing.T, db boil.ContextExecutor) *WingsEcnActionLog {
	sub := &pgmodel.WingsEcnActionLog{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	return &WingsEcnActionLog{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

func (c *WingsEcnActionLog) IsValid() bool {
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
func (c *WingsEcnActionLog) EnsureFKDeps() {
	if !c.FactoryUser.IsValid() && c.Subject.UserRefID == "" {
		c.FactoryUser = (&factory.Entity[*User]{}).New(c.t, c.db)
	}
	if c.Subject.UserRefID == "" && c.FactoryUser.IsValid() {
		c.Subject.UserRefID = c.FactoryUser.Subject.ID
	}
}

// SetRequiredFields sets default required fields.
func (c *WingsEcnActionLog) SetRequiredFields() *WingsEcnActionLog {
	if c.Subject.ExtDomainRefID == "" {
		c.Subject.ExtDomainRefID = RandomString()
	}
	// ActionLogType is now a string enum
	if c.Subject.ActionLogType == "" {
		c.Subject.ActionLogType = string(enums.WingsEconomyActionLogDailyCheckIn)
	}
	// IsCredit is required but has a sane zero-value (false), so no-op if unset.
	return c
}

// Save persists the WingsEcnActionLog instance to the database.
func (c *WingsEcnActionLog) Save() *WingsEcnActionLog {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting WingsEcnActionLog")
	return c
}
