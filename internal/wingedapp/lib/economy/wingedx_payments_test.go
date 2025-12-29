package economy_test

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/economy"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCaseActionLoggerWingedX struct {
	name                string
	inserterActionLog   *economy.InsertActionLog
	mutations           func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX)
	assertions          func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX, err error)
	userPremiumDuration null.Time
	userID              string
	cleanupFn           []func()
}

func (t *testCaseActionLoggerWingedX) addCleanup(cleanupFn func()) {
	if cleanupFn == nil {
		t.cleanupFn = make([]func(), 0)
	}
	t.cleanupFn = append(t.cleanupFn, cleanupFn)
}

func (t *testCaseActionLoggerWingedX) executeCleanups() {
	for _, cleanup := range t.cleanupFn {
		cleanup()
	}
}

func wingedXWeeklyPaymentTestsCases(t *testing.T) []testCaseActionLoggerWingedX {
	return []testCaseActionLoggerWingedX{
		{
			name:              "fail-missing-params",
			inserterActionLog: &economy.InsertActionLog{},
			mutations: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX) {
				// no profiles inserted

			},
			assertions: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX, err error) {
				require.Error(th.T, err, "expected error to be non-nil")

				// ref_id is conditionally required based on action type
				expects := []string{
					"'user_id' must have a value",
					"'category' must have a value",
				}
				for _, expect := range expects {
					assert.Contains(th.T, err.Error(), expect, fmt.Sprintf("expected error to contain '%s'", expect))
				}
			},
		},
		{
			name: "fail-user-non-existing",
			inserterActionLog: &economy.InsertActionLog{
				RefID: "",
			},
			mutations: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX) {

			},
			assertions: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX, err error) {
				require.Error(th.T, err, "expected error to be non-nil")

				// ref_id is conditionally required based on action type
				expects := []string{
					"'user_id' must have a value",
					"'category' must have a value",
				}
				for _, expect := range expects {
					assert.Contains(th.T, err.Error(), expect, fmt.Sprintf("expected error to contain '%s'", expect))
				}
			},
		},
		{
			name:              "success-insert-weekly-payment-day-0",
			inserterActionLog: &economy.InsertActionLog{},
			mutations: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX) {
				user := th.PersistRegisteredUser()

				tc.inserterActionLog = &economy.InsertActionLog{
					UserID: user.ID,
					RefID:  uuid.NewString(),
					Type:   economy.ActionWingedXWeeklyPayment,
				}
				tc.userID = user.ID
			},
			assertions: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX, err error) {
				require.NoError(t, err, "expected no error")

				store := repo.Store{}
				actionLog, err := store.WingsEcnActionLog(context.Background(), exec, &repo.QueryFilterWingsEcnActionLog{
					UserRefID: null.StringFrom(tc.userID),
				})
				require.NoError(t, err, "expected no error")
				require.NotNil(t, actionLog, "expected to find winged action log")

				// equal action type
				require.Equal(th.T, economy.CategoryWingedxWeeklyPayment, actionLog.ActionLogType)

				// equal user id
				require.Equal(th.T, actionLog.UserRefID, tc.userID)

				// equal ext ID
				require.Equal(th.T, actionLog.ExtDomainRefID, tc.inserterActionLog.RefID)

				// assert we are added like 7 days lol
				userTotals, err := store.WingsEcnUserTotal(context.Background(), exec, &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.userID),
				})
				require.NoError(t, err, "expected no error")
				require.NotNil(t, userTotals, "expected to find winged user totals")

				premiumExpiry := userTotals.PremiumExpiresIn.Time
				require.False(t, premiumExpiry.IsZero(), "expected premium expiry to be zero")

				// check the gap of it VS now
				gapInDays := int(premiumExpiry.Sub(time.Now()).Hours() / 24)

				// adjusted time to 1 day for test flakiness
				assert.GreaterOrEqual(th.T, gapInDays, 6, "expected gap in days to be at least 6")
			},
		},
		{
			name:              "success-insert-weekly-payment-currently-7-days-expect-14",
			inserterActionLog: &economy.InsertActionLog{},
			mutations: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX) {
				user := th.PersistRegisteredUser()

				tc.inserterActionLog = &economy.InsertActionLog{
					UserID: user.ID,
					RefID:  uuid.NewString(),
					Type:   economy.ActionWingedXWeeklyPayment,
				}
				tc.userID = user.ID

				store := repo.Store{}
				userTotal, err := store.WingsEcnUserTotal(context.Background(), exec, &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.userID),
				})
				require.NoError(t, err, "expected no error when finding user")
				require.NotNil(t, userTotal, "expected to find winged user totals")

				require.True(th.T, userTotal.PremiumExpiresIn.Time.IsZero(), "expected premium expiry to be zero")

				err = store.UpdateWingsEcnUserTotals(context.Background(), exec, &repo.UpdateWingsEcnUserTotals{
					ID:               userTotal.ID,
					PremiumExpiresIn: null.TimeFrom(time.Now().Add(7 * 24 * time.Hour)), // 7 days from now
				})
				require.NoError(t, err, "expected no error when finding user")
			},
			assertions: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX, err error) {
				require.NoError(t, err, "expected no error")

				store := repo.Store{}
				actionLog, err := store.WingsEcnActionLog(context.Background(), exec, &repo.QueryFilterWingsEcnActionLog{
					UserRefID: null.StringFrom(tc.userID),
				})
				require.NoError(t, err, "expected no error")
				require.NotNil(t, actionLog, "expected to find winged action log")

				// equal action type
				require.Equal(th.T, economy.CategoryWingedxWeeklyPayment, actionLog.ActionLogType)

				// equal user id
				require.Equal(th.T, actionLog.UserRefID, tc.userID)

				// equal ext ID
				require.Equal(th.T, actionLog.ExtDomainRefID, tc.inserterActionLog.RefID)

				// assert we are added like 7 days lol
				userTotals, err := store.WingsEcnUserTotal(context.Background(), exec, &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.userID),
				})
				require.NoError(t, err, "expected no error")
				require.NotNil(t, userTotals, "expected to find winged user totals")

				premiumExpiry := userTotals.PremiumExpiresIn.Time
				require.False(t, premiumExpiry.IsZero(), "expected premium expiry to be zero")

				// check the gap of it VS now
				gapInDays := int(premiumExpiry.Sub(time.Now()).Hours() / 24)

				// adjusted time to 1 day for test flakiness
				assert.GreaterOrEqual(th.T, gapInDays, 13, "expected gap in days to be at least 6")
			},
		},
	}
}

func TestEconomy_AddWingedXWeeklyPayment(t *testing.T) {
	for _, tc := range wingedXWeeklyPaymentTestsCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tSuite := testsuite.New(t)
			tSuite.FakeAPI().App() // init fakes

			cleanup := tSuite.UseBackendDB()
			defer cleanup()

			db := tSuite.BackendAppDb()
			tc.mutations(tSuite, db, &tc)
			e := tSuite.FakeContainer().GetLibEconomy()
			err := e.CreateActionLog(context.Background(), db, tc.inserterActionLog)
			tc.assertions(tSuite, db, &tc, err)
		})
	}
}

func wingedXMonthlyPaymentTestsCases(t *testing.T) []testCaseActionLoggerWingedX {
	return []testCaseActionLoggerWingedX{
		{
			name:              "fail-missing-params",
			inserterActionLog: &economy.InsertActionLog{},
			mutations: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX) {
				// no profiles inserted
			},
			assertions: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX, err error) {
				require.Error(th.T, err, "expected error to be non-nil")

				expects := []string{
					"'user_id' must have a value",
					"'ref_id' must have a value",
					"'category' must have a value",
				}
				for _, expect := range expects {
					assert.Contains(th.T, err.Error(), expect, fmt.Sprintf("expected error to contain '%s'", expect))
				}
			},
		},
		{
			name: "fail-user-non-existing",
			inserterActionLog: &economy.InsertActionLog{
				RefID: "",
			},
			mutations: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX) {

			},
			assertions: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX, err error) {
				require.Error(th.T, err, "expected error to be non-nil")

				expects := []string{
					"'user_id' must have a value",
					"'ref_id' must have a value",
					"'category' must have a value",
				}
				for _, expect := range expects {
					assert.Contains(th.T, err.Error(), expect, fmt.Sprintf("expected error to contain '%s'", expect))
				}
			},
		},
		{
			name:              "success-insert-monthly-payment-day-0",
			inserterActionLog: &economy.InsertActionLog{},
			mutations: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX) {
				user := th.PersistRegisteredUser()

				tc.inserterActionLog = &economy.InsertActionLog{
					UserID: user.ID,
					RefID:  uuid.NewString(),
					Type:   economy.ActionWingedXMonthlyPayment,
				}
				tc.userID = user.ID
			},
			assertions: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX, err error) {
				require.NoError(t, err, "expected no error")

				store := repo.Store{}
				actionLog, err := store.WingsEcnActionLog(context.Background(), exec, &repo.QueryFilterWingsEcnActionLog{
					UserRefID: null.StringFrom(tc.userID),
				})
				require.NoError(t, err, "expected no error")
				require.NotNil(t, actionLog, "expected to find winged action log")

				// equal action type
				require.Equal(th.T, economy.CategoryWingedxMonthlyPayment, actionLog.ActionLogType)

				// equal user id
				require.Equal(th.T, actionLog.UserRefID, tc.userID)

				// equal ext ID
				require.Equal(th.T, actionLog.ExtDomainRefID, tc.inserterActionLog.RefID)

				// assert we are added like 7 days lol
				userTotals, err := store.WingsEcnUserTotal(context.Background(), exec, &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.userID),
				})
				require.NoError(t, err, "expected no error")
				require.NotNil(t, userTotals, "expected to find winged user totals")

				premiumExpiry := userTotals.PremiumExpiresIn.Time
				require.False(t, premiumExpiry.IsZero(), "expected premium expiry to be zero")

				expected := time.Now().UTC().AddDate(0, 1, 0)
				actual := premiumExpiry.UTC()

				// pretty formats
				expectedNice := expected.Format(time.RFC3339)
				actualNice := actual.Format(time.RFC3339)
				t.Logf("Expected premium expiry around: %s", expectedNice)
				t.Logf("Actual premium expiry:          %s", actualNice)

				// Allow a small delta for test flakiness (e.g., a few seconds)
				delta := actual.Sub(expected)
				assert.LessOrEqual(t, math.Abs(delta.Seconds()), 5.0, "expiry should be about one month from now")
			},
		},
		{
			name:              "success-already-7-days-premium-expect-one-more-month-roughly-37-days",
			inserterActionLog: &economy.InsertActionLog{},
			mutations: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX) {
				user := th.PersistRegisteredUser()

				tc.inserterActionLog = &economy.InsertActionLog{
					UserID: user.ID,
					RefID:  uuid.NewString(),
					Type:   economy.ActionWingedXMonthlyPayment,
				}
				tc.userID = user.ID

				store := repo.Store{}
				userTotal, err := store.WingsEcnUserTotal(context.Background(), exec, &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.userID),
				})
				require.NoError(t, err, "expected no error when finding user")
				require.NotNil(t, userTotal, "expected to find winged user totals")

				require.True(th.T, userTotal.PremiumExpiresIn.Time.IsZero(), "expected premium expiry to be zero")

				tc.userPremiumDuration = null.TimeFrom(time.Now().Add(7 * 24 * time.Hour)) // 7 days from now
				err = store.UpdateWingsEcnUserTotals(context.Background(), exec, &repo.UpdateWingsEcnUserTotals{
					ID:               userTotal.ID,
					PremiumExpiresIn: tc.userPremiumDuration, // 7 days from now
				})
				require.NoError(t, err, "expected no error when finding user")
			},
			assertions: func(th *testsuite.Helper, exec boil.ContextExecutor, tc *testCaseActionLoggerWingedX, err error) {
				require.NoError(t, err, "expected no error")

				store := repo.Store{}
				actionLog, err := store.WingsEcnActionLog(context.Background(), exec, &repo.QueryFilterWingsEcnActionLog{
					UserRefID: null.StringFrom(tc.userID),
				})
				require.NoError(t, err, "expected no error")
				require.NotNil(t, actionLog, "expected to find winged action log")

				// equal action type
				require.Equal(th.T, economy.CategoryWingedxMonthlyPayment, actionLog.ActionLogType)

				// equal user id
				require.Equal(th.T, actionLog.UserRefID, tc.userID)

				// equal ext ID
				require.Equal(th.T, actionLog.ExtDomainRefID, tc.inserterActionLog.RefID)

				// assert we are added like 7 days lol
				userTotals, err := store.WingsEcnUserTotal(context.Background(), exec, &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.userID),
				})
				require.NoError(t, err, "expected no error")
				require.NotNil(t, userTotals, "expected to find winged user totals")

				premiumExpiry := userTotals.PremiumExpiresIn.Time
				require.False(t, premiumExpiry.IsZero(), "expected premium expiry to be zero")

				expected := tc.userPremiumDuration.Time.AddDate(0, 1, 0)
				actual := premiumExpiry

				expectedNice := expected.Format(time.RFC3339)
				actualNice := actual.Format(time.RFC3339)
				t.Logf("Expected premium expiry around: %s", expectedNice)
				t.Logf("Actual premium expiry:          %s", actualNice)

				// Allow a small delta for test flakiness (e.g., a few seconds)
				delta := actual.Sub(expected)
				assert.LessOrEqual(t, math.Abs(delta.Seconds()), 5.0, "expiry should be about one month from now")
			},
		},
	}
}

// AddWingedXMonthlyPayment adds a monthly winged x payment to the accounting system.
// Deprecated: initial version of wings economy, no longer used —
// but serves as good reference code.
func TestEconomy_AddWingedXMonthlyPayment(t *testing.T) {
	t.Skip() // deprecated initial version of wings economy
	for _, tc := range wingedXMonthlyPaymentTestsCases(t) {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tSuite := testsuite.New(t)
			tSuite.FakeAPI().App() // init fakes

			cleanup := tSuite.UseBackendDB()
			defer cleanup()

			db := tSuite.BackendAppDb()
			tc.mutations(tSuite, db, &tc)
			e := tSuite.FakeContainer().GetLibEconomy()
			err := e.CreateActionLog(context.Background(), db, tc.inserterActionLog)
			tc.assertions(tSuite, db, &tc, err)
		})
	}
}
