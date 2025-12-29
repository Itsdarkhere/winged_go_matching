package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type Notification struct {
	t           *testing.T
	db          boil.ContextExecutor
	Subject     *pgmodel.Notification
	FactoryUser *User
}

// New initializes a new Notification instance for testing.
func (c *Notification) New(t *testing.T, db boil.ContextExecutor) *Notification {
	sub := &pgmodel.Notification{}
	var factoryUser *User

	if c != nil {
		if c.Subject != nil {
			sub = c.Subject
		}
		factoryUser = c.FactoryUser
	}

	return &Notification{
		t:           t,
		db:          db,
		Subject:     sub,
		FactoryUser: factoryUser,
	}
}

func (c *Notification) IsValid() bool {
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

// Save persists the Notification instance to the database.
func (c *Notification) Save() *Notification {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting Notification")
	return c
}

// SetRequiredFields sets the required default fields for Notification.
func (c *Notification) SetRequiredFields() *Notification {
	if c.Subject.ID == "" {
		c.Subject.ID = uuid.NewString()
	}
	if c.Subject.Message == "" {
		c.Subject.Message = "Test notification message"
	}
	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *Notification) EnsureFKDeps() {
	// Only create user if not already set
	if c.Subject.UserRefID == "" {
		if !c.FactoryUser.IsValid() {
			c.FactoryUser = (&factory.Entity[*User]{}).New(c.t, c.db)
		}
		c.Subject.UserRefID = c.FactoryUser.Subject.ID
	}
}
