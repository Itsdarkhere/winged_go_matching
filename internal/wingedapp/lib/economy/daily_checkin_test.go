package economy_test

import (
	"context"
	"testing"
	"time"

	basefactory "wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/economy"
	"wingedapp/pgtester/internal/wingedapp/lib/economy/store"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCasePerformCheckin struct {
	name            string
	setup           func(th *testsuite.Helper) (userID string)
	extraAssertions func(th *testsuite.Helper, userID string, result *economy.CheckinResult, err error)
}

func performCheckinTestCases() []testCasePerformCheckin {
	return []testCasePerformCheckin{
		{
			name: "success-first-checkin-starts-streak-at-1",
			setup: func(th *testsuite.Helper) string {
				user := (&basefactory.Entity[*factory.User]{}).New(th.T, th.BackendAppDb())
				createUserTotals(th, user.Subject.ID, 0, null.Time{}, 0, 0)
				return user.Subject.ID
			},
			extraAssertions: func(th *testsuite.Helper, userID string, result *economy.CheckinResult, err error) {
				require.NoError(th.T, err)
				assert.Equal(th.T, 1, result.NewStreak)
				assert.False(th.T, result.MilestoneReached)
				assert.Equal(th.T, 0, result.WingsAwarded)

				// Verify DB state
				totals := getTestUserTotals(th, userID)
				assert.Equal(th.T, 1, totals.StreakCurrentDays)
				assert.Equal(th.T, 1, totals.StreakLongestDays)
				assert.True(th.T, totals.StreakLastDate.Valid)
				assert.Equal(th.T, 0, totals.TotalWings) // No wings for regular check-in
			},
		},
		{
			name: "success-consecutive-day-increments-streak",
			setup: func(th *testsuite.Helper) string {
				user := (&basefactory.Entity[*factory.User]{}).New(th.T, th.BackendAppDb())
				yesterday := time.Now().UTC().AddDate(0, 0, -1).Truncate(24 * time.Hour)
				createUserTotals(th, user.Subject.ID, 0, null.TimeFrom(yesterday), 5, 5)
				return user.Subject.ID
			},
			extraAssertions: func(th *testsuite.Helper, userID string, result *economy.CheckinResult, err error) {
				require.NoError(th.T, err)
				assert.Equal(th.T, 6, result.NewStreak)
				assert.False(th.T, result.MilestoneReached)
				assert.Equal(th.T, 0, result.WingsAwarded)

				totals := getTestUserTotals(th, userID)
				assert.Equal(th.T, 6, totals.StreakCurrentDays)
				assert.Equal(th.T, 6, totals.StreakLongestDays)
			},
		},
		{
			name: "success-missed-day-resets-streak-to-1",
			setup: func(th *testsuite.Helper) string {
				user := (&basefactory.Entity[*factory.User]{}).New(th.T, th.BackendAppDb())
				twoDaysAgo := time.Now().UTC().AddDate(0, 0, -2).Truncate(24 * time.Hour)
				createUserTotals(th, user.Subject.ID, 0, null.TimeFrom(twoDaysAgo), 10, 15)
				return user.Subject.ID
			},
			extraAssertions: func(th *testsuite.Helper, userID string, result *economy.CheckinResult, err error) {
				require.NoError(th.T, err)
				assert.Equal(th.T, 1, result.NewStreak)
				assert.False(th.T, result.MilestoneReached)

				totals := getTestUserTotals(th, userID)
				assert.Equal(th.T, 1, totals.StreakCurrentDays)
				assert.Equal(th.T, 15, totals.StreakLongestDays) // Longest preserved
			},
		},
		{
			name: "success-7-day-milestone-awards-2-wings",
			setup: func(th *testsuite.Helper) string {
				user := (&basefactory.Entity[*factory.User]{}).New(th.T, th.BackendAppDb())
				yesterday := time.Now().UTC().AddDate(0, 0, -1).Truncate(24 * time.Hour)
				createUserTotals(th, user.Subject.ID, 10, null.TimeFrom(yesterday), 6, 6)
				return user.Subject.ID
			},
			extraAssertions: func(th *testsuite.Helper, userID string, result *economy.CheckinResult, err error) {
				require.NoError(th.T, err)
				assert.Equal(th.T, 7, result.NewStreak)
				assert.True(th.T, result.MilestoneReached)
				assert.Equal(th.T, 7, result.MilestoneType)
				assert.Equal(th.T, 2, result.WingsAwarded)

				totals := getTestUserTotals(th, userID)
				assert.Equal(th.T, 7, totals.StreakCurrentDays)
				assert.Equal(th.T, 12, totals.TotalWings) // 10 + 2

				// Verify transaction was created
				txns := getTestTransactionsByUser(th, userID)
				require.Len(th.T, txns, 1)
				assert.Equal(th.T, 2, txns[0].Amount)
				assert.Equal(th.T, string(economy.ActionStreak7Day), txns[0].ActionLogType)
			},
		},
		{
			name: "success-30-day-milestone-awards-6-wings",
			setup: func(th *testsuite.Helper) string {
				user := (&basefactory.Entity[*factory.User]{}).New(th.T, th.BackendAppDb())
				yesterday := time.Now().UTC().AddDate(0, 0, -1).Truncate(24 * time.Hour)
				createUserTotals(th, user.Subject.ID, 20, null.TimeFrom(yesterday), 29, 29)
				return user.Subject.ID
			},
			extraAssertions: func(th *testsuite.Helper, userID string, result *economy.CheckinResult, err error) {
				require.NoError(th.T, err)
				assert.Equal(th.T, 30, result.NewStreak)
				assert.True(th.T, result.MilestoneReached)
				assert.Equal(th.T, 30, result.MilestoneType)
				assert.Equal(th.T, 6, result.WingsAwarded)

				totals := getTestUserTotals(th, userID)
				assert.Equal(th.T, 30, totals.StreakCurrentDays)
				assert.Equal(th.T, 26, totals.TotalWings) // 20 + 6

				// Verify transaction was created
				txns := getTestTransactionsByUser(th, userID)
				require.Len(th.T, txns, 1)
				assert.Equal(th.T, 6, txns[0].Amount)
				assert.Equal(th.T, string(economy.ActionStreak30Day), txns[0].ActionLogType)
			},
		},
		{
			name: "error-already-checked-in-today",
			setup: func(th *testsuite.Helper) string {
				user := (&basefactory.Entity[*factory.User]{}).New(th.T, th.BackendAppDb())
				today := time.Now().UTC().Truncate(24 * time.Hour)
				createUserTotals(th, user.Subject.ID, 0, null.TimeFrom(today), 3, 3)
				return user.Subject.ID
			},
			extraAssertions: func(th *testsuite.Helper, userID string, result *economy.CheckinResult, err error) {
				require.Error(th.T, err)
				assert.ErrorIs(th.T, err, economy.ErrAlreadyCheckedInToday)
				assert.Nil(th.T, result)
			},
		},
		{
			name: "success-new-longest-streak",
			setup: func(th *testsuite.Helper) string {
				user := (&basefactory.Entity[*factory.User]{}).New(th.T, th.BackendAppDb())
				yesterday := time.Now().UTC().AddDate(0, 0, -1).Truncate(24 * time.Hour)
				createUserTotals(th, user.Subject.ID, 0, null.TimeFrom(yesterday), 5, 5)
				return user.Subject.ID
			},
			extraAssertions: func(th *testsuite.Helper, userID string, result *economy.CheckinResult, err error) {
				require.NoError(th.T, err)
				assert.True(th.T, result.IsNewLongestStreak)

				totals := getTestUserTotals(th, userID)
				assert.Equal(th.T, 6, totals.StreakLongestDays)
			},
		},
	}
}

func TestDailyCheckinLogic_PerformCheckin(t *testing.T) {
	for _, tt := range performCheckinTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testSuite := testsuite.New(t)
			t.Cleanup(testSuite.UseBackendDB())

			ctx := context.Background()
			userID := tt.setup(testSuite)

			logic := createTestDailyCheckinLogic(t)
			result, err := logic.PerformCheckin(ctx, testSuite.BackendAppDb(), userID)

			tt.extraAssertions(testSuite, userID, result, err)
		})
	}
}

type testCaseGetStatus struct {
	name            string
	setup           func(th *testsuite.Helper) (userID string)
	extraAssertions func(th *testsuite.Helper, status *economy.CheckinStatus, err error)
}

func getStatusTestCases() []testCaseGetStatus {
	return []testCaseGetStatus{
		{
			name: "success-not-checked-in-today",
			setup: func(th *testsuite.Helper) string {
				user := (&basefactory.Entity[*factory.User]{}).New(th.T, th.BackendAppDb())
				yesterday := time.Now().UTC().AddDate(0, 0, -1).Truncate(24 * time.Hour)
				createUserTotals(th, user.Subject.ID, 0, null.TimeFrom(yesterday), 5, 10)
				return user.Subject.ID
			},
			extraAssertions: func(th *testsuite.Helper, status *economy.CheckinStatus, err error) {
				require.NoError(th.T, err)
				assert.False(th.T, status.CheckedInToday)
				assert.Equal(th.T, 5, status.StreakCurrentDays)
				assert.Equal(th.T, 10, status.StreakLongestDays)
				assert.Equal(th.T, 7, status.NextMilestone)
				assert.Equal(th.T, 2, status.DaysToMilestone)
				assert.Equal(th.T, 2, status.MilestoneWings)
			},
		},
		{
			name: "success-checked-in-today",
			setup: func(th *testsuite.Helper) string {
				user := (&basefactory.Entity[*factory.User]{}).New(th.T, th.BackendAppDb())
				today := time.Now().UTC().Truncate(24 * time.Hour)
				createUserTotals(th, user.Subject.ID, 0, null.TimeFrom(today), 3, 5)
				return user.Subject.ID
			},
			extraAssertions: func(th *testsuite.Helper, status *economy.CheckinStatus, err error) {
				require.NoError(th.T, err)
				assert.True(th.T, status.CheckedInToday)
				assert.Equal(th.T, 3, status.StreakCurrentDays)
				assert.Equal(th.T, 7, status.NextMilestone)
				assert.Equal(th.T, 4, status.DaysToMilestone)
			},
		},
		{
			name: "success-between-milestones",
			setup: func(th *testsuite.Helper) string {
				user := (&basefactory.Entity[*factory.User]{}).New(th.T, th.BackendAppDb())
				today := time.Now().UTC().Truncate(24 * time.Hour)
				createUserTotals(th, user.Subject.ID, 0, null.TimeFrom(today), 15, 15)
				return user.Subject.ID
			},
			extraAssertions: func(th *testsuite.Helper, status *economy.CheckinStatus, err error) {
				require.NoError(th.T, err)
				assert.Equal(th.T, 30, status.NextMilestone)
				assert.Equal(th.T, 15, status.DaysToMilestone)
				assert.Equal(th.T, 6, status.MilestoneWings)
			},
		},
		{
			name: "success-past-all-milestones",
			setup: func(th *testsuite.Helper) string {
				user := (&basefactory.Entity[*factory.User]{}).New(th.T, th.BackendAppDb())
				today := time.Now().UTC().Truncate(24 * time.Hour)
				createUserTotals(th, user.Subject.ID, 0, null.TimeFrom(today), 45, 45)
				return user.Subject.ID
			},
			extraAssertions: func(th *testsuite.Helper, status *economy.CheckinStatus, err error) {
				require.NoError(th.T, err)
				assert.Equal(th.T, 0, status.NextMilestone)
				assert.Equal(th.T, 0, status.DaysToMilestone)
				assert.Equal(th.T, 0, status.MilestoneWings)
			},
		},
	}
}

func TestDailyCheckinLogic_GetStatus(t *testing.T) {
	for _, tt := range getStatusTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testSuite := testsuite.New(t)
			t.Cleanup(testSuite.UseBackendDB())

			ctx := context.Background()
			userID := tt.setup(testSuite)

			logic := createTestDailyCheckinLogic(t)
			status, err := logic.GetStatus(ctx, testSuite.BackendAppDb(), userID)

			tt.extraAssertions(testSuite, status, err)
		})
	}
}

// createTestDailyCheckinLogic creates a DailyCheckinLogic for tests.
func createTestDailyCheckinLogic(t *testing.T) *economy.DailyCheckinLogic {
	t.Helper()
	logger := applog.NewLogrus("test")
	stores := store.NewEconomyStores(logger)

	logic, err := economy.NewDailyCheckinLogic(logger, stores.UserTotalsStore, stores.ActionLogStore, stores.TransactionStore)
	require.NoError(t, err)
	return logic
}

// createUserTotals creates a WingsEcnUserTotal directly in DB with streak fields.
func createUserTotals(th *testsuite.Helper, userID string, wings int, streakLastDate null.Time, streakCurrent, streakLongest int) {
	th.T.Helper()
	ut := pgmodel.WingsEcnUserTotal{
		UserRefID:         userID,
		TotalWings:        wings,
		StreakLastDate:    streakLastDate,
		StreakCurrentDays: streakCurrent,
		StreakLongestDays: streakLongest,
	}
	err := ut.Insert(context.Background(), th.BackendAppDb(), boil.Infer())
	require.NoError(th.T, err)
}

func getTestUserTotals(th *testsuite.Helper, userID string) *pgmodel.WingsEcnUserTotal {
	th.T.Helper()
	ctx := context.Background()
	totals, err := pgmodel.WingsEcnUserTotals(
		pgmodel.WingsEcnUserTotalWhere.UserRefID.EQ(userID),
	).One(ctx, th.BackendAppDb())
	require.NoError(th.T, err)
	return totals
}

func getTestTransactionsByUser(th *testsuite.Helper, userID string) pgmodel.WingsEcnTransactionSlice {
	th.T.Helper()
	ctx := context.Background()
	txns, err := pgmodel.WingsEcnTransactions(
		pgmodel.WingsEcnTransactionWhere.UserRefID.EQ(userID),
		qm.OrderBy(pgmodel.WingsEcnTransactionColumns.CreatedDate+" DESC"),
	).All(ctx, th.BackendAppDb())
	require.NoError(th.T, err)
	return txns
}
