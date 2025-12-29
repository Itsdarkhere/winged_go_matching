package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type Lovestory struct {
	t       *testing.T
	db      boil.ContextExecutor
	Subject *aipgmodel.Lovestory
}

// New initializes a new Lovestory instance for testing.
func (c *Lovestory) New(t *testing.T, db boil.ContextExecutor) *Lovestory {
	sub := &aipgmodel.Lovestory{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	return &Lovestory{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

func (c *Lovestory) IsValid() bool {
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

// Save persists the Lovestory instance to the database.
func (c *Lovestory) Save() *Lovestory {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting Lovestory")
	return c
}

// SetRequiredFields sets the required default fields for Lovestory.
func (c *Lovestory) SetRequiredFields() *Lovestory {
	if c.Subject.MatchResultRefID == "" {
		c.Subject.MatchResultRefID = uuid.New().String()
	}
	if c.Subject.UserARefID == "" {
		c.Subject.UserARefID = uuid.New().String()
	}
	if c.Subject.UserBRefID == "" {
		c.Subject.UserBRefID = uuid.New().String()
	}
	if c.Subject.Status == "" {
		c.Subject.Status = "pending"
	}
	return c
}

func (c *Lovestory) EnsureFKDeps() {
	// No FK dependencies - match_result_ref_id is just a UUID reference, not an FK constraint
}
