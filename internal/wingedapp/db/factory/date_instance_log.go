package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type DateInstanceLog struct {
	t                   *testing.T
	db                  boil.ContextExecutor
	Subject             *pgmodel.DateInstanceLog
	FactoryDateInstance *DateInstance
	FactoryUser         *User
}

// New initializes a new DateInstanceLog instance for testing.
func (c *DateInstanceLog) New(t *testing.T, db boil.ContextExecutor) *DateInstanceLog {
	sub := &pgmodel.DateInstanceLog{}
	var factoryDateInstance *DateInstance
	var factoryUser *User

	if c != nil {
		if c.Subject != nil {
			sub = c.Subject
		}
		factoryDateInstance = c.FactoryDateInstance
		factoryUser = c.FactoryUser
	}

	return &DateInstanceLog{
		t:                   t,
		db:                  db,
		Subject:             sub,
		FactoryDateInstance: factoryDateInstance,
		FactoryUser:         factoryUser,
	}
}

func (c *DateInstanceLog) IsValid() bool {
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

// Save persists the DateInstanceLog instance to the database.
func (c *DateInstanceLog) Save() *DateInstanceLog {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting DateInstanceLog")
	return c
}

// SetRequiredFields sets the required default fields for DateInstanceLog.
func (c *DateInstanceLog) SetRequiredFields() *DateInstanceLog {
	if c.Subject.ID == "" {
		c.Subject.ID = uuid.NewString()
	}
	if c.Subject.EventType == "" {
		c.Subject.EventType = "test_event"
	}
	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *DateInstanceLog) EnsureFKDeps() {
	// Only create date instance if not already set
	if c.Subject.DateInstanceRefID == "" {
		if !c.FactoryDateInstance.IsValid() {
			c.FactoryDateInstance = (&factory.Entity[*DateInstance]{}).New(c.t, c.db)
		}
		c.Subject.DateInstanceRefID = c.FactoryDateInstance.Subject.ID
	}

	// UserRefID is nullable - only set if FactoryUser was explicitly provided
	if !c.Subject.UserRefID.Valid && c.FactoryUser != nil && c.FactoryUser.IsValid() {
		c.Subject.UserRefID = null.StringFrom(c.FactoryUser.Subject.ID)
	}
}
