package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/require"
)

type Profile struct {
	t           *testing.T
	db          boil.ContextExecutor
	Subject     *aipgmodel.Profile
	FactoryUser *User
}

// New initializes a new Profile instance for testing.
func (c *Profile) New(t *testing.T, db boil.ContextExecutor) *Profile {
	sub := &aipgmodel.Profile{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	return &Profile{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

func (c *Profile) IsValid() bool {
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

// Save persists the Profile instance to the database.
func (c *Profile) Save() *Profile {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting UserFromToken")
	return c
}

//// SetRequiredFields sets the required default fields for ProfileDeprecated.
//func (c *ProfileDeprecated) SetRequiredFields() *ProfileDeprecated {
//	if c.Subject.UserID == "" {
//		c.Subject.UserID = c.FactoryUser.Subject.Code
//	}
//	return c
//}
//
//// EnsureFKDeps ensures that the foreign key dependencies are persisted.
//func (c *User) EnsureFKDeps() {
//
//}
