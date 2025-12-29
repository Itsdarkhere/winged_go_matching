package factory

import (
	"context"
	"fmt"
	"testing"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/enums"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/require"
)

type GeneralAIContext struct {
	t       *testing.T
	db      boil.ContextExecutor
	Subject *pgmodel.GeneralAIContext
}

func (c *GeneralAIContext) New(t *testing.T, db boil.ContextExecutor) *GeneralAIContext {
	sub := &pgmodel.GeneralAIContext{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	sub.ID = "" // ensure PK always empty for new
	return &GeneralAIContext{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

func (c *GeneralAIContext) IsValid() bool {
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

func (c *GeneralAIContext) Save() *GeneralAIContext {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting general ai context set")
	return c
}

func (c *GeneralAIContext) EnsureFKDeps() {
	// No FK deps - ai_context_type is now a string enum
}

// SetRequiredFields sets the required default fields for GeneralAIContext.
func (c *GeneralAIContext) SetRequiredFields() *GeneralAIContext {
	if c.Subject.Context == "" {
		fmt.Println("=== I am setting a random string")
		c.Subject.Context = RandomString()
	}
	// AiContextType is now a string enum with default
	if c.Subject.AiContextType == "" {
		c.Subject.AiContextType = string(enums.GeneralAIContextYourAgent)
	}
	return c
}

// NewWithOverrides creates and inserts a GeneralAIContext with optional field overrides.
func (c *GeneralAIContext) NewWithOverrides(
	t *testing.T,
	db boil.ContextExecutor,
	overrides func(subject *pgmodel.GeneralAIContext),
) *GeneralAIContext {
	inst := c.
		New(t, db).
		SetRequiredFields()

	if overrides != nil {
		overrides(inst.Subject)
	}

	require.NoError(t, inst.Subject.Insert(context.Background(), db, boil.Infer()))
	return inst
}
