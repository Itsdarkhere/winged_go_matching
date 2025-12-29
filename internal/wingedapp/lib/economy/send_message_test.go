package economy_test

import (
	"context"
	"testing"
	"time"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/economy"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type testCaseSendMessage struct {
	name string

	user         *pgmodel.User
	messageID    string
	inserter     *economy.InsertActionLog
	callCount    int // how many times to call CreateActionLog
	initialWings int // starting wings for user

	mutations       func(th *testsuite.Helper, tc *testCaseSendMessage)
	assertions      func(th *testsuite.Helper, tc *testCaseSendMessage, err error)
	extraAssertions func(th *testsuite.Helper, tc *testCaseSendMessage, err error)
}

func sendMessageTestCases() []testCaseSendMessage {
	return []testCaseSendMessage{
		{
			name:         "success-message-counted-no-deduction-yet",
			callCount:    1,
			initialWings: 10,
			mutations: func(th *testsuite.Helper, tc *testCaseSendMessage) {
				tc.user = th.PersistRegisteredUser()
				tc.messageID = uuid.New().String()

				// Set initial wings
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
					RefID:  tc.messageID,
					Type:   economy.ActionSendMessage,
				}
			},
			assertions: func(th *testsuite.Helper, tc *testCaseSendMessage, err error) {
				require.NoError(th.T, err, "send message should succeed")

				beStore := repo.Store{}

				// Verify action log created
				actionLogs, err := beStore.WingsEcnActionLogs(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnActionLog{
					UserRefID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch action logs")
				require.Len(th.T, actionLogs, 1, "user should have exactly 1 action log")
				require.Equal(th.T, string(economy.ActionSendMessage), actionLogs[0].ActionLogType)

				// Verify sent_messages incremented
				userTotals, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch user totals")
				require.Equal(th.T, 1, userTotals.CounterSentMessages, "sent_messages should be 1")
				require.Equal(th.T, tc.initialWings, userTotals.TotalWings, "wings should NOT be deducted (not at threshold)")

				// Verify NO transaction yet (not at threshold)
				transactions, err := beStore.WingsEcnTransactions(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnTransaction{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch transactions")
				require.Len(th.T, transactions, 0, "no transaction until threshold reached")
			},
		},
		{
			name:         "success-5th-message-deducts-1-wing",
			callCount:    1,
			initialWings: 10,
			mutations: func(th *testsuite.Helper, tc *testCaseSendMessage) {
				tc.user = th.PersistRegisteredUser()

				// Set initial wings and sent_messages to 4 (next one triggers deduction)
				beStore := repo.Store{}
				userTotals, _ := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				_ = beStore.UpdateWingsEcnUserTotals(context.Background(), th.BackendAppDb(), &repo.UpdateWingsEcnUserTotals{
					ID:           userTotals.ID,
					TotalWings:   null.IntFrom(tc.initialWings),
					SentMessages: null.IntFrom(4), // 4 already sent
				})

				tc.messageID = uuid.New().String()
				tc.inserter = &economy.InsertActionLog{
					UserID: tc.user.ID,
					RefID:  tc.messageID,
					Type:   economy.ActionSendMessage,
				}
			},
			assertions: func(th *testsuite.Helper, tc *testCaseSendMessage, err error) {
				require.NoError(th.T, err, "send message should succeed")

				beStore := repo.Store{}

				// Verify sent_messages = 5
				userTotals, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch user totals")
				require.Equal(th.T, 5, userTotals.CounterSentMessages, "sent_messages should be 5")
				require.Equal(th.T, tc.initialWings-economy.SendMessageWingsCost, userTotals.TotalWings, "1 wing should be deducted")

				// Verify debit transaction created
				transactions, err := beStore.WingsEcnTransactions(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnTransaction{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch transactions")
				require.Len(th.T, transactions, 1, "should have 1 debit transaction")
				require.Equal(th.T, economy.SendMessageWingsCost, transactions[0].Amount)
				require.False(th.T, transactions[0].IsCredit, "should be debit (not credit)")
			},
		},
		{
			name:         "success-idempotency-same-message-not-counted-twice",
			callCount:    3, // call 3 times with same message ID
			initialWings: 10,
			mutations: func(th *testsuite.Helper, tc *testCaseSendMessage) {
				tc.user = th.PersistRegisteredUser()
				tc.messageID = uuid.New().String()

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
					RefID:  tc.messageID,
					Type:   economy.ActionSendMessage,
				}
			},
			assertions: func(th *testsuite.Helper, tc *testCaseSendMessage, err error) {
				require.NoError(th.T, err, "send message should succeed")

				beStore := repo.Store{}

				// Verify only 1 action log despite 3 calls
				actionLogs, err := beStore.WingsEcnActionLogs(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnActionLog{
					UserRefID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch action logs")
				require.Len(th.T, actionLogs, 1, "only 1 action log despite 3 calls (idempotency)")

				// Verify sent_messages only incremented once
				userTotals, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch user totals")
				require.Equal(th.T, 1, userTotals.CounterSentMessages, "sent_messages should be 1 (idempotency)")
			},
		},
		{
			name:         "error-insufficient-wings",
			callCount:    1,
			initialWings: 0,
			mutations: func(th *testsuite.Helper, tc *testCaseSendMessage) {
				tc.user = th.PersistRegisteredUser()
				tc.messageID = uuid.New().String()

				// User has 0 wings
				beStore := repo.Store{}
				userTotals, _ := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				_ = beStore.UpdateWingsEcnUserTotals(context.Background(), th.BackendAppDb(), &repo.UpdateWingsEcnUserTotals{
					ID:         userTotals.ID,
					TotalWings: null.IntFrom(0),
				})

				tc.inserter = &economy.InsertActionLog{
					UserID: tc.user.ID,
					RefID:  tc.messageID,
					Type:   economy.ActionSendMessage,
				}
			},
			assertions: func(th *testsuite.Helper, tc *testCaseSendMessage, err error) {
				require.Error(th.T, err, "should error with insufficient wings")
				require.ErrorIs(th.T, err, economy.ErrInsufficientWings)

				beStore := repo.Store{}

				// Verify NO action log created
				actionLogs, err := beStore.WingsEcnActionLogs(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnActionLog{
					UserRefID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch action logs")
				require.Len(th.T, actionLogs, 0, "no action log when insufficient wings")
			},
		},
		{
			name:         "success-10th-message-deducts-2nd-wing",
			callCount:    1,
			initialWings: 10,
			mutations: func(th *testsuite.Helper, tc *testCaseSendMessage) {
				tc.user = th.PersistRegisteredUser()

				// Set sent_messages to 9 (10th triggers 2nd deduction)
				beStore := repo.Store{}
				userTotals, _ := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				_ = beStore.UpdateWingsEcnUserTotals(context.Background(), th.BackendAppDb(), &repo.UpdateWingsEcnUserTotals{
					ID:           userTotals.ID,
					TotalWings:   null.IntFrom(tc.initialWings),
					SentMessages: null.IntFrom(9),
				})

				tc.messageID = uuid.New().String()
				tc.inserter = &economy.InsertActionLog{
					UserID: tc.user.ID,
					RefID:  tc.messageID,
					Type:   economy.ActionSendMessage,
				}
			},
			assertions: func(th *testsuite.Helper, tc *testCaseSendMessage, err error) {
				require.NoError(th.T, err, "send message should succeed")

				beStore := repo.Store{}

				userTotals, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				require.NoError(th.T, err, "fetch user totals")
				require.Equal(th.T, 10, userTotals.CounterSentMessages, "sent_messages should be 10")
				require.Equal(th.T, tc.initialWings-economy.SendMessageWingsCost, userTotals.TotalWings, "wing deducted at 10th message")
			},
		},
	}
}

func TestEconomy_SendMessage(t *testing.T) {
	for _, tt := range sendMessageTestCases() {
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

type testCaseCanPerformAction struct {
	name string

	user         *pgmodel.User
	initialWings int
	actionType   economy.ActionType

	mutations  func(th *testsuite.Helper, tc *testCaseCanPerformAction)
	assertions func(th *testsuite.Helper, tc *testCaseCanPerformAction, canPerform bool, err error)
}

func canPerformActionTestCases() []testCaseCanPerformAction {
	return []testCaseCanPerformAction{
		{
			name:         "success-can-send-message-with-wings",
			initialWings: 5,
			actionType:   economy.ActionSendMessage,
			mutations: func(th *testsuite.Helper, tc *testCaseCanPerformAction) {
				tc.user = th.PersistRegisteredUser()

				beStore := repo.Store{}
				userTotals, _ := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				_ = beStore.UpdateWingsEcnUserTotals(context.Background(), th.BackendAppDb(), &repo.UpdateWingsEcnUserTotals{
					ID:         userTotals.ID,
					TotalWings: null.IntFrom(tc.initialWings),
				})
			},
			assertions: func(th *testsuite.Helper, tc *testCaseCanPerformAction, canPerform bool, err error) {
				require.NoError(th.T, err)
				require.True(th.T, canPerform, "should be able to send message with wings")
			},
		},
		{
			name:         "error-cannot-send-message-without-wings",
			initialWings: 0,
			actionType:   economy.ActionSendMessage,
			mutations: func(th *testsuite.Helper, tc *testCaseCanPerformAction) {
				tc.user = th.PersistRegisteredUser()

				beStore := repo.Store{}
				userTotals, _ := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				_ = beStore.UpdateWingsEcnUserTotals(context.Background(), th.BackendAppDb(), &repo.UpdateWingsEcnUserTotals{
					ID:         userTotals.ID,
					TotalWings: null.IntFrom(0),
				})
			},
			assertions: func(th *testsuite.Helper, tc *testCaseCanPerformAction, canPerform bool, err error) {
				require.False(th.T, canPerform, "should NOT be able to send message without wings")
				require.ErrorIs(th.T, err, economy.ErrInsufficientWings)
			},
		},
		{
			name:         "success-earning-actions-always-allowed",
			initialWings: 0,
			actionType:   economy.ActionAttendDate,
			mutations: func(th *testsuite.Helper, tc *testCaseCanPerformAction) {
				tc.user = th.PersistRegisteredUser()

				// Even with 0 wings, earning actions should be allowed
				beStore := repo.Store{}
				userTotals, _ := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				_ = beStore.UpdateWingsEcnUserTotals(context.Background(), th.BackendAppDb(), &repo.UpdateWingsEcnUserTotals{
					ID:         userTotals.ID,
					TotalWings: null.IntFrom(0),
				})
			},
			assertions: func(th *testsuite.Helper, tc *testCaseCanPerformAction, canPerform bool, err error) {
				require.NoError(th.T, err)
				require.True(th.T, canPerform, "earning actions always allowed")
			},
		},
		{
			name:         "success-premium-subscriber-bypasses-wings-check",
			initialWings: 0, // zero wings but has premium
			actionType:   economy.ActionSendMessage,
			mutations: func(th *testsuite.Helper, tc *testCaseCanPerformAction) {
				tc.user = th.PersistRegisteredUser()

				// Set 0 wings but active premium subscription
				beStore := repo.Store{}
				userTotals, _ := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				futureExpiry := time.Now().Add(24 * time.Hour) // expires tomorrow
				_ = beStore.UpdateWingsEcnUserTotals(context.Background(), th.BackendAppDb(), &repo.UpdateWingsEcnUserTotals{
					ID:               userTotals.ID,
					TotalWings:       null.IntFrom(0),
					PremiumExpiresIn: null.TimeFrom(futureExpiry),
				})
			},
			assertions: func(th *testsuite.Helper, tc *testCaseCanPerformAction, canPerform bool, err error) {
				require.NoError(th.T, err)
				require.True(th.T, canPerform, "premium subscriber should bypass wings check")
			},
		},
		{
			name:         "error-expired-premium-does-not-bypass",
			initialWings: 0,
			actionType:   economy.ActionSendMessage,
			mutations: func(th *testsuite.Helper, tc *testCaseCanPerformAction) {
				tc.user = th.PersistRegisteredUser()

				// Set 0 wings with EXPIRED premium subscription
				beStore := repo.Store{}
				userTotals, _ := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.user.ID),
				})
				pastExpiry := time.Now().Add(-24 * time.Hour) // expired yesterday
				_ = beStore.UpdateWingsEcnUserTotals(context.Background(), th.BackendAppDb(), &repo.UpdateWingsEcnUserTotals{
					ID:               userTotals.ID,
					TotalWings:       null.IntFrom(0),
					PremiumExpiresIn: null.TimeFrom(pastExpiry),
				})
			},
			assertions: func(th *testsuite.Helper, tc *testCaseCanPerformAction, canPerform bool, err error) {
				require.False(th.T, canPerform, "expired premium should NOT bypass wings check")
				require.ErrorIs(th.T, err, economy.ErrInsufficientWings)
			},
		},
	}
}

func TestEconomy_CanPerformAction(t *testing.T) {
	for _, tt := range canPerformActionTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tSuite := testsuite.New(t)
			tSuite.FakeAPI().App()
			ctn := tSuite.FakeContainer()

			cleanup := tSuite.UseBackendDB()
			defer cleanup()

			if tt.mutations != nil {
				tt.mutations(tSuite, &tt)
			}

			e := ctn.GetLibEconomy()

			canPerform, err := e.CanPerformAction(context.Background(), tSuite.BackendAppDb(), &economy.CanPerformActionParams{
				UserID:     tt.user.ID,
				ActionType: tt.actionType,
			})

			tt.assertions(tSuite, &tt, canPerform, err)
		})
	}
}
