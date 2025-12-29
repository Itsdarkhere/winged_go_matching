package factory

import (
	"context"
	"testing"
	"time"

	supabasepgmodel "wingedapp/pgtester/internal/wingedapp/supabase/db/supabasepgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// SupabaseAuthUser is a test factory for creating supabase auth.users records.
type SupabaseAuthUser struct {
	t       *testing.T
	db      boil.ContextExecutor
	Subject *supabasepgmodel.User
}

// New initializes a new SupabaseAuthUser instance for testing.
func (c *SupabaseAuthUser) New(t *testing.T, db boil.ContextExecutor) *SupabaseAuthUser {
	sub := &supabasepgmodel.User{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	return &SupabaseAuthUser{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

// IsValid checks if the SupabaseAuthUser instance is valid for use in tests.
func (c *SupabaseAuthUser) IsValid() bool {
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

// Save persists the SupabaseAuthUser instance to the database.
func (c *SupabaseAuthUser) Save() *SupabaseAuthUser {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting SupabaseAuthUser")
	return c
}

// SetRequiredFields sets the required default fields for SupabaseAuthUser.
func (c *SupabaseAuthUser) SetRequiredFields() *SupabaseAuthUser {
	if c.Subject.ID == "" {
		c.Subject.ID = uuid.NewString()
	}
	if !c.Subject.Email.Valid || c.Subject.Email.String == "" {
		c.Subject.Email = null.StringFrom(uuid.NewString() + "@example.com")
	}
	if !c.Subject.Aud.Valid {
		c.Subject.Aud = null.StringFrom("authenticated")
	}
	if !c.Subject.Role.Valid {
		c.Subject.Role = null.StringFrom("authenticated")
	}
	if !c.Subject.CreatedAt.Valid {
		c.Subject.CreatedAt = null.TimeFrom(time.Now())
	}
	if !c.Subject.UpdatedAt.Valid {
		c.Subject.UpdatedAt = null.TimeFrom(time.Now())
	}
	if !c.Subject.EmailConfirmedAt.Valid {
		c.Subject.EmailConfirmedAt = null.TimeFrom(time.Now())
	}
	if !c.Subject.ConfirmedAt.Valid {
		c.Subject.ConfirmedAt = null.TimeFrom(time.Now())
	}

	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *SupabaseAuthUser) EnsureFKDeps() {
	// No FK deps for supabase auth users
}

// WithEmail sets the email for the user.
func (c *SupabaseAuthUser) WithEmail(email string) *SupabaseAuthUser {
	c.Subject.Email = null.StringFrom(email)
	return c
}

// WithID sets the ID for the user.
func (c *SupabaseAuthUser) WithID(id string) *SupabaseAuthUser {
	c.Subject.ID = id
	return c
}

// WithPhone sets the phone for the user.
func (c *SupabaseAuthUser) WithPhone(phone string) *SupabaseAuthUser {
	c.Subject.Phone = null.StringFrom(phone)
	c.Subject.PhoneConfirmedAt = null.TimeFrom(time.Now())
	return c
}

// WithEncryptedPassword sets the encrypted password for the user.
func (c *SupabaseAuthUser) WithEncryptedPassword(password string) *SupabaseAuthUser {
	c.Subject.EncryptedPassword = null.StringFrom(password)
	return c
}

// WithRawUserMetaData sets the raw user meta data for the user.
func (c *SupabaseAuthUser) WithRawUserMetaData(data []byte) *SupabaseAuthUser {
	c.Subject.RawUserMetaData = null.JSONFrom(data)
	return c
}

// WithRawAppMetaData sets the raw app meta data for the user.
func (c *SupabaseAuthUser) WithRawAppMetaData(data []byte) *SupabaseAuthUser {
	c.Subject.RawAppMetaData = null.JSONFrom(data)
	return c
}
