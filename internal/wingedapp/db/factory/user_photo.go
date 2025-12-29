package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type UserPhoto struct {
	t           *testing.T
	db          boil.ContextExecutor
	Subject     *pgmodel.UserPhoto
	FactoryUser *User
}

// New initializes a new UserPhoto instance for testing.
func (c *UserPhoto) New(t *testing.T, db boil.ContextExecutor) *UserPhoto {
	sub := &pgmodel.UserPhoto{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	sub.ID = "" // ensure PK always empty for new
	return &UserPhoto{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

func (c *UserPhoto) IsValid() bool {
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

// Save persists the UserPhoto instance to the database.
func (c *UserPhoto) Save() *UserPhoto {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting UserPhoto")
	return c
}

// SetRequiredFields sets the required default fields for UserPhoto.
func (c *UserPhoto) SetRequiredFields() *UserPhoto {
	if c.Subject.UserID == "" {
		c.Subject.UserID = c.FactoryUser.Subject.ID
	}
	if c.Subject.Bucket == "" {
		c.Subject.Bucket = uuid.NewString()
	}
	if c.Subject.Key == "" {
		c.Subject.Key = uuid.NewString()
	}
	if c.Subject.OrderNo == 0 {
		c.Subject.OrderNo = 1
	}
	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *UserPhoto) EnsureFKDeps() {
	if !c.FactoryUser.IsValid() && c.Subject.UserID == "" {
		c.FactoryUser = (&factory.Entity[*User]{}).New(c.t, c.db)
	}
}
