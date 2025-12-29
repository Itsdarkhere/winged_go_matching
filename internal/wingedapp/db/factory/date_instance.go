package factory

import (
	"context"
	"testing"
	"time"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/enums"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type DateInstance struct {
	t                  *testing.T
	db                 boil.ContextExecutor
	Subject            *pgmodel.DateInstance
	FactoryMatchResult *MatchResult
}

// New initializes a new DateInstance instance for testing.
func (c *DateInstance) New(t *testing.T, db boil.ContextExecutor) *DateInstance {
	sub := &pgmodel.DateInstance{}
	var factoryMatchResult *MatchResult

	if c != nil {
		if c.Subject != nil {
			sub = c.Subject
		}
		factoryMatchResult = c.FactoryMatchResult
	}

	return &DateInstance{
		t:                  t,
		db:                 db,
		Subject:            sub,
		FactoryMatchResult: factoryMatchResult,
	}
}

func (c *DateInstance) IsValid() bool {
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

// Save persists the DateInstance instance to the database.
func (c *DateInstance) Save() *DateInstance {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting DateInstance")
	return c
}

// SetRequiredFields sets the required default fields for DateInstance.
func (c *DateInstance) SetRequiredFields() *DateInstance {
	if c.Subject.ID == "" {
		c.Subject.ID = uuid.NewString()
	}
	if c.Subject.DecisionWindowEnd.IsZero() {
		c.Subject.DecisionWindowEnd = time.Now().UTC().Add(48 * time.Hour)
	}
	// Status is now a string enum with default
	if c.Subject.Status == "" {
		c.Subject.Status = string(enums.DateInstanceStatusProposed)
	}
	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *DateInstance) EnsureFKDeps() {
	// Only create match result if not already set
	if c.Subject.MatchResultRefID == "" {
		if !c.FactoryMatchResult.IsValid() {
			c.FactoryMatchResult = (&factory.Entity[*MatchResult]{}).New(c.t, c.db)
		}
		c.Subject.MatchResultRefID = c.FactoryMatchResult.Subject.ID
	}
}
