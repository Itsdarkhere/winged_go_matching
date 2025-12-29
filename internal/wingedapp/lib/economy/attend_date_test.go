package economy_test

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/economy"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type testCaseAttendDate struct {
	name string

	user           *pgmodel.User
	dateInstanceID string
	inserter       *economy.InsertActionLog
	callCount      int // how many times to call CreateActionLog
	initialWings   int // starting wings for user

	mutations       func(th *testsuite.Helper, tc *testCaseAttendDate)
	assertions      func(th *testsuite.Helper, tc *testCaseAttendDate, err error)
	extraAssertions func(th *testsuite.Helper, tc *testCaseAttendDate, err error)
}

func attendDateTestCases() []testCaseAttendDate {
	return []testCaseAttendDate{
		{
			name:         "success-user-credited-once",
			callCount:    1,
			initialWings: 0,
			mutations: func(th *testsuite.Helper, tc *testCaseAttendDate) {
				tc.user = th.PersistRegisteredUser()
				tc.dateInstanceID = uuid.New().String()

				tc.inserter = &economy.InsertActionLog{
					UserID: tc.user.ID,
					RefID:  tc.dateInstanceID,
					Type:   economy.ActionAttendDate,
				}
			},
			assertions: func(th *testsuite.Helper, tc *testCaseAttendDate, err error) {
				require.NoError(th.T, err, "attend date bonus should succeed")

				beStore := repo.Store{}

				// Verify user has action log
				actionLogs, err := beStore.WingsEcnActionLogs(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnActionLog{
					UserRefID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch action logs")
				require.Len(th.T, actionLogs, 1, "user should have exactly 1 action log")
				require.Equal(th.T, string(economy.ActionAttendDate), actionLogs[0].ActionLogType)

				// Verify user has bonus wings
				userTotals, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch user totals")
				require.Equal(th.T, economy.AttendDateWings, userTotals.TotalWings, "user should have attend date bonus wings")

				// Verify transaction
				transactions, err := beStore.WingsEcnTransactions(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnTransaction{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch transactions")
				require.Len(th.T, transactions, 1, "user should have exactly 1 transaction")
				require.Equal(th.T, economy.AttendDateWings, transactions[0].Amount)
				require.True(th.T, transactions[0].IsCredit)
			},
		},
		{
			name:         "success-idempotency-user-only-credited-once",
			callCount:    3, // call 3 times
			initialWings: 0,
			mutations: func(th *testsuite.Helper, tc *testCaseAttendDate) {
				tc.user = th.PersistRegisteredUser()
				tc.dateInstanceID = uuid.New().String()

				tc.inserter = &economy.InsertActionLog{
					UserID: tc.user.ID,
					RefID:  tc.dateInstanceID,
					Type:   economy.ActionAttendDate,
				}
			},
			assertions: func(th *testsuite.Helper, tc *testCaseAttendDate, err error) {
				require.NoError(th.T, err, "attend date bonus should succeed")

				beStore := repo.Store{}

				// Verify user only has ONE action log (idempotency)
				actionLogs, err := beStore.WingsEcnActionLogs(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnActionLog{
					UserRefID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch action logs")
				require.Len(th.T, actionLogs, 1, "user should have exactly 1 action log despite 3 calls")

				// Verify user only has bonus amount (not 3x)
				userTotals, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch user totals")
				require.Equal(th.T, economy.AttendDateWings, userTotals.TotalWings, "user should have exactly 1x bonus wings")

				// Verify only 1 transaction
				transactions, err := beStore.WingsEcnTransactions(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnTransaction{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch transactions")
				require.Len(th.T, transactions, 1, "user should have exactly 1 transaction")
			},
		},
		{
			name:         "success-user-credited-with-existing-wings",
			callCount:    1,
			initialWings: 50,
			mutations: func(th *testsuite.Helper, tc *testCaseAttendDate) {
				tc.user = th.PersistRegisteredUser()
				tc.dateInstanceID = uuid.New().String()

				// Set initial wings for user
				beStore := repo.Store{}
				userTotals, _ := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				_ = beStore.UpdateWingsEcnUserTotals(context.Background(), th.BackendAppDb(), &repo.UpdateWingsEcnUserTotals{
					ID:         userTotals.ID,
					TotalWings: null.IntFrom(tc.initialWings),
				})

				tc.inserter = &economy.InsertActionLog{
					UserID: tc.user.ID,
					RefID:  tc.dateInstanceID,
					Type:   economy.ActionAttendDate,
				}
			},
			assertions: func(th *testsuite.Helper, tc *testCaseAttendDate, err error) {
				require.NoError(th.T, err, "attend date bonus should succeed")

				beStore := repo.Store{}
				expectedWings := tc.initialWings + economy.AttendDateWings

				// Verify user has initial + bonus
				userTotals, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch user totals")
				require.Equal(th.T, expectedWings, userTotals.TotalWings, "user should have initial + bonus wings")
			},
		},
		{
			name:         "success-different-date-instances-credited-separately",
			callCount:    1,
			initialWings: 0,
			mutations: func(th *testsuite.Helper, tc *testCaseAttendDate) {
				tc.user = th.PersistRegisteredUser()
				tc.dateInstanceID = uuid.New().String()

				tc.inserter = &economy.InsertActionLog{
					UserID: tc.user.ID,
					RefID:  tc.dateInstanceID,
					Type:   economy.ActionAttendDate,
				}
			},
			assertions: func(th *testsuite.Helper, tc *testCaseAttendDate, err error) {
				require.NoError(th.T, err, "first attend date bonus should succeed")

				beStore := repo.Store{}

				// First date instance credited
				userTotals, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch user totals")
				require.Equal(th.T, economy.AttendDateWings, userTotals.TotalWings)
			},
			extraAssertions: func(th *testsuite.Helper, tc *testCaseAttendDate, err error) {
				// Credit for a DIFFERENT date instance should work
				e := th.FakeContainer().GetLibEconomy()
				secondDateInstanceID := uuid.New().String()

				err = e.CreateActionLog(context.Background(), th.BackendAppDb(), &economy.InsertActionLog{
					UserID: tc.user.ID,
					RefID:  secondDateInstanceID,
					Type:   economy.ActionAttendDate,
				})
				require.NoError(th.T, err, "second attend date bonus should succeed")

				beStore := repo.Store{}

				// User should now have 2x wings
				userTotals, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch user totals")
				require.Equal(th.T, economy.AttendDateWings*2, userTotals.TotalWings, "user should have wings from 2 different dates")

				// User should have 2 action logs
				actionLogs, err := beStore.WingsEcnActionLogs(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnActionLog{
					UserRefID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch action logs")
				require.Len(th.T, actionLogs, 2, "user should have 2 action logs for 2 different dates")
			},
		},
	}
}

func TestEconomy_AttendDate(t *testing.T) {
	for _, tt := range attendDateTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tSuite := testsuite.New(t)
			tSuite.FakeAPI().App() // init fakes
			ctn := tSuite.FakeContainer()

			cleanup := tSuite.UseBackendDB()
			defer cleanup()

			// setup
			if tt.mutations != nil {
				tt.mutations(tSuite, &tt)
			}

			e := ctn.GetLibEconomy()

			// Call CreateActionLog multiple times to test idempotency
			var lastErr error
			for i := 0; i < tt.callCount; i++ {
				lastErr = e.CreateActionLog(context.Background(), tSuite.BackendAppDb(), tt.inserter)
			}

			// assertions
			tt.assertions(tSuite, &tt, lastErr)
			if tt.extraAssertions != nil {
				tt.extraAssertions(tSuite, &tt, lastErr)
			}
		})
	}
}
