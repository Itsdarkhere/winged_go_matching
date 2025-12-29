package factory

import (
	"context"
	"testing"
	"time"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/require"
)

type User struct {
	t       *testing.T
	db      boil.ContextExecutor
	Subject *pgmodel.User
}

// New initializes a new User instance for testing.
func (c *User) New(t *testing.T, db boil.ContextExecutor) *User {
	sub := &pgmodel.User{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	return &User{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

func (c *User) IsValid() bool {
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

// Save persists the User instance to the database.
func (c *User) Save() *User {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting UserFromToken")
	return c
}

// SetRequiredFields sets the required default fields for User.
func (c *User) SetRequiredFields() *User {
	if c.Subject.Email == "" {
		c.Subject.Email = randomEmail()
	}
	if !c.Subject.FirstName.Valid {
		c.Subject.FirstName = null.StringFrom("TestUser")
	}
	if !c.Subject.Birthday.Valid {
		// Default to 25 years old
		c.Subject.Birthday = null.TimeFrom(time.Now().AddDate(-25, 0, 0))
	}
	if !c.Subject.Gender.Valid {
		c.Subject.Gender = null.StringFrom("Female")
	}
	if !c.Subject.Sexuality.Valid {
		c.Subject.Sexuality = null.StringFrom("Straight")
	}

	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *User) EnsureFKDeps() {

}
