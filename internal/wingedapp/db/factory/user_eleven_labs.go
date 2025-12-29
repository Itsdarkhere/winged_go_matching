package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/require"
)

type UserElevenLabs struct {
	t           *testing.T
	db          boil.ContextExecutor
	Subject     *pgmodel.UserElevenLab
	FactoryUser *User
}

// New initializes a new UserElevenLabs instance for testing.
func (c *UserElevenLabs) New(t *testing.T, db boil.ContextExecutor) *UserElevenLabs {
	sub := &pgmodel.UserElevenLab{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	return &UserElevenLabs{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

// IsValid checks if the UserElevenLabs instance is valid for use in tests.
func (c *UserElevenLabs) IsValid() bool {
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

// EnsureFKDeps ensures that any foreign key dependencies are persisted.
func (c *UserElevenLabs) EnsureFKDeps() {
	if !c.FactoryUser.IsValid() {
		c.FactoryUser = (&factory.Entity[*User]{}).New(c.t, c.db)
	}
	c.Subject.UserID = c.FactoryUser.Subject.ID
}

// Save persists the UserElevenLabs instance to the database.
func (c *UserElevenLabs) Save() *UserElevenLabs {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting UserElevenLabs")
	return c
}

// SetRequiredFields sets the required default fields for UserElevenLabs.
func (c *UserElevenLabs) SetRequiredFields() *UserElevenLabs {
	if c.Subject.UserID == "" {
		c.Subject.UserID = c.FactoryUser.Subject.ID
	}

	if c.Subject.Conversation == nil {
		c.Subject.Conversation = marshal(c.t, map[string]any{
			"dummy": "Type",
		})
	}
	return c
}
