package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/require"
)

type AnonymizedContact struct {
	t           *testing.T
	db          boil.ContextExecutor
	Subject     *pgmodel.AnonymizedContact
	FactoryUser *User
}

// New initializes a new AnonymizedContact instance for testing.
func (c *AnonymizedContact) New(t *testing.T, db boil.ContextExecutor) *AnonymizedContact {
	sub := &pgmodel.AnonymizedContact{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	return &AnonymizedContact{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

func (c *AnonymizedContact) IsValid() bool {
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

// Save persists the AnonymizedContact instance to the database.
func (c *AnonymizedContact) Save() *AnonymizedContact {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting AnonymizedContact")
	return c
}

// SetRequiredFields sets the required default fields for AnonymizedContact.
func (c *AnonymizedContact) SetRequiredFields() *AnonymizedContact {
	if c.Subject.OwnerHash == "" {
		c.Subject.OwnerHash = randomSha256()
	}
	if c.Subject.ContactHash == "" {
		c.Subject.ContactHash = randomSha256()
	}
	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *AnonymizedContact) EnsureFKDeps() {
	if !c.FactoryUser.IsValid() {
		c.FactoryUser = (&factory.Entity[*User]{}).New(c.t, c.db)
	}
}
