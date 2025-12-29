package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/enums"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/require"
)

type UserInviteCode struct {
	t       *testing.T
	db      boil.ContextExecutor
	Subject *pgmodel.UserInviteCode
}

// New initializes a new UserInviteCode instance for testing.
func (c *UserInviteCode) New(t *testing.T, db boil.ContextExecutor) *UserInviteCode {
	sub := &pgmodel.UserInviteCode{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	return &UserInviteCode{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

// IsValid checks if the UserInviteCode instance is valid for use in tests.
func (c *UserInviteCode) IsValid() bool {
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

// EnsureFKDeps ensures that any foreign key dependencies are persisted.
func (c *UserInviteCode) EnsureFKDeps() {
	// No FK deps - InviteCodeType is now a string enum
}

// Save persists the UserInviteCode instance to the database.
func (c *UserInviteCode) Save() *UserInviteCode {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting UserInviteCodeID")
	return c
}

// SetRequiredFields sets the required default fields for UserInviteCode.
func (c *UserInviteCode) SetRequiredFields() *UserInviteCode {
	if c.Subject.InviteCode == "" {
		c.Subject.InviteCode = RandomDigits(6)
	}
	// InviteCodeType is now a string enum
	if c.Subject.InviteCodeType == "" {
		c.Subject.InviteCodeType = string(enums.UserInviteCodeReferral)
	}
	if c.Subject.ReferralSource == "" {
		c.Subject.ReferralSource = RandomString()
	}
	return c
}
