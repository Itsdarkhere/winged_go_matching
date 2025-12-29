package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/require"
)

type UserBlockedContact struct {
	t           *testing.T
	db          boil.ContextExecutor
	Subject     *pgmodel.UserBlockedContact
	FactoryUser *User
}

// New initializes a new UserBlockedContact instance for testing.
func (c *UserBlockedContact) New(t *testing.T, db boil.ContextExecutor) *UserBlockedContact {
	sub := &pgmodel.UserBlockedContact{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	sub.ID = "" // ensure PK always empty for new
	return &UserBlockedContact{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

func (c *UserBlockedContact) IsValid() bool {
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

// Save persists the UserBlockedContact instance to the database.
func (c *UserBlockedContact) Save() *UserBlockedContact {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting UserBlockedContact")
	return c
}

// SetRequiredFields sets the required default fields for UserBlockedContact.
func (c *UserBlockedContact) SetRequiredFields() *UserBlockedContact {
	if c.Subject.UserID == "" {
		c.Subject.UserID = c.FactoryUser.Subject.ID
	}
	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *UserBlockedContact) EnsureFKDeps() {
	if !c.FactoryUser.IsValid() && c.Subject.UserID == "" {
		c.FactoryUser = (&factory.Entity[*User]{}).New(c.t, c.db)
	}
}
