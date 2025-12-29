package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/require"
)

type AgentLog struct {
	t           *testing.T
	db          boil.ContextExecutor
	Subject     *pgmodel.AgentLog
	FactoryUser *User
}

// New initializes a new AgentLog instance for testing.
func (c *AgentLog) New(t *testing.T, db boil.ContextExecutor) *AgentLog {
	sub := &pgmodel.AgentLog{}
	var factoryUser *User

	if c != nil {
		if c.Subject != nil {
			sub = c.Subject
		}
		factoryUser = c.FactoryUser
	}

	return &AgentLog{
		t:           t,
		db:          db,
		Subject:     sub,
		FactoryUser: factoryUser,
	}
}

func (c *AgentLog) IsValid() bool {
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

// Save persists the AgentLog instance to the database.
func (c *AgentLog) Save() *AgentLog {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting AgentLog")
	return c
}

// SetRequiredFields sets the required default fields for AgentLog.
func (c *AgentLog) SetRequiredFields() *AgentLog {
	if c.Subject.Log == "" {
		c.Subject.Log = "{}"
	}
	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *AgentLog) EnsureFKDeps() {
	// Only create user if not already set
	if c.Subject.UserRefID == "" {
		if !c.FactoryUser.IsValid() {
			c.FactoryUser = (&factory.Entity[*User]{}).New(c.t, c.db)
		}
		c.Subject.UserRefID = c.FactoryUser.Subject.ID
	}
}
