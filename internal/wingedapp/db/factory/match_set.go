package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/stretchr/testify/require"
)

type MatchSet struct {
	t       *testing.T
	db      boil.ContextExecutor
	Subject *pgmodel.MatchSet
}

// New initializes a new MatchSet instance for testing.
func (c *MatchSet) New(t *testing.T, db boil.ContextExecutor) *MatchSet {
	sub := &pgmodel.MatchSet{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	return &MatchSet{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

func (c *MatchSet) IsValid() bool {
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

// Save persists the MatchSet instance to the database.
func (c *MatchSet) Save() *MatchSet {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting MatchSet")
	return c
}

// SetRequiredFields sets the required default fields for MatchSet.
func (c *MatchSet) SetRequiredFields() *MatchSet {
	if c.Subject.Name == "" {
		c.Subject.Name = RandomString()
	}

	if c.Subject.NumberOfParticipants == 0 {
		c.Subject.NumberOfParticipants = 2
	}

	if len(c.Subject.MatchConfiguration) == 0 {
		// default match configuration
		c.Subject.MatchConfiguration = types.JSON(marshal(c.t, map[string]interface{}{
			"type":    "standard",
			"version": "1.0",
		}))
	}

	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *MatchSet) EnsureFKDeps() {

}
