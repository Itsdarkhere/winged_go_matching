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

type UserDietaryRestriction struct {
	t           *testing.T
	db          boil.ContextExecutor
	Subject     *pgmodel.UserDietaryRestriction
	FactoryUser *User
}

// New initializes a new UserDietaryRestriction instance for testing.
func (c *UserDietaryRestriction) New(t *testing.T, db boil.ContextExecutor) *UserDietaryRestriction {
	sub := &pgmodel.UserDietaryRestriction{}
	var factoryUser *User

	if c != nil {
		if c.Subject != nil {
			sub = c.Subject
		}
		factoryUser = c.FactoryUser
	}

	return &UserDietaryRestriction{
		t:           t,
		db:          db,
		Subject:     sub,
		FactoryUser: factoryUser,
	}
}

func (c *UserDietaryRestriction) IsValid() bool {
	return c != nil && c.Subject != nil && c.Subject.ID != ""
}

// Save persists the UserDietaryRestriction instance to the database.
func (c *UserDietaryRestriction) Save() *UserDietaryRestriction {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting UserDietaryRestriction")
	return c
}

// SetRequiredFields sets the required default fields for UserDietaryRestriction.
func (c *UserDietaryRestriction) SetRequiredFields() *UserDietaryRestriction {
	// DietaryRestriction is now a string enum
	if c.Subject.DietaryRestriction == "" {
		c.Subject.DietaryRestriction = string(enums.DietaryRestrictionNoRestriction)
	}
	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *UserDietaryRestriction) EnsureFKDeps() {
	// User FK - only create if not already set
	if c.Subject.UserID == "" {
		if c.FactoryUser == nil || !c.FactoryUser.IsValid() {
			c.FactoryUser = (&factory.Entity[*User]{}).New(c.t, c.db)
		}
		c.Subject.UserID = c.FactoryUser.Subject.ID
	}
}
