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

type UserDateTypePreference struct {
	t           *testing.T
	db          boil.ContextExecutor
	Subject     *pgmodel.UserDateTypePreference
	FactoryUser *User
}

// New initializes a new UserDateTypePreference instance for testing.
func (c *UserDateTypePreference) New(t *testing.T, db boil.ContextExecutor) *UserDateTypePreference {
	sub := &pgmodel.UserDateTypePreference{}
	var factoryUser *User

	if c != nil {
		if c.Subject != nil {
			sub = c.Subject
		}
		factoryUser = c.FactoryUser
	}

	return &UserDateTypePreference{
		t:           t,
		db:          db,
		Subject:     sub,
		FactoryUser: factoryUser,
	}
}

func (c *UserDateTypePreference) IsValid() bool {
	return c != nil && c.Subject != nil && c.Subject.ID != ""
}

// Save persists the UserDateTypePreference instance to the database.
func (c *UserDateTypePreference) Save() *UserDateTypePreference {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting UserDateTypePreference")
	return c
}

// SetRequiredFields sets the required default fields for UserDateTypePreference.
func (c *UserDateTypePreference) SetRequiredFields() *UserDateTypePreference {
	// DateTypeCore is now a string enum
	if c.Subject.DateTypeCore == "" {
		c.Subject.DateTypeCore = string(enums.DateTypeCoreCoffee)
	}
	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *UserDateTypePreference) EnsureFKDeps() {
	// User FK - only create if not already set
	if c.Subject.UserID == "" {
		if c.FactoryUser == nil || !c.FactoryUser.IsValid() {
			c.FactoryUser = (&factory.Entity[*User]{}).New(c.t, c.db)
		}
		c.Subject.UserID = c.FactoryUser.Subject.ID
	}
}
