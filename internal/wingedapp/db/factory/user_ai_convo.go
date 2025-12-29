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

type UserAIConvo struct {
	t           *testing.T
	db          boil.ContextExecutor
	Subject     *pgmodel.UserAIConvo
	FactoryUser *User
}

// New initializes a new UserAIConvo instance for testing.
func (c *UserAIConvo) New(t *testing.T, db boil.ContextExecutor) *UserAIConvo {
	sub := &pgmodel.UserAIConvo{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	sub.ID = "" // ensure PK always empty for new
	return &UserAIConvo{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

func (c *UserAIConvo) IsValid() bool {
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

// Save persists the UserAIConvo instance to the database.
func (c *UserAIConvo) Save() *UserAIConvo {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting UserAIConvo")
	return c
}

// SetRequiredFields sets the required default fields for UserAIConvo.
func (c *UserAIConvo) SetRequiredFields() *UserAIConvo {
	// AiConvoType is now a string enum
	if c.Subject.AiConvoType == "" {
		c.Subject.AiConvoType = string(enums.AIConvoAI)
	}
	if c.Subject.PromptResponseID == "" {
		c.Subject.PromptResponseID = RandomString()
	}
	if c.Subject.Message == "" {
		c.Subject.Message = RandomString()
	}
	if c.Subject.Response == "" {
		c.Subject.Response = RandomString()
	}
	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *UserAIConvo) EnsureFKDeps() {
	if !c.FactoryUser.IsValid() && c.Subject.UserID == "" {
		c.FactoryUser = (&factory.Entity[*User]{}).New(c.t, c.db)
	}
	if c.Subject.UserID == "" && c.FactoryUser.IsValid() {
		c.Subject.UserID = c.FactoryUser.Subject.ID
	}
}
