package economy_test

import (
	"context"
	"testing"
	"time"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/economy"
	store2 "wingedapp/pgtester/internal/wingedapp/lib/economy/store"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCaseWingedSubscriptionPayment struct {
	testName         string             // ex: 'test-top-up'
	actionType       economy.ActionType // ex: economy.ActionWingedXWeeklyPayment
	current          int                // ex: 0
	expectedAddition int                // ex: 1000 (the one the subscription gives)
	expectedTotals   int                // ex: 1000
}

// TestEconomy_WingedPayments is a test of all valid payments.
// The variables provided are used to extrapolate the action_logger inputs.
func TestEconomy_WingedPayments(t *testing.T) {

	tests := []testCaseWingedSubscriptionPayment{
		{
			testName:         "winged-plus-weekly",
			actionType:       economy.ActionWingedPlusWeeklyPayment,
			current:          0,
			expectedAddition: 25,
			expectedTotals:   25,
		},
		{
			testName:         "winged-plus-monthly",
			actionType:       economy.ActionWingedPlusMonthlyPayment,
			current:          0,
			expectedAddition: 55,
			expectedTotals:   55,
		},
		{
			testName:         "winged-plus-monthly-current-65",
			actionType:       economy.ActionWingedPlusMonthlyPayment,
			current:          65,
			expectedAddition: 55,
			expectedTotals:   120,
		},
		{
			testName:         "winged-plus-three-monthly",
			actionType:       economy.ActionWingedPlusThreeMonthPayment,
			current:          0,
			expectedAddition: 180,
			expectedTotals:   180,
		},
		{
			testName:         "winged-plus-three-monthly-added-current",
			actionType:       economy.ActionWingedPlusThreeMonthPayment,
			current:          425,
			expectedAddition: 180,
			expectedTotals:   605,
		},
		{
			testName:         "winged-plus-six-monthly",
			actionType:       economy.ActionWingedPlusSixMonthPayment,
			current:          0,
			expectedAddition: 360,
			expectedTotals:   360,
		},
		{
			testName:         "winged-plus-six-monthly-added-current",
			actionType:       economy.ActionWingedPlusSixMonthPayment,
			current:          180,
			expectedAddition: 360,
			expectedTotals:   540,
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			// Validate required test case fields (zero values would indicate misconfigured test case).
			require.NotZero(t, tc.testName, "test name is required")
			require.NotZero(t, tc.actionType, "action type is required")
			require.NotZero(t, tc.expectedTotals, "expected user wings is required")

			tSuite := testsuite.New(t)
			tSuite.FakeAPI().App() // init fakes

			cleanup := tSuite.UseBackendDB()
			defer cleanup()

			user := tSuite.PersistRegisteredUser()

			beStore := repo.Store{}

			// update total wings to current wings
			userTotals, err := beStore.WingsEcnUserTotal(context.Background(), tSuite.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
				UserID: null.StringFrom(user.ID),
			})
			require.NoError(t, err, "fetch user wings ecn totals")

			err = beStore.UpdateWingsEcnUserTotals(context.Background(), tSuite.BackendAppDb(), &repo.UpdateWingsEcnUserTotals{
				ID:         userTotals.ID,
				TotalWings: null.IntFrom(tc.current),
			})
			require.NoError(tSuite.T, err, "update user wings ecn totals")

			// do action log entry, expect success, and assert expected total wings
			refID := uuid.NewString() // dummy whatever
			inserterActionLog := &economy.InsertActionLog{
				UserID: user.ID,
				RefID:  refID, // dummy whatever
				Type:   tc.actionType,
			}

			db := tSuite.BackendAppDb()
			e := tSuite.FakeContainer().GetLibEconomy()
			err = e.CreateActionLog(context.Background(), db, inserterActionLog)
			require.NoError(t, err, "create action log")

			// insert action log
			actionLog, err := beStore.WingsEcnActionLog(context.Background(), tSuite.BackendAppDb(), &repo.QueryFilterWingsEcnActionLog{
				UserRefID: null.StringFrom(user.ID),
			})
			require.NoError(t, err, "fetch action log")
			require.NotNil(t, actionLog, "action logs should not be nil")

			t.Run("success-action-log-inserted", func(t *testing.T) {
				require.Equal(t, user.ID, actionLog.UserRefID, "action log ref ID should match")
				require.Equal(t, refID, actionLog.ExtDomainRefID, "action log ext domain ref ID should match")
				require.Equal(t, actionLog.IsCredit, actionLog.IsCredit, "action log is credit should match")

				updatedUserTotal, err := beStore.WingsEcnUserTotal(context.Background(), tSuite.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(user.ID),
				})
				require.NoError(t, err, "fetch updated user wings ecn totals")
				require.NotNil(t, updatedUserTotal, "updated user wings ecn totals should not be nil")

				// check user wings - expected totals
				require.Equal(t, tc.expectedTotals, updatedUserTotal.TotalWings, "user wings total should match expected")

				userTransactions, err := beStore.WingsEcnTransactions(context.Background(), tSuite.BackendAppDb(), &repo.QueryFilterWingsEcnTransaction{
					UserID: null.StringFrom(user.ID),
				})
				require.NoError(t, err, "fetch user wings ecn transactions")
				require.NotEmpty(t, userTransactions, "user wings ecn transactions should not be empty")

				// check transactions - expected addition
				require.Equal(t, tc.expectedAddition, userTransactions[0].Amount, "transaction wings amount should match expected")
			})

			t.Run("success-deleting-reverts-current-balance-and-sets-inactives", func(t *testing.T) {
				err = e.DeleteActionLog(context.Background(), tSuite.BackendAppDb(), actionLog.ID)
				require.NoError(tSuite.T, err, "delete action log")

				// ensure action log is inactive
				actionLog, err = beStore.WingsEcnActionLog(context.Background(), tSuite.BackendAppDb(), &repo.QueryFilterWingsEcnActionLog{
					ID: null.StringFrom(actionLog.ID),
				})
				require.NoError(tSuite.T, err, "fetch action log for deletion")
				require.NotNil(tSuite.T, actionLog, "action logs should not be nil for deletion")
				require.Zero(tSuite.T, actionLog.IsActive.Int, "action log should be inactive after deletion")

				// ensure transaction is inactive
				transaction, err := beStore.WingsEcnTransaction(context.Background(), tSuite.BackendAppDb(), &repo.QueryFilterWingsEcnTransaction{
					ActionLogID: null.StringFrom(actionLog.ID),
				})
				require.NoError(tSuite.T, err, "fetch transactions after action log deletion")
				require.NotNil(tSuite.T, transaction, "transactions should not be nil after action log deletion")
				require.Zero(tSuite.T, transaction.IsActive.Int, "transaction should be inactive after action log deletion")

				// ensure user totals is reverted to current
				econStor := store2.NewEconomyStores(applog.NewLogrus("test"))
				userTotals, err := econStor.UserTotalsStore.Totals(context.Background(), tSuite.BackendAppDb(), user.ID)
				require.NoError(tSuite.T, err, "fetch user totals after action log deletion")
				require.NotNil(tSuite.T, userTotals, "user totals should not be nil after action log deletion")
				require.Equal(tSuite.T, tc.current, userTotals.Wings, "user totals wings should be reverted after action log deletion")
			})
		})
	}
}

// TestEconomy_WingedPayments_SetsExpiry verifies that subscription payments set expires_at.
func TestEconomy_WingedPayments_SetsExpiry(t *testing.T) {
	t.Parallel()

	tSuite := testsuite.New(t)
	tSuite.FakeAPI().App()

	cleanup := tSuite.UseBackendDB()
	defer cleanup()

	user := tSuite.PersistRegisteredUser()

	beStore := repo.Store{}

	// Create action log for Winged+ Weekly payment
	refID := uuid.NewString()
	inserterActionLog := &economy.InsertActionLog{
		UserID: user.ID,
		RefID:  refID,
		Type:   economy.ActionWingedPlusWeeklyPayment,
	}

	db := tSuite.BackendAppDb()
	e := tSuite.FakeContainer().GetLibEconomy()
	err := e.CreateActionLog(context.Background(), db, inserterActionLog)
	require.NoError(t, err, "create action log")

	// Fetch the transaction and verify expires_at is set
	transaction, err := beStore.WingsEcnTransaction(context.Background(), db, &repo.QueryFilterWingsEcnTransaction{
		UserID: null.StringFrom(user.ID),
	})
	require.NoError(t, err, "fetch transaction")
	require.NotNil(t, transaction, "transaction should not be nil")

	// Verify expires_at is set and is approximately 30 days from now
	assert.True(t, transaction.ExpiresAt.Valid, "expires_at should be set")

	expectedExpiry := time.Now().AddDate(0, 0, economy.EarnedWingsExpiryDays)
	diff := transaction.ExpiresAt.Time.Sub(expectedExpiry)
	assert.True(t, diff < time.Minute && diff > -time.Minute,
		"expires_at should be ~30 days from now, got diff: %v", diff)
}
