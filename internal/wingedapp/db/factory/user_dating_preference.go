package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/require"
)

type UserDatingPreference struct {
	t           *testing.T
	db          boil.ContextExecutor
	Subject     *pgmodel.UserDatingPreference
	FactoryUser *User
}

// New initializes a new UserDatingPreference instance for testing.
func (c *UserDatingPreference) New(t *testing.T, db boil.ContextExecutor) *UserDatingPreference {
	sub := &pgmodel.UserDatingPreference{}
	var factoryUser *User

	if c != nil {
		if c.Subject != nil {
			sub = c.Subject
		}
		factoryUser = c.FactoryUser
	}

	return &UserDatingPreference{
		t:           t,
		db:          db,
		Subject:     sub,
		FactoryUser: factoryUser,
	}
}

func (c *UserDatingPreference) IsValid() bool {
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

// Save persists the UserDatingPreference instance to the database.
func (c *UserDatingPreference) Save() *UserDatingPreference {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting UserDatingPreference")
	return c
}

// SetRequiredFields sets the required default fields for UserDatingPreference.
func (c *UserDatingPreference) SetRequiredFields() *UserDatingPreference {
	if c.Subject.UserID == "" {
		c.Subject.UserID = c.FactoryUser.Subject.ID
	}
	if c.Subject.DatingPreference == "" {
		c.Subject.DatingPreference = pgmodel.DatingPreferencesNonBinary
	}

	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *UserDatingPreference) EnsureFKDeps() {
	if !c.FactoryUser.IsValid() && c.Subject.UserID == "" {
		c.FactoryUser = (&factory.Entity[*User]{}).New(c.t, c.db)
	}
	if c.Subject.UserID == "" {
		c.Subject.UserID = c.FactoryUser.Subject.ID
	}
}
