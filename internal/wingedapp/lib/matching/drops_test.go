package matching_test

import (
	"context"
	"testing"
	"time"
	"wingedapp/pgtester/internal/db/factory"
	wingedFactory "wingedapp/pgtester/internal/wingedapp/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/enums"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCaseDropMatches struct {
	name            string
	setup           func(th *testsuite.Helper) []*wingedFactory.MatchResult
	extraAssertions func(th *testsuite.Helper, factories []*wingedFactory.MatchResult, err error)
}

func TestLogic_DropMatches(t *testing.T) {

	testCases := []testCaseDropMatches{
		{
			name: "no-match-results-to-drop",
			setup: func(th *testsuite.Helper) []*wingedFactory.MatchResult {
				return nil
			},
			extraAssertions: func(th *testsuite.Helper, _ []*wingedFactory.MatchResult, err error) {
				require.NoError(th.T, err, "DropOneMatchPerUser should succeed with no results")
			},
		},
		{
			name: "single-approved-not-dropped-match-result-gets-dropped",
			setup: func(th *testsuite.Helper) []*wingedFactory.MatchResult {
				exec := th.BackendAppDb()
				mr := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved: true,
						IsDropped:  false,
					},
				}).New(th.T, exec)
				return []*wingedFactory.MatchResult{mr}
			},
			extraAssertions: func(th *testsuite.Helper, factories []*wingedFactory.MatchResult, err error) {
				require.NoError(th.T, err, "DropOneMatchPerUser should succeed")
				require.Len(th.T, factories, 1, "expecting 1 factory")

				ctx := context.Background()
				exec := th.BackendAppDb()

				mr, err := pgmodel.FindMatchResult(ctx, exec, factories[0].Subject.ID)
				require.NoError(th.T, err, "finding match result")
				assert.True(th.T, mr.IsDropped, "match result should be dropped")
				assert.True(th.T, mr.DroppedTS.Valid, "dropped_ts should be set")
			},
		},
		{
			name: "already-dropped-match-result-not-dropped-again",
			setup: func(th *testsuite.Helper) []*wingedFactory.MatchResult {
				exec := th.BackendAppDb()
				droppedTime := time.Now().Add(-24 * time.Hour)
				mr := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved: true,
						IsDropped:  true,
						DroppedTS:  null.TimeFrom(droppedTime),
					},
				}).New(th.T, exec)
				return []*wingedFactory.MatchResult{mr}
			},
			extraAssertions: func(th *testsuite.Helper, factories []*wingedFactory.MatchResult, err error) {
				require.NoError(th.T, err, "DropOneMatchPerUser should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				mr, err := pgmodel.FindMatchResult(ctx, exec, factories[0].Subject.ID)
				require.NoError(th.T, err, "finding match result")
				assert.True(th.T, mr.IsDropped, "match result should still be dropped")
				assert.True(th.T, mr.DroppedTS.Valid, "dropped_ts should be valid")
			},
		},
		{
			name: "not-approved-match-result-not-dropped",
			setup: func(th *testsuite.Helper) []*wingedFactory.MatchResult {
				exec := th.BackendAppDb()
				mr := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved: false,
						IsDropped:  false,
					},
				}).New(th.T, exec)
				return []*wingedFactory.MatchResult{mr}
			},
			extraAssertions: func(th *testsuite.Helper, factories []*wingedFactory.MatchResult, err error) {
				require.NoError(th.T, err, "DropOneMatchPerUser should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				mr, err := pgmodel.FindMatchResult(ctx, exec, factories[0].Subject.ID)
				require.NoError(th.T, err, "finding match result")
				assert.False(th.T, mr.IsDropped, "match result should NOT be dropped since not approved")
			},
		},
		{
			name: "multiple-match-results-user-only-gets-one-drop",
			setup: func(th *testsuite.Helper) []*wingedFactory.MatchResult {
				exec := th.BackendAppDb()
				ctx := context.Background()

				userA := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				userB := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				userC := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				matchSet := factory.NewEntity[*wingedFactory.MatchSet](&wingedFactory.MatchSet{}).New(th.T, exec)

				mr1 := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved:           true,
						IsDropped:            false,
						MatchLifecycleStatus: null.StringFrom(string(enums.MatchLifecycleStatusScheduling)),
					},
					FactoryMatchSet: matchSet,
					FactoryUserA:    userA,
					FactoryUserB:    userB,
				}).New(th.T, exec)

				// Manually set created_at to ensure ordering
				_, err := exec.ExecContext(ctx, "UPDATE match_result SET created_at = created_at - interval '1 minute' WHERE id = $1", mr1.Subject.ID)
				require.NoError(th.T, err, "updating created_at for mr1")

				mr2 := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved:           true,
						IsDropped:            false,
						MatchLifecycleStatus: null.StringFrom(string(enums.MatchLifecycleStatusScheduling)),
					},
					FactoryMatchSet: matchSet,
					FactoryUserA:    userA,
					FactoryUserB:    userC,
				}).New(th.T, exec)

				return []*wingedFactory.MatchResult{mr1, mr2}
			},
			extraAssertions: func(th *testsuite.Helper, factories []*wingedFactory.MatchResult, err error) {
				require.NoError(th.T, err, "DropOneMatchPerUser should succeed")
				require.Len(th.T, factories, 2, "expecting 2 factories")

				ctx := context.Background()
				exec := th.BackendAppDb()

				// Verify UserA is the same in both match results (test setup validation)
				assert.Equal(th.T, factories[0].Subject.UserARefID, factories[1].Subject.UserARefID, "UserA should be the same in both match results")

				mr1, err := pgmodel.FindMatchResult(ctx, exec, factories[0].Subject.ID)
				require.NoError(th.T, err, "finding first match result")
				assert.True(th.T, mr1.IsDropped, "first match result should be dropped")
				assert.True(th.T, mr1.DroppedTS.Valid, "first match result dropped_ts should be set")

				mr2, err := pgmodel.FindMatchResult(ctx, exec, factories[1].Subject.ID)
				require.NoError(th.T, err, "finding second match result")
				assert.False(th.T, mr2.IsDropped, "second match result should NOT be dropped (UserA already dropped)")
			},
		},
		{
			name: "multiple-independent-users-all-get-drops",
			setup: func(th *testsuite.Helper) []*wingedFactory.MatchResult {
				exec := th.BackendAppDb()

				userA := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				userB := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				userC := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				userD := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				matchSet := factory.NewEntity[*wingedFactory.MatchSet](&wingedFactory.MatchSet{}).New(th.T, exec)

				mr1 := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved:           true,
						IsDropped:            false,
						MatchLifecycleStatus: null.StringFrom(string(enums.MatchLifecycleStatusScheduling)),
					},
					FactoryMatchSet: matchSet,
					FactoryUserA:    userA,
					FactoryUserB:    userB,
				}).New(th.T, exec)

				mr2 := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved:           true,
						IsDropped:            false,
						MatchLifecycleStatus: null.StringFrom(string(enums.MatchLifecycleStatusScheduling)),
					},
					FactoryMatchSet: matchSet,
					FactoryUserA:    userC,
					FactoryUserB:    userD,
				}).New(th.T, exec)

				return []*wingedFactory.MatchResult{mr1, mr2}
			},
			extraAssertions: func(th *testsuite.Helper, factories []*wingedFactory.MatchResult, err error) {
				require.NoError(th.T, err, "DropOneMatchPerUser should succeed")
				require.Len(th.T, factories, 2, "expecting 2 factories")

				ctx := context.Background()
				exec := th.BackendAppDb()

				mr1, err := pgmodel.FindMatchResult(ctx, exec, factories[0].Subject.ID)
				require.NoError(th.T, err, "finding first match result")
				assert.True(th.T, mr1.IsDropped, "first match result should be dropped")

				mr2, err := pgmodel.FindMatchResult(ctx, exec, factories[1].Subject.ID)
				require.NoError(th.T, err, "finding second match result")
				assert.True(th.T, mr2.IsDropped, "second match result should be dropped")
			},
		},
		{
			name: "fifo-order-oldest-dropped-first",
			setup: func(th *testsuite.Helper) []*wingedFactory.MatchResult {
				exec := th.BackendAppDb()
				ctx := context.Background()

				sharedUser := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				userB := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				userC := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				matchSet := factory.NewEntity[*wingedFactory.MatchSet](&wingedFactory.MatchSet{}).New(th.T, exec)

				mr1 := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved:           true,
						IsDropped:            false,
						MatchLifecycleStatus: null.StringFrom(string(enums.MatchLifecycleStatusScheduling)),
					},
					FactoryMatchSet: matchSet,
					FactoryUserA:    sharedUser,
					FactoryUserB:    userB,
				}).New(th.T, exec)

				// Manually set created_at to ensure ordering (mr1 is older)
				_, err := exec.ExecContext(ctx, "UPDATE match_result SET created_at = created_at - interval '1 minute' WHERE id = $1", mr1.Subject.ID)
				require.NoError(th.T, err, "updating created_at for mr1")

				mr2 := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved:           true,
						IsDropped:            false,
						MatchLifecycleStatus: null.StringFrom(string(enums.MatchLifecycleStatusScheduling)),
					},
					FactoryMatchSet: matchSet,
					FactoryUserA:    sharedUser,
					FactoryUserB:    userC,
				}).New(th.T, exec)

				return []*wingedFactory.MatchResult{mr1, mr2}
			},
			extraAssertions: func(th *testsuite.Helper, factories []*wingedFactory.MatchResult, err error) {
				require.NoError(th.T, err, "DropOneMatchPerUser should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				mr1, err := pgmodel.FindMatchResult(ctx, exec, factories[0].Subject.ID)
				require.NoError(th.T, err, "finding first match result")
				assert.True(th.T, mr1.IsDropped, "older match result should be dropped (FIFO)")

				mr2, err := pgmodel.FindMatchResult(ctx, exec, factories[1].Subject.ID)
				require.NoError(th.T, err, "finding second match result")
				assert.False(th.T, mr2.IsDropped, "newer match result should NOT be dropped")
			},
		},
		{
			name: "user-b-side-also-tracked",
			setup: func(th *testsuite.Helper) []*wingedFactory.MatchResult {
				exec := th.BackendAppDb()
				ctx := context.Background()

				userA := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				userB := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				userC := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				matchSet := factory.NewEntity[*wingedFactory.MatchSet](&wingedFactory.MatchSet{}).New(th.T, exec)

				// UserB is on B side
				mr1 := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved:           true,
						IsDropped:            false,
						MatchLifecycleStatus: null.StringFrom(string(enums.MatchLifecycleStatusScheduling)),
					},
					FactoryMatchSet: matchSet,
					FactoryUserA:    userA,
					FactoryUserB:    userB,
				}).New(th.T, exec)

				// Manually set created_at to ensure ordering (mr1 is older)
				_, err := exec.ExecContext(ctx, "UPDATE match_result SET created_at = created_at - interval '1 minute' WHERE id = $1", mr1.Subject.ID)
				require.NoError(th.T, err, "updating created_at for mr1")

				// UserB is on B side again
				mr2 := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved:           true,
						IsDropped:            false,
						MatchLifecycleStatus: null.StringFrom(string(enums.MatchLifecycleStatusScheduling)),
					},
					FactoryMatchSet: matchSet,
					FactoryUserA:    userC,
					FactoryUserB:    userB,
				}).New(th.T, exec)

				return []*wingedFactory.MatchResult{mr1, mr2}
			},
			extraAssertions: func(th *testsuite.Helper, factories []*wingedFactory.MatchResult, err error) {
				require.NoError(th.T, err, "DropOneMatchPerUser should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				mr1, err := pgmodel.FindMatchResult(ctx, exec, factories[0].Subject.ID)
				require.NoError(th.T, err, "finding first match result")
				assert.True(th.T, mr1.IsDropped, "first match result should be dropped")

				mr2, err := pgmodel.FindMatchResult(ctx, exec, factories[1].Subject.ID)
				require.NoError(th.T, err, "finding second match result")
				assert.False(th.T, mr2.IsDropped, "second match result should NOT be dropped (UserB already got drop)")
			},
		},
		{
			name: "idempotent-second-call-does-not-change-dropped-ts",
			setup: func(th *testsuite.Helper) []*wingedFactory.MatchResult {
				exec := th.BackendAppDb()
				mr := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved: true,
						IsDropped:  false,
					},
				}).New(th.T, exec)
				return []*wingedFactory.MatchResult{mr}
			},
			extraAssertions: func(th *testsuite.Helper, factories []*wingedFactory.MatchResult, err error) {
				require.NoError(th.T, err, "first DropOneMatchPerUser should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				mr, err := pgmodel.FindMatchResult(ctx, exec, factories[0].Subject.ID)
				require.NoError(th.T, err, "finding match result after first drop")
				assert.True(th.T, mr.IsDropped, "should be dropped after first call")
				firstDroppedTS := mr.DroppedTS

				// Second call
				matchLib := th.FakeContainer().GetLibMatching()
				err = matchLib.DropOneMatchPerUser(ctx, exec)
				require.NoError(th.T, err, "second DropOneMatchPerUser should succeed")

				mr, err = pgmodel.FindMatchResult(ctx, exec, factories[0].Subject.ID)
				require.NoError(th.T, err, "finding match result after second drop")
				assert.True(th.T, mr.IsDropped, "should still be dropped")
				assert.Equal(th.T, firstDroppedTS.Time, mr.DroppedTS.Time, "dropped_ts should not change on second call")
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotEmpty(t, tc.name, "test case name required")
			require.NotNil(t, tc.setup, "setup function required")
			require.NotNil(t, tc.extraAssertions, "extraAssertions function required")

			testSuite := testsuite.New(t)
			t.Cleanup(testSuite.UseBackendDB())

			factories := tc.setup(testSuite)

			ctx := context.Background()
			exec := testSuite.BackendAppDb()
			matchLib := testSuite.FakeContainer().GetLibMatching()
			err := matchLib.DropOneMatchPerUser(ctx, exec)

			tc.extraAssertions(testSuite, factories, err)
		})
	}
}
