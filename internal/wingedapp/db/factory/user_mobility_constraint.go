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

type UserMobilityConstraint struct {
	t           *testing.T
	db          boil.ContextExecutor
	Subject     *pgmodel.UserMobilityConstraint
	FactoryUser *User
}

// New initializes a new UserMobilityConstraint instance for testing.
func (c *UserMobilityConstraint) New(t *testing.T, db boil.ContextExecutor) *UserMobilityConstraint {
	sub := &pgmodel.UserMobilityConstraint{}
	var factoryUser *User

	if c != nil {
		if c.Subject != nil {
			sub = c.Subject
		}
		factoryUser = c.FactoryUser
	}

	return &UserMobilityConstraint{
		t:           t,
		db:          db,
		Subject:     sub,
		FactoryUser: factoryUser,
	}
}

func (c *UserMobilityConstraint) IsValid() bool {
	return c != nil && c.Subject != nil && c.Subject.ID != ""
}

// Save persists the UserMobilityConstraint instance to the database.
func (c *UserMobilityConstraint) Save() *UserMobilityConstraint {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting UserMobilityConstraint")
	return c
}

// SetRequiredFields sets the required default fields for UserMobilityConstraint.
func (c *UserMobilityConstraint) SetRequiredFields() *UserMobilityConstraint {
	// MobilityConstraint is now a string enum
	if c.Subject.MobilityConstraint == "" {
		c.Subject.MobilityConstraint = string(enums.MobilityConstraintNone)
	}
	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *UserMobilityConstraint) EnsureFKDeps() {
	// User FK - only create if not already set
	if c.Subject.UserID == "" {
		if c.FactoryUser == nil || !c.FactoryUser.IsValid() {
			c.FactoryUser = (&factory.Entity[*User]{}).New(c.t, c.db)
		}
		c.Subject.UserID = c.FactoryUser.Subject.ID
	}
}
