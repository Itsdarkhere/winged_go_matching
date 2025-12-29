package economy_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"wingedapp/pgtester/internal/wingedapp/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/economy"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCaseReferralBonus struct {
	name string

	invitee      *pgmodel.User
	referrer     *pgmodel.User
	inviteCode   *pgmodel.UserInviteCode
	inserter     *economy.InsertActionLog
	callCount    int // how many times to call CreateActionLog
	initialWings int // starting wings for referrer

	mutations       func(th *testsuite.Helper, tc *testCaseReferralBonus)
	assertions      func(th *testsuite.Helper, tc *testCaseReferralBonus, err error)
	extraAssertions func(th *testsuite.Helper, tc *testCaseReferralBonus, err error)
}

func referralBonusTestCases() []testCaseReferralBonus {
	return []testCaseReferralBonus{
		{
			// Per MVP spec: invitee gets NOTHING, only referrer gets wings
			name:         "success-no-referrer-invitee-gets-nothing",
			callCount:    1,
			initialWings: 0,
			mutations: func(th *testsuite.Helper, tc *testCaseReferralBonus) {
				// Create invitee
				tc.invitee = th.PersistRegisteredUser()

				// Create invite code WITHOUT referrer (system-generated)
				tc.inviteCode = (&factory.UserInviteCode{Subject: &pgmodel.UserInviteCode{}}).
					New(th.T, th.BackendAppDb()).
					SetRequiredFields().
					Save().
					Subject

				tc.inserter = &economy.InsertActionLog{
					UserID: tc.invitee.ID,
					RefID:  tc.inviteCode.ID,
					Type:   economy.ActionReferralComplete,
				}
			},
			assertions: func(th *testsuite.Helper, tc *testCaseReferralBonus, err error) {
				require.NoError(th.T, err, "referral bonus should succeed even without referrer")

				beStore := repo.Store{}

				// Verify invitee has action log (for idempotency tracking)
				actionLogs, err := beStore.WingsEcnActionLogs(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnActionLog{
					UserRefID: null.StringFrom(tc.invitee.ID),
				})
				require.NoError(th.T, err, "fetch action logs")
				require.Len(th.T, actionLogs, 1, "invitee should have 1 action log for tracking")

				// Verify invitee has ZERO wings (per MVP spec)
				userTotals, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.invitee.ID),
				})
				require.NoError(th.T, err, "fetch user totals")
				// User totals might not exist or have 0 wings
				if userTotals != nil {
					assert.Equal(th.T, 0, userTotals.TotalWings, "invitee should have 0 wings per MVP spec")
				}

				// Verify no transaction for invitee
				transactions, err := beStore.WingsEcnTransactions(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnTransaction{
					UserID: null.StringFrom(tc.invitee.ID),
				})
				require.NoError(th.T, err, "fetch transactions")
				assert.Len(th.T, transactions, 0, "invitee should have no transactions per MVP spec")
			},
		},
		{
			// Per MVP spec: referrer gets 4 wings, invitee gets nothing
			name:         "success-referrer-gets-4-wings-invitee-gets-nothing",
			callCount:    1,
			initialWings: 0,
			mutations: func(th *testsuite.Helper, tc *testCaseReferralBonus) {
				// Create invitee
				tc.invitee = th.PersistRegisteredUser()

				// Create referrer with sha256_hash column set
				referrerMobile := "+14155551234"
				hash := sha256.Sum256([]byte(referrerMobile))
				referrerMobileHash := hex.EncodeToString(hash[:])

				tc.referrer = (&factory.User{Subject: &pgmodel.User{
					MobileNumber: null.StringFrom(referrerMobile),
					Sha256Hash:   null.StringFrom(referrerMobileHash),
				}}).New(th.T, th.BackendAppDb()).SetRequiredFields().Save().Subject

				// Create invite code WITH referrer hash
				tc.inviteCode = (&factory.UserInviteCode{Subject: &pgmodel.UserInviteCode{
					ReferrerNumberHash: null.StringFrom(referrerMobileHash),
				}}).New(th.T, th.BackendAppDb()).SetRequiredFields().Save().Subject

				tc.inserter = &economy.InsertActionLog{
					UserID: tc.invitee.ID,
					RefID:  tc.inviteCode.ID,
					Type:   economy.ActionReferralComplete,
				}
			},
			assertions: func(th *testsuite.Helper, tc *testCaseReferralBonus, err error) {
				require.NoError(th.T, err, "referral bonus should succeed")

				beStore := repo.Store{}

				// Verify referrer has 4 wings (per MVP spec)
				referrerTotals, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.referrer.ID),
				})
				require.NoError(th.T, err, "fetch referrer totals")
				require.NotNil(th.T, referrerTotals, "referrer should have totals")
				assert.Equal(th.T, economy.ReferralBonusWings, referrerTotals.TotalWings, "referrer should have exactly 4 wings")

				// Verify referrer has 1 action log
				referrerLogs, err := beStore.WingsEcnActionLogs(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnActionLog{
					UserRefID: null.StringFrom(tc.referrer.ID),
				})
				require.NoError(th.T, err, "fetch referrer action logs")
				assert.Len(th.T, referrerLogs, 1, "referrer should have 1 action log")

				// Verify invitee has NO wings (per MVP spec)
				inviteeTotals, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.invitee.ID),
				})
				require.NoError(th.T, err, "fetch invitee totals")
				if inviteeTotals != nil {
					assert.Equal(th.T, 0, inviteeTotals.TotalWings, "invitee should have 0 wings per MVP spec")
				}

				// Verify invitee has no transactions
				inviteeTransactions, err := beStore.WingsEcnTransactions(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnTransaction{
					UserID: null.StringFrom(tc.invitee.ID),
				})
				require.NoError(th.T, err, "fetch invitee transactions")
				assert.Len(th.T, inviteeTransactions, 0, "invitee should have no transactions per MVP spec")
			},
		},
		{
			// Idempotency: multiple calls only credit referrer once
			name:         "success-idempotency-referrer-only-credited-once",
			callCount:    3, // call 3 times
			initialWings: 0,
			mutations: func(th *testsuite.Helper, tc *testCaseReferralBonus) {
				// Create invitee
				tc.invitee = th.PersistRegisteredUser()

				// Create referrer with sha256_hash column set
				referrerMobile := "+14155559999"
				hash := sha256.Sum256([]byte(referrerMobile))
				referrerMobileHash := hex.EncodeToString(hash[:])

				tc.referrer = (&factory.User{Subject: &pgmodel.User{
					MobileNumber: null.StringFrom(referrerMobile),
					Sha256Hash:   null.StringFrom(referrerMobileHash),
				}}).New(th.T, th.BackendAppDb()).SetRequiredFields().Save().Subject

				// Create invite code WITH referrer hash
				tc.inviteCode = (&factory.UserInviteCode{Subject: &pgmodel.UserInviteCode{
					ReferrerNumberHash: null.StringFrom(referrerMobileHash),
				}}).New(th.T, th.BackendAppDb()).SetRequiredFields().Save().Subject

				tc.inserter = &economy.InsertActionLog{
					UserID: tc.invitee.ID,
					RefID:  tc.inviteCode.ID,
					Type:   economy.ActionReferralComplete,
				}
			},
			assertions: func(th *testsuite.Helper, tc *testCaseReferralBonus, err error) {
				require.NoError(th.T, err, "referral bonus should succeed")

				beStore := repo.Store{}

				// Verify referrer only has ONE action log despite 3 calls
				referrerLogs, err := beStore.WingsEcnActionLogs(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnActionLog{
					UserRefID: null.StringFrom(tc.referrer.ID),
				})
				require.NoError(th.T, err, "fetch referrer action logs")
				assert.Len(th.T, referrerLogs, 1, "referrer should have exactly 1 action log despite 3 calls")

				// Verify referrer only has 4 wings (not 12)
				referrerTotals, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.referrer.ID),
				})
				require.NoError(th.T, err, "fetch referrer totals")
				require.NotNil(th.T, referrerTotals, "referrer should have totals")
				assert.Equal(th.T, economy.ReferralBonusWings, referrerTotals.TotalWings, "referrer should have exactly 4 wings (not 3x)")

				// Verify only 1 transaction for referrer
				referrerTransactions, err := beStore.WingsEcnTransactions(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnTransaction{
					UserID: null.StringFrom(tc.referrer.ID),
				})
				require.NoError(th.T, err, "fetch referrer transactions")
				assert.Len(th.T, referrerTransactions, 1, "referrer should have exactly 1 transaction")
			},
		},
		{
			// Referrer with existing wings gets bonus added
			name:         "success-referrer-with-existing-wings-gets-bonus-added",
			callCount:    1,
			initialWings: 100,
			mutations: func(th *testsuite.Helper, tc *testCaseReferralBonus) {
				// Create invitee
				tc.invitee = th.PersistRegisteredUser()

				// Create referrer with sha256_hash column set
				referrerMobile := "+14155558888"
				hash := sha256.Sum256([]byte(referrerMobile))
				referrerMobileHash := hex.EncodeToString(hash[:])

				tc.referrer = (&factory.User{Subject: &pgmodel.User{
					MobileNumber: null.StringFrom(referrerMobile),
					Sha256Hash:   null.StringFrom(referrerMobileHash),
				}}).New(th.T, th.BackendAppDb()).SetRequiredFields().Save().Subject

				// Create invite code WITH referrer hash
				tc.inviteCode = (&factory.UserInviteCode{Subject: &pgmodel.UserInviteCode{
					ReferrerNumberHash: null.StringFrom(referrerMobileHash),
				}}).New(th.T, th.BackendAppDb()).SetRequiredFields().Save().Subject

				// Set initial wings for referrer
				newTotals := &pgmodel.WingsEcnUserTotal{
					UserRefID:  tc.referrer.ID,
					TotalWings: tc.initialWings,
				}
				_ = newTotals.Insert(context.Background(), th.BackendAppDb(), boil.Infer())

				tc.inserter = &economy.InsertActionLog{
					UserID: tc.invitee.ID,
					RefID:  tc.inviteCode.ID,
					Type:   economy.ActionReferralComplete,
				}
			},
			assertions: func(th *testsuite.Helper, tc *testCaseReferralBonus, err error) {
				require.NoError(th.T, err, "referral bonus should succeed")

				beStore := repo.Store{}
				expectedWings := tc.initialWings + economy.ReferralBonusWings // 100 + 4 = 104

				// Verify referrer has initial + bonus
				referrerTotals, err := beStore.WingsEcnUserTotal(context.Background(), th.BackendAppDb(), &repo.QueryFilterWingsEcnUserTotal{
					UserID: null.StringFrom(tc.referrer.ID),
				})
				require.NoError(th.T, err, "fetch referrer totals")
				require.NotNil(th.T, referrerTotals, "referrer should have totals")
				assert.Equal(th.T, expectedWings, referrerTotals.TotalWings, "referrer should have initial + bonus wings")
			},
		},
	}
}

func TestEconomy_ReferralBonus(t *testing.T) {
	for _, tt := range referralBonusTestCases() {
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
