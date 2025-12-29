package economy_test

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/economy"
	store2 "wingedapp/pgtester/internal/wingedapp/lib/economy/store"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type testCaseWingedPayment struct {
	current          int
	expectedAddition int
	expectedTotals   int
}

type testCaseCreateActionLog struct {
	name             string // ex: 'test-top-up'
	subscriptionType string // ex: 'Top Up'
	subscriptionName string // ex: 'Weekly'

	/* domain specific */
	testCaseWingedPayment testCaseWingedPayment

	inserter *economy.InsertActionLog
	user     *pgmodel.User

	/* test-driven funcs */
	mutations       func(th *testsuite.Helper, tc *testCaseCreateActionLog)
	assertions      func(th *testsuite.Helper, tc *testCaseCreateActionLog, err error)
	extraAssertions func(th *testsuite.Helper, tc *testCaseCreateActionLog, err error)
}

func addActionLogTestCases() []testCaseCreateActionLog {
	return []testCaseCreateActionLog{
		{
			name: "success-add-action-log-winged-plus-weekly",
			testCaseWingedPayment: testCaseWingedPayment{
				current:          0,
				expectedAddition: 25,
				expectedTotals:   25,
			},
			mutations: func(th *testsuite.Helper, tc *testCaseCreateActionLog) {
				tc.inserter = &economy.InsertActionLog{
					UserID: tc.user.ID,
					RefID:  uuid.NewString(),
					Type:   economy.ActionWingedPlusWeeklyPayment,
				}

				beStore := repo.Store{}
				// update total wings to current wings
				userTotals, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch user wings ecn totals")

				err = beStore.UpdateWingsEcnUserTotals(context.Background(), th.BackendAppDb(), &repo.UpdateWingsEcnUserTotals{
					ID:         userTotals.ID,
					TotalWings: null.IntFrom(tc.testCaseWingedPayment.current),
				})
				require.NoError(th.T, err, "update user wings ecn totals")
			},
			assertions: func(th *testsuite.Helper, tc *testCaseCreateActionLog, err error) {
				require.NoError(th.T, err, "add action log")

				beStore := repo.Store{}
				actionLog, err := beStore.WingsEcnActionLog(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnActionLog{
					UserRefID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch action log")
				require.NotNil(th.T, actionLog, "action logs should not be nil")

				require.Equal(th.T, tc.user.ID, actionLog.UserRefID, "action log ref ID should match")
				require.Equal(th.T, tc.inserter.RefID, actionLog.ExtDomainRefID, "action log ext domain ref ID should match")
				require.Equal(th.T, actionLog.IsCredit, actionLog.IsCredit, "action log is credit should match")

				updatedUserTotal, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch updated user wings ecn totals")
				require.NotNil(th.T, updatedUserTotal, "updated user wings ecn totals should not be nil")

				// check user wings - expected totals
				require.Equal(th.T, tc.testCaseWingedPayment.expectedTotals, updatedUserTotal.TotalWings, "user wings total should match expected")

				userTransactions, err := beStore.WingsEcnTransactions(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnTransaction{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch user wings ecn transactions")
				require.NotEmpty(th.T, userTransactions, "user wings ecn transactions should not be empty")

				// check transactions - expected addition
				amount := userTransactions[0].Amount
				require.Equal(th.T, tc.testCaseWingedPayment.expectedAddition, amount, "transaction wings amount should match expected")
			},
			extraAssertions: func(th *testsuite.Helper, tc *testCaseCreateActionLog, err error) {
				store := repo.Store{}
				actionLog, err := store.WingsEcnActionLog(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnActionLog{
					UserRefID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch action log for deletion")
				require.NotNil(th.T, actionLog, "action logs should not be nil for deletion")

				e := th.FakeContainer().GetLibEconomy()

				// ensure no error
				err = e.DeleteActionLog(context.Background(), th.BackendAppDb(), actionLog.ID)
				require.NoError(th.T, err, "delete action log")

				// ensure action log is inactive
				actionLog, err = store.WingsEcnActionLog(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnActionLog{
					ID: null.StringFrom(actionLog.ID),
				})
				require.NoError(th.T, err, "fetch action log for deletion")
				require.NotNil(th.T, actionLog, "action logs should not be nil for deletion")
				require.Zero(th.T, actionLog.IsActive.Int, "action log should be inactive after deletion")

				// ensure transaction is inactive
				transaction, err := store.WingsEcnTransaction(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnTransaction{
					ActionLogID: null.StringFrom(actionLog.ID),
				})
				require.NoError(th.T, err, "fetch transactions after action log deletion")
				require.NotNil(th.T, transaction, "transactions should not be nil after action log deletion")
				require.Zero(th.T, transaction.IsActive.Int, "transaction should be inactive after action log deletion")

				// ensure user totals is reverted to current
				econStor := store2.NewEconomyStores(applog.NewLogrus("test"))
				userTotals, err := econStor.UserTotalsStore.Totals(context.Background(), th.BackendAppDb(), tc.user.ID)
				require.NoError(th.T, err, "fetch user totals after action log deletion")
				require.NotNil(th.T, userTotals, "user totals should not be nil after action log deletion")
				require.Equal(th.T, tc.testCaseWingedPayment.current, userTotals.Wings, "user totals wings should be reverted after action log deletion")
			},
		},
	}
}

func TestEconomy_AddActionLog(t *testing.T) {
	for _, tt := range addActionLogTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tSuite := testsuite.New(t)
			tSuite.FakeAPI().App() // init fakes
			ctn := tSuite.FakeContainer()

			cleanup := tSuite.UseBackendDB()
			defer cleanup()
			tt.user = tSuite.PersistRegisteredUser()

			// setup
			if tt.mutations != nil {
				tt.mutations(tSuite, &tt)
			}

			e := ctn.GetLibEconomy()
			err := e.CreateActionLog(context.Background(), tSuite.BackendAppDb(), tt.inserter)

			// assertions
			tt.assertions(tSuite, &tt, err)
			if tt.extraAssertions != nil {
				tt.extraAssertions(tSuite, &tt, err)
			}
		})
	}
}
