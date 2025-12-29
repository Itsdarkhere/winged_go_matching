package factory

import (
	"context"
	"testing"
	"time"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/factory"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/stretchr/testify/require"
)

type Transcript struct {
	t       *testing.T
	db      boil.ContextExecutor
	Subject *aipgmodel.Transcript
}

// New initializes a new Transcript instance for testing.
func (c *Transcript) New(t *testing.T, db boil.ContextExecutor) *Transcript {
	sub := &aipgmodel.Transcript{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	return &Transcript{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

func (c *Transcript) IsValid() bool {
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

// Save persists the Transcript instance to the database.
func (c *Transcript) Save() *Transcript {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting Transcript")
	return c
}

// SetRequiredFields sets the required default fields for Transcript.
func (c *Transcript) SetRequiredFields() *Transcript {
	if c.Subject.UserID == "" {
		c.Subject.UserID = factory.RandomString()
	}
	if c.Subject.WebhookType == "" {
		c.Subject.WebhookType = factory.RandomString()
	}
	if c.Subject.EventTimestamp == 0 {
		c.Subject.EventTimestamp = int(time.Now().Unix())
	}
	if c.Subject.CreatedAt.IsZero() {
		c.Subject.CreatedAt = time.Now()
	}
	if len(c.Subject.RawPayload) == 0 {
		c.Subject.RawPayload = types.JSON([]byte(`{}`))
	}

	return c
}

func (c *Transcript) EnsureFKDeps() {

}
