package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type Venue struct {
	t       *testing.T
	db      boil.ContextExecutor
	Subject *pgmodel.Venue
}

// New initializes a new Venue instance for testing.
func (c *Venue) New(t *testing.T, db boil.ContextExecutor) *Venue {
	sub := &pgmodel.Venue{}

	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	return &Venue{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

func (c *Venue) IsValid() bool {
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

// Save persists the Venue instance to the database.
func (c *Venue) Save() *Venue {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting Venue")
	return c
}

// SetRequiredFields sets the required default fields for Venue.
func (c *Venue) SetRequiredFields() *Venue {
	if c.Subject.ID == "" {
		c.Subject.ID = uuid.NewString()
	}
	if c.Subject.ExternalProvider == "" {
		c.Subject.ExternalProvider = "google_places"
	}
	if c.Subject.ExternalID == "" {
		c.Subject.ExternalID = "place_" + uuid.NewString()[:8]
	}
	if c.Subject.Name == "" {
		c.Subject.Name = "Test Venue " + uuid.NewString()[:8]
	}
	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *Venue) EnsureFKDeps() {
	// No FK dependencies - venue is standalone
}
