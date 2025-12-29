package factory

import (
	"context"
	"testing"
	"time"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type DateInstanceProposal struct {
	t                   *testing.T
	db                  boil.ContextExecutor
	Subject             *pgmodel.DateInstanceProposal
	FactoryDateInstance *DateInstance
	FactorySuggestedBy  *User
}

// New initializes a new DateInstanceProposal instance for testing.
func (c *DateInstanceProposal) New(t *testing.T, db boil.ContextExecutor) *DateInstanceProposal {
	sub := &pgmodel.DateInstanceProposal{}
	var factoryDateInstance *DateInstance
	var factorySuggestedBy *User

	if c != nil {
		if c.Subject != nil {
			sub = c.Subject
		}
		factoryDateInstance = c.FactoryDateInstance
		factorySuggestedBy = c.FactorySuggestedBy
	}

	return &DateInstanceProposal{
		t:                   t,
		db:                  db,
		Subject:             sub,
		FactoryDateInstance: factoryDateInstance,
		FactorySuggestedBy:  factorySuggestedBy,
	}
}

func (c *DateInstanceProposal) IsValid() bool {
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

// Save persists the DateInstanceProposal instance to the database.
func (c *DateInstanceProposal) Save() *DateInstanceProposal {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting DateInstanceProposal")
	return c
}

// SetRequiredFields sets the required default fields for DateInstanceProposal.
func (c *DateInstanceProposal) SetRequiredFields() *DateInstanceProposal {
	if c.Subject.ID == "" {
		c.Subject.ID = uuid.NewString()
	}
	if c.Subject.ProposedTime.IsZero() {
		c.Subject.ProposedTime = time.Now().UTC().Add(48 * time.Hour)
	}
	if c.Subject.Status == "" {
		c.Subject.Status = "pending"
	}
	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *DateInstanceProposal) EnsureFKDeps() {
	// Only create date instance if not already set
	if c.Subject.DateInstanceRefID == "" {
		if !c.FactoryDateInstance.IsValid() {
			c.FactoryDateInstance = (&factory.Entity[*DateInstance]{}).New(c.t, c.db)
		}
		c.Subject.DateInstanceRefID = c.FactoryDateInstance.Subject.ID
	}

	// Only create suggested by user if not already set
	if c.Subject.SuggestedByRefID == "" {
		if !c.FactorySuggestedBy.IsValid() {
			c.FactorySuggestedBy = (&factory.Entity[*User]{}).New(c.t, c.db)
		}
		c.Subject.SuggestedByRefID = c.FactorySuggestedBy.Subject.ID
	}
}
