package factory

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/stretchr/testify/require"
)

type MatchConfig struct {
	t       *testing.T
	db      boil.ContextExecutor
	Subject *pgmodel.MatchConfig
}

// New initializes a new MatchConfig instance for testing.
func (c *MatchConfig) New(t *testing.T, db boil.ContextExecutor) *MatchConfig {
	sub := &pgmodel.MatchConfig{}
	if c != nil && c.Subject != nil {
		sub = c.Subject
	}

	return &MatchConfig{
		t:       t,
		db:      db,
		Subject: sub,
	}
}

func (c *MatchConfig) IsValid() bool {
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

// Save persists the MatchConfig instance to the database.
func (c *MatchConfig) Save() *MatchConfig {
	err := c.Subject.Insert(context.TODO(), c.db, boil.Infer())
	require.NoError(c.t, err, "expected no error inserting MatchConfig")
	return c
}

// SetRequiredFields sets the required default fields for MatchConfig.
// Note: Decimal fields (HeightMaleGreaterByCM, LocationRadiusKM, ScoreRange*)
// use database defaults when not explicitly set.
func (c *MatchConfig) SetRequiredFields() *MatchConfig {
	// Age range defaults
	if c.Subject.AgeRangeWomanOlderBy == 0 {
		c.Subject.AgeRangeWomanOlderBy = 5
	}
	if c.Subject.AgeRangeManOlderBy == 0 {
		c.Subject.AgeRangeManOlderBy = 10
	}

	// Location adaptive expansion default
	if len(c.Subject.LocationAdaptiveExpansion) == 0 {
		c.Subject.LocationAdaptiveExpansion = types.Int64Array{200, 350, 500}
	}

	// Drop hours defaults
	if len(c.Subject.DropHours) == 0 {
		c.Subject.DropHours = types.StringArray{"19:00", "20:00", "22:00", "23:00"}
	}
	if len(c.Subject.DropHoursUtc) == 0 {
		c.Subject.DropHoursUtc = types.StringArray{"GMT+3"}
	}

	// Stale chat defaults
	if c.Subject.StaleChatNudge == 0 {
		c.Subject.StaleChatNudge = 24
	}
	if c.Subject.StaleChatAgentSetup == 0 {
		c.Subject.StaleChatAgentSetup = 84
	}

	// Match expiration/block defaults
	if c.Subject.MatchExpirationHours == 0 {
		c.Subject.MatchExpirationHours = 72
	}
	if c.Subject.MatchBlockDeclined == 0 {
		c.Subject.MatchBlockDeclined = 168
	}
	if c.Subject.MatchBlockIgnored == 0 {
		c.Subject.MatchBlockIgnored = 168
	}
	if c.Subject.MatchBlockClosed == 0 {
		c.Subject.MatchBlockClosed = 168
	}

	return c
}

// EnsureFKDeps ensures that the foreign key dependencies are persisted.
func (c *MatchConfig) EnsureFKDeps() {
	// MatchConfig has no foreign key dependencies
}
