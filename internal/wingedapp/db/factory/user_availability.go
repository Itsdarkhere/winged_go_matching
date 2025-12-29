package factory

import (
	"context"
	"fmt"
	"testing"
	"time"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/require"
)

type UserAvailability struct {
	t           *testing.T
	db          boil.ContextExecutor
	Subject     *pgmodel.UserAvailability
	FactoryUser *User
}

// New initializes a new UserAvailability instance for testing.
func (c *UserAvailability) New(t *testing.T, db boil.ContextExecutor) *UserAvailability {
	sub := &pgmodel.UserAvailability{}
	var factoryUser *User

	if c != nil {
		if c.Subject != nil {
			sub = c.Subject
		}
		factoryUser = c.FactoryUser
	}

	return &UserAvailability{
		t:           t,
		db:          db,
		Subject:     sub,
		FactoryUser: factoryUser,
	}
}

func (c *UserAvailability) IsValid() bool {
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

// Save persists the UserAvailability instance to the database.
func (c *UserAvailability) Save() *UserAvailability {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting UserAvailability")
	return c
}

// SetRequiredFields sets the required default fields for UserAvailability.
func (c *UserAvailability) SetRequiredFields() *UserAvailability {
	if c.Subject.TimeBlock == "" {
		now := time.Now().UTC().Truncate(time.Hour)
		c.Subject.TimeBlock = fmt.Sprintf("[%s,%s)",
			now.Add(24*time.Hour).Format(time.RFC3339),
			now.Add(26*time.Hour).Format(time.RFC3339))
	}

	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *UserAvailability) EnsureFKDeps() {
	if c.Subject.UserID == "" {
		if !c.FactoryUser.IsValid() {
			c.FactoryUser = (&factory.Entity[*User]{}).New(c.t, c.db)
		}
		c.Subject.UserID = c.FactoryUser.Subject.ID
	}
}
