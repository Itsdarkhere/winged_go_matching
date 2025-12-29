package factory

import (
	"context"
	"testing"
	"time"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/enums"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/require"
)

type MatchResult struct {
	t                *testing.T
	db               boil.ContextExecutor
	Subject          *pgmodel.MatchResult
	FactoryMatchSet  *MatchSet
	FactoryUserA     *User
	FactoryUserB     *User
	FactoryInitiator *User
	FactoryReceiver  *User
}

// New initializes a new MatchResult instance for testing.
func (c *MatchResult) New(t *testing.T, db boil.ContextExecutor) *MatchResult {
	sub := &pgmodel.MatchResult{}
	var factoryMatchSet *MatchSet
	var factoryUserA *User
	var factoryUserB *User
	var factoryInitiator *User
	var factoryReceiver *User

	if c != nil {
		if c.Subject != nil {
			sub = c.Subject
		}
		factoryMatchSet = c.FactoryMatchSet
		factoryUserA = c.FactoryUserA
		factoryUserB = c.FactoryUserB
		factoryInitiator = c.FactoryInitiator
		factoryReceiver = c.FactoryReceiver
	}

	return &MatchResult{
		t:                t,
		db:               db,
		Subject:          sub,
		FactoryMatchSet:  factoryMatchSet,
		FactoryUserA:     factoryUserA,
		FactoryUserB:     factoryUserB,
		FactoryInitiator: factoryInitiator,
		FactoryReceiver:  factoryReceiver,
	}
}

func (c *MatchResult) IsValid() bool {
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

// Save persists the MatchResult instance to the database.
func (c *MatchResult) Save() *MatchResult {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting MatchResult")
	return c
}

// SetRequiredFields sets the required default fields for MatchResult.
func (c *MatchResult) SetRequiredFields() *MatchResult {
	// String enum fields with defaults
	if c.Subject.MatchStatus == "" {
		c.Subject.MatchStatus = string(enums.MatchStatusActive)
	}
	if c.Subject.UserAAction == "" {
		c.Subject.UserAAction = string(enums.MatchUserActionPending)
	}
	if c.Subject.UserBAction == "" {
		c.Subject.UserBAction = string(enums.MatchUserActionPending)
	}
	if !c.Subject.MatchLifecycleStatus.Valid {
		c.Subject.MatchLifecycleStatus.SetValid(string(enums.MatchLifecycleStatusQueued))
	}
	// Set dropped_ts when is_dropped is true
	if c.Subject.IsDropped && !c.Subject.DroppedTS.Valid {
		c.Subject.DroppedTS = null.TimeFrom(time.Now())
	}
	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *MatchResult) EnsureFKDeps() {
	// Only create match set if not already set
	if c.Subject.MatchSetRefID == "" {
		if !c.FactoryMatchSet.IsValid() {
			c.FactoryMatchSet = (&factory.Entity[*MatchSet]{}).New(c.t, c.db)
		}
		c.Subject.MatchSetRefID = c.FactoryMatchSet.Subject.ID
	}

	// Only create user A if not already set
	if c.Subject.UserARefID == "" {
		if !c.FactoryUserA.IsValid() {
			c.FactoryUserA = (&factory.Entity[*User]{}).New(c.t, c.db)
		}
		c.Subject.UserARefID = c.FactoryUserA.Subject.ID
	}

	// Only create user B if not already set
	if c.Subject.UserBRefID == "" {
		if !c.FactoryUserB.IsValid() {
			c.FactoryUserB = (&factory.Entity[*User]{}).New(c.t, c.db)
		}
		c.Subject.UserBRefID = c.FactoryUserB.Subject.ID
	}

	// Only create initiator if not already set (default to user A)
	if c.Subject.InitiatorUserRefID == "" {
		if !c.FactoryInitiator.IsValid() {
			// Default initiator to user A if not specified
			if c.FactoryUserA.IsValid() {
				c.Subject.InitiatorUserRefID = c.FactoryUserA.Subject.ID
			} else {
				c.FactoryInitiator = (&factory.Entity[*User]{}).New(c.t, c.db)
				c.Subject.InitiatorUserRefID = c.FactoryInitiator.Subject.ID
			}
		} else {
			c.Subject.InitiatorUserRefID = c.FactoryInitiator.Subject.ID
		}
	}

	// Only create receiver if not already set (default to user B)
	if c.Subject.ReceiverUserRefID == "" {
		if !c.FactoryReceiver.IsValid() {
			// Default receiver to user B if not specified
			if c.FactoryUserB.IsValid() {
				c.Subject.ReceiverUserRefID = c.FactoryUserB.Subject.ID
			} else {
				c.FactoryReceiver = (&factory.Entity[*User]{}).New(c.t, c.db)
				c.Subject.ReceiverUserRefID = c.FactoryReceiver.Subject.ID
			}
		} else {
			c.Subject.ReceiverUserRefID = c.FactoryReceiver.Subject.ID
		}
	}
}
