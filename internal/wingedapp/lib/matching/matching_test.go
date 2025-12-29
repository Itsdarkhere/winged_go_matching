package matching_test

import (
	"context"
	"strings"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/util/strutil"
	wingedFactory "wingedapp/pgtester/internal/wingedapp/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/enums"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"
	"wingedapp/pgtester/internal/wingedapp/lib/matching/testhelper"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCaseProcessMatch struct {
	name           string
	users          any // do I seed with specific users with X parameters? hmmm
	qResults       *matching.QualifierResults
	populationFile string
	setup          func(th *testsuite.Helper, l *matching.Logic)
	assertions     func(th *testsuite.Helper, matchResult *matching.MatchResult)
}

// asserFailSome asserts all hard qualifiers fail with some specific errors.
func asserFailSome(th *testsuite.Helper, matchResult *matching.MatchResult) {
	require.NotNil(th.T, matchResult, "match result should not be nil")
	require.NotNil(th.T, matchResult.QualifierResults.Valid, "match qualifier results should not be nil")
	require.NotNil(th.T, matchResult.QualifierResults.JSON, "match qualifier results should not be nil")
	require.NotEmpty(th.T, string(matchResult.QualifierResults.JSON), "match qualifier results JSON should not be empty")

	actual := string(matchResult.QualifierResults.JSON)
	expected := []error{
		matching.ErrAgeGapMaleExceeds,
		matching.ErrDistanceExceeds,
		matching.ErrHeightGapExists,
	}

	for _, e := range expected {
		assert.Contains(th.T, actual, e.Error(), "expected error not found in qualifier results")
	}
}

// assertPassAllWithQualitative asserts all hard qualifiers fail.
func assertPassAllWithQualitative(th *testsuite.Helper, matchResult *matching.MatchResult) {
	require.NotNil(th.T, matchResult, "match result should not be nil")
	require.NotNil(th.T, matchResult.QualifierResults.Valid, "match qualifier results should not be nil")
	require.NotNil(th.T, matchResult.QualifierResults.JSON, "match qualifier results should not be nil")
	require.NotEmpty(th.T, string(matchResult.QualifierResults.JSON), "match qualifier results JSON should not be empty")

	actual := string(matchResult.QualifierResults.JSON)
	expected := []error{
		matching.ErrAgeGapMaleExceeds,
		matching.ErrAgeGapFemaleExceeds,
		matching.ErrAgeGapSameSexExceeds,
		matching.ErrDistanceExceeds,
		matching.ErrNotInDatePrefs,
		matching.ErrHeightGapExists,
	}

	for _, e := range expected {
		assert.NotContains(th.T, actual, e.Error(), "expected no error not found in qualifier results")
		if strings.Contains(actual, e.Error()) {
			th.T.Log("please check", strutil.GetAsJson(actual))
		}
	}

	// assert we got values in match result
	storMatching := th.FakeContainer().GetStoreMatching()

	q := &matching.QueryFilterMatchResult{} // get all, no filter
	paginatedMatchResults, err := storMatching.MatchResultStore.MatchResults(context.Background(), th.BackendAppDb(), q)
	require.NoError(th.T, err, "fetching match results for assertion")

	matchResults := paginatedMatchResults.Data
	require.Len(th.T, matchResults, 1, "expecting match results for assertion")
	assert.NotEmpty(th.T, matchResults[0].IsPossibleMatch, "expecting match results for assertion")
	assert.True(th.T, matchResults[0].QualifierResults.Valid, "expecting match results for assertion")
	assert.NotEmpty(th.T, string(matchResults[0].QualifierResults.JSON), "expecting match results for assertion")
	assert.True(th.T, matchResults[0].MatchedQualitatively.Bool, "expecting match results for assertion")
}

func setupPassAllWithQualitative(th *testsuite.Helper, l *matching.Logic) {
	// Profiles are now created automatically by the CSV ingestor with population_details
	// Just set up the mock for qualitative matching to always pass
	mockQM := testhelper.MockSuccessQualitativeMatch(th.T)
	l.SetQualitativeQuantifier(mockQM)
}

func TestProcessMatch(t *testing.T) {
	testCases := []testCaseProcessMatch{
		{
			name:           "fail-all-hard-qualifiers",
			populationFile: "./testdata/population_2_fail_all.csv",
			assertions:     asserFailSome,
		},
		{
			name:           "pass-all-hard-qualifiers-with-qualitative",
			populationFile: "./testdata/population_3_pass_all.csv",
			assertions:     assertPassAllWithQualitative,
			setup:          setupPassAllWithQualitative,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.NotEmpty(t, tc.name, "test case name must be provided")
			require.NotEmpty(t, tc.populationFile, "population file must be provided")
			require.NotNil(t, tc.assertions, "extraAssertions function must be provided")

			tSuite := testsuite.New(t)

			// seed dummy dbs
			t.Cleanup(tSuite.UseBackendDB())
			t.Cleanup(tSuite.UseAiDB())
			t.Cleanup(tSuite.UseSupabaseAuthDB())

			ctn := tSuite.FakeContainer()
			matchingLib := ctn.GetLibMatching()

			// ingest population data
			ingestor := testhelper.NewPopulationIngestor(t, tSuite, matchingLib)
			parseResult, populateResult, err := ingestor.IngestFromCSVFile(tc.populationFile)
			require.NoError(t, err, "ingesting population data")
			require.NotEmpty(t, parseResult.Rows, "expected rows to not be empty")
			require.Equal(t, parseResult.ValidRows, populateResult.BackendAppUsers, "populate should create all users")
			require.Equal(t, parseResult.ValidRows, populateResult.AIBackendProfiles, "populate should create all profiles")
			require.Empty(t, populateResult.Errors, "expected no populate errors")

			// ingest seed population data into matching tables
			matchSet, err := matchingLib.IngestAll(context.Background(), tSuite.BackendAppDb())
			require.NoError(t, err, "ingesting population data")
			require.NotNil(t, matchSet, "expecting match set")

			if tc.setup != nil {
				tc.setup(tSuite, matchingLib)
			}

			// pick 1 entry from the matchSet, and run the matching algorithm on it
			storMatching := ctn.GetStoreMatching()
			paginatedMatchResults, err := storMatching.MatchResultStore.MatchResults(
				context.Background(),
				tSuite.BackendAppDb(),
				&matching.QueryFilterMatchResult{
					MatchSetID: null.StringFrom(matchSet.ID.String()),
				},
			)
			require.NoError(t, err, "fetching match results")

			matchResults := paginatedMatchResults.Data
			require.NotEmpty(t, matchResults, "expecting match results")
			require.Len(t, matchResults, 1, "expecting one 1 result (ensure your population file has only 2 entries — given we only focus on testing the matching algorithm here)")

			// run matching on the first one
			aiExec := tSuite.AiBackendDb()
			matchResult, err := matchingLib.ProcessMatchResult(context.Background(), tSuite.BackendAppDb(), aiExec, &matchResults[0])
			require.NoError(t, err, "running matching algorithm")
			tc.assertions(tSuite, matchResult)

			if tc.setup != nil {
				tc.setup(tSuite, matchingLib)
			}
		})
	}
}

type testCaseRunMatchForUnmatchedUsers struct {
	name            string
	setup           func(th *testsuite.Helper)
	extraAssertions func(th *testsuite.Helper, err error)
}

func TestLogic_RunMatchForUnmatchedUsers(t *testing.T) {

	testCases := []testCaseRunMatchForUnmatchedUsers{
		{
			name: "no-active-users-returns-nil",
			setup: func(th *testsuite.Helper) {
				// no users seeded - empty DB from harness
			},
			extraAssertions: func(th *testsuite.Helper, err error) {
				require.NoError(th.T, err, "RunMatchForUnmatchedUsers should succeed with no users")
			},
		},
		{
			name: "single-active-user-no-match-set-created",
			setup: func(th *testsuite.Helper) {
				exec := th.BackendAppDb()
				factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
			},
			extraAssertions: func(th *testsuite.Helper, err error) {
				require.NoError(th.T, err, "RunMatchForUnmatchedUsers should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				count, err := pgmodel.MatchSets().Count(ctx, exec)
				require.NoError(th.T, err, "counting match sets")
				assert.Equal(th.T, int64(0), count, "no match set should be created for single user")
			},
		},
		{
			name: "two-active-users-without-pending-matches-creates-match-set",
			setup: func(th *testsuite.Helper) {
				exec := th.BackendAppDb()
				factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
			},
			extraAssertions: func(th *testsuite.Helper, err error) {
				require.NoError(th.T, err, "RunMatchForUnmatchedUsers should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				matchSets, err := pgmodel.MatchSets().All(ctx, exec)
				require.NoError(th.T, err, "fetching match sets")
				require.Len(th.T, matchSets, 1, "one match set should be created")
				assert.Equal(th.T, 2, matchSets[0].NumberOfParticipants, "match set should have 2 participants")

				matchResults, err := pgmodel.MatchResults().All(ctx, exec)
				require.NoError(th.T, err, "fetching match results")
				assert.Len(th.T, matchResults, 1, "one match result should be created for 2 users")
			},
		},
		{
			name: "users-with-pending-matches-excluded",
			setup: func(th *testsuite.Helper) {
				exec := th.BackendAppDb()

				userA := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				userB := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)

				factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved: true,
						IsDropped:  false,
					},
					FactoryUserA: userA,
					FactoryUserB: userB,
				}).New(th.T, exec)
			},
			extraAssertions: func(th *testsuite.Helper, err error) {
				require.NoError(th.T, err, "RunMatchForUnmatchedUsers should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				matchSets, err := pgmodel.MatchSets().All(ctx, exec)
				require.NoError(th.T, err, "fetching match sets")
				assert.Len(th.T, matchSets, 1, "no new match set should be created")
			},
		},
		{
			name: "users-with-dropped-matches-included",
			setup: func(th *testsuite.Helper) {
				exec := th.BackendAppDb()

				userA := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				userB := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)

				factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved: true,
						IsDropped:  true,
					},
					FactoryUserA: userA,
					FactoryUserB: userB,
				}).New(th.T, exec)
			},
			extraAssertions: func(th *testsuite.Helper, err error) {
				require.NoError(th.T, err, "RunMatchForUnmatchedUsers should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				matchSets, err := pgmodel.MatchSets().All(ctx, exec)
				require.NoError(th.T, err, "fetching match sets")
				assert.Len(th.T, matchSets, 2, "new match set should be created for users with dropped matches")
			},
		},
		{
			name: "mixed-users-only-unmatched-included",
			setup: func(th *testsuite.Helper) {
				exec := th.BackendAppDb()

				userA := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				userB := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)

				// two more users without pending matches
				factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)

				factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved: true,
						IsDropped:  false,
					},
					FactoryUserA: userA,
					FactoryUserB: userB,
				}).New(th.T, exec)
			},
			extraAssertions: func(th *testsuite.Helper, err error) {
				require.NoError(th.T, err, "RunMatchForUnmatchedUsers should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				matchSets, err := pgmodel.MatchSets().All(ctx, exec)
				require.NoError(th.T, err, "fetching match sets")
				require.Len(th.T, matchSets, 2, "should have 2 match sets")

				var newestSet *pgmodel.MatchSet
				for i := range matchSets {
					if newestSet == nil || matchSets[i].CreatedAt.Time.After(newestSet.CreatedAt.Time) {
						newestSet = matchSets[i]
					}
				}

				assert.Equal(th.T, 2, newestSet.NumberOfParticipants, "new match set should only have 2 unmatched users")

				newResults, err := pgmodel.MatchResults(
					pgmodel.MatchResultWhere.MatchSetRefID.EQ(newestSet.ID),
				).All(ctx, exec)
				require.NoError(th.T, err, "fetching new match results")
				assert.Len(th.T, newResults, 1, "should have 1 match result for 2 unmatched users")
			},
		},
		{
			name: "three-unmatched-users-creates-three-pairings",
			setup: func(th *testsuite.Helper) {
				exec := th.BackendAppDb()
				factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
			},
			extraAssertions: func(th *testsuite.Helper, err error) {
				require.NoError(th.T, err, "RunMatchForUnmatchedUsers should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				matchSets, err := pgmodel.MatchSets().All(ctx, exec)
				require.NoError(th.T, err, "fetching match sets")
				require.Len(th.T, matchSets, 1, "one match set should be created")
				assert.Equal(th.T, 3, matchSets[0].NumberOfParticipants, "should have 3 participants")

				matchResults, err := pgmodel.MatchResults().All(ctx, exec)
				require.NoError(th.T, err, "fetching match results")
				assert.Len(th.T, matchResults, 3, "should have 3 match results for 3 users")
			},
		},
		{
			name: "unapproved-matches-dont-block-users",
			setup: func(th *testsuite.Helper) {
				exec := th.BackendAppDb()

				userA := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)
				userB := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(th.T, exec)

				factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved: false,
						IsDropped:  false,
					},
					FactoryUserA: userA,
					FactoryUserB: userB,
				}).New(th.T, exec)
			},
			extraAssertions: func(th *testsuite.Helper, err error) {
				require.NoError(th.T, err, "RunMatchForUnmatchedUsers should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				matchSets, err := pgmodel.MatchSets().All(ctx, exec)
				require.NoError(th.T, err, "fetching match sets")
				assert.Len(th.T, matchSets, 2, "new match set should be created (unapproved doesn't count as pending)")
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

			tc.setup(testSuite)

			matchLib := testSuite.FakeContainer().GetLibMatching()
			_, err := matchLib.RunMatchForUnmatchedUsers(context.Background(), testSuite.BackendAppDb())

			tc.extraAssertions(testSuite, err)
		})
	}
}

type testCaseProposeMatch struct {
	name            string
	setup           func(th *testsuite.Helper) (userAID, userBID, matchResultID string)
	extraAssertions func(th *testsuite.Helper, result *matching.ProposeMatchResult, err error, matchResultID string)
}

func proposeMatchTestCases() []testCaseProposeMatch {
	return []testCaseProposeMatch{
		{
			name: "success-first-proposal-no-date-instance",
			setup: func(th *testsuite.Helper) (string, string, string) {
				exec := th.BackendAppDb()

				// Create users with names (required for UserMatch query)
				userA := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{
					Subject: &pgmodel.User{
						FirstName: null.StringFrom("UserA"),
					},
				}).New(th.T, exec)
				userB := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{
					Subject: &pgmodel.User{
						FirstName: null.StringFrom("UserB"),
					},
				}).New(th.T, exec)

				// Create dropped match with:
				// - UserA = Pending (hasn't proposed yet)
				// - UserB = Pending (also hasn't proposed)
				// So when userA proposes, it's NOT mutual (partner B hasn't proposed yet)
				matchResult := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved:  true,
						IsDropped:   true,
						UserAAction: string(enums.MatchUserActionPending),
						UserBAction: string(enums.MatchUserActionPending),
					},
					FactoryUserA: userA,
					FactoryUserB: userB,
				}).New(th.T, exec)

				return userA.Subject.ID, userB.Subject.ID, matchResult.Subject.ID
			},
			extraAssertions: func(th *testsuite.Helper, result *matching.ProposeMatchResult, err error, matchResultID string) {
				require.NoError(th.T, err, "ProposeMatch should succeed")
				require.NotNil(th.T, result, "result should not be nil")
				assert.True(th.T, result.Success, "result.Success should be true")
				assert.False(th.T, result.MutualProposal, "should not be mutual proposal when partner hasn't proposed yet")
				assert.Empty(th.T, result.DateInstanceID, "no date instance should be created on first propose")

				// Verify no date_instance was created
				ctx := context.Background()
				count, err := pgmodel.DateInstances().Count(ctx, th.BackendAppDb())
				require.NoError(th.T, err, "counting date instances")
				assert.Equal(th.T, int64(0), count, "no date instance should exist")
			},
		},
		{
			name: "success-mutual-proposal-creates-date-instance",
			setup: func(th *testsuite.Helper) (string, string, string) {
				exec := th.BackendAppDb()

				// Create users with names (required for UserMatch query)
				userA := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{
					Subject: &pgmodel.User{
						FirstName: null.StringFrom("UserA"),
					},
				}).New(th.T, exec)
				userB := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{
					Subject: &pgmodel.User{
						FirstName: null.StringFrom("UserB"),
					},
				}).New(th.T, exec)

				// Create dropped match with:
				// - UserA = Pending (hasn't proposed yet - will propose now)
				// - UserB = Proposed (already proposed)
				// So when userA proposes, it IS mutual (partner B already proposed)
				matchResult := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
					Subject: &pgmodel.MatchResult{
						IsApproved:  true,
						IsDropped:   true,
						UserAAction: string(enums.MatchUserActionPending),
						UserBAction: string(enums.MatchUserActionProposed),
					},
					FactoryUserA: userA,
					FactoryUserB: userB,
				}).New(th.T, exec)

				// match_config is seeded by migration with match_expiration_hours = 72

				return userA.Subject.ID, userB.Subject.ID, matchResult.Subject.ID
			},
			extraAssertions: func(th *testsuite.Helper, result *matching.ProposeMatchResult, err error, matchResultID string) {
				require.NoError(th.T, err, "ProposeMatch should succeed")
				require.NotNil(th.T, result, "result should not be nil")
				assert.True(th.T, result.Success, "result.Success should be true")
				assert.True(th.T, result.MutualProposal, "should be mutual proposal when partner already proposed")
				assert.NotEmpty(th.T, result.DateInstanceID, "date instance should be created on mutual proposal")

				ctx := context.Background()
				exec := th.BackendAppDb()

				// Verify date_instance was created
				dateInstance, err := pgmodel.FindDateInstance(ctx, exec, result.DateInstanceID)
				require.NoError(th.T, err, "finding date instance")
				require.NotNil(th.T, dateInstance, "date instance should exist")
				assert.Equal(th.T, matchResultID, dateInstance.MatchResultRefID, "date instance should reference match")
				assert.NotEmpty(th.T, dateInstance.Status, "date instance should have status")
				assert.False(th.T, dateInstance.DecisionWindowEnd.IsZero(), "decision window end should be set")

				// Verify date_instance_log was created
				logs, err := pgmodel.DateInstanceLogs(
					pgmodel.DateInstanceLogWhere.DateInstanceRefID.EQ(result.DateInstanceID),
				).All(ctx, exec)
				require.NoError(th.T, err, "fetching date instance logs")
				require.Len(th.T, logs, 1, "should have 1 log entry")
				assert.Equal(th.T, "created", logs[0].EventType, "log should be 'created' event")

				// Verify match_result was updated
				matchResult, err := pgmodel.FindMatchResult(ctx, exec, matchResultID)
				require.NoError(th.T, err, "finding match result")
				assert.True(th.T, matchResult.CurrentDateInstanceID.Valid, "match should have current_date_instance_id")
				assert.Equal(th.T, result.DateInstanceID, matchResult.CurrentDateInstanceID.String, "match current_date_instance_id should match")
				assert.True(th.T, matchResult.MatchLifecycleStatus.Valid, "match should have lifecycle status")
			},
		},
	}
}

func TestLogic_ProposeMatch(t *testing.T) {

	for _, tc := range proposeMatchTestCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotEmpty(t, tc.name, "test case name required")
			require.NotNil(t, tc.setup, "setup function required")
			require.NotNil(t, tc.extraAssertions, "extraAssertions function required")

			testSuite := testsuite.New(t)
			t.Cleanup(testSuite.UseBackendDB())

			userAID, _, matchResultID := tc.setup(testSuite)

			matchLib := testSuite.FakeContainer().GetLibMatching()
			result, err := matchLib.ProposeMatch(
				context.Background(),
				testSuite.BackendAppDb(),
				&matching.ProposeMatchParams{
					MatchResultID: uuid.MustParse(matchResultID),
					UserID:        uuid.MustParse(userAID),
				},
			)

			tc.extraAssertions(testSuite, result, err, matchResultID)
		})
	}
}
