package matching_test

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	wingedFactory "wingedapp/pgtester/internal/wingedapp/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"
	"wingedapp/pgtester/internal/wingedapp/lib/matching/testhelper"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCaseRunIngestionSet struct {
	name       string
	setup      func(th *testsuite.Helper, matchLib *matching.Logic) *matching.MatchSet
	assertions func(th *testsuite.Helper, matchSetID string, err error)
}

func TestLogic_RunIngestionSet(t *testing.T) {

	testCases := []testCaseRunIngestionSet{
		{
			name: "success-processes-all-match-results",
			setup: func(th *testsuite.Helper, matchLib *matching.Logic) *matching.MatchSet {
				// Seed 2 users that will fail hard qualifiers (different locations)
				ingestor := testhelper.NewPopulationIngestor(th.T, th, matchLib)
				parseResult, populateResult, err := ingestor.IngestFromCSVFile("./testdata/population_2_fail_all.csv")
				require.NoError(th.T, err, "ingesting population data")
				require.Equal(th.T, 2, parseResult.ValidRows, "expected 2 valid rows")
				require.Equal(th.T, 2, populateResult.BackendAppUsers, "expected 2 backend users")
				require.Equal(th.T, 2, populateResult.AIBackendProfiles, "expected 2 profiles")

				// Create MatchSet without running matching
				matchSet, err := matchLib.Ingest(context.Background(), th.BackendAppDb(), &matching.QueryFilterUser{
					IsActive: null.BoolFrom(true),
				})
				require.NoError(th.T, err, "creating match set")
				return matchSet
			},
			assertions: func(th *testsuite.Helper, matchSetID string, err error) {
				require.NoError(th.T, err, "RunIngestionSet should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				// Verify all MatchResults have QualifierResults populated
				matchResults, err := pgmodel.MatchResults(
					pgmodel.MatchResultWhere.MatchSetRefID.EQ(matchSetID),
				).All(ctx, exec)
				require.NoError(th.T, err, "fetching match results")
				require.NotEmpty(th.T, matchResults, "should have match results")

				for _, mr := range matchResults {
					assert.True(th.T, mr.QualifierResults.Valid, "QualifierResults should be valid for result %s", mr.ID)
					assert.NotEmpty(th.T, string(mr.QualifierResults.JSON), "QualifierResults JSON should not be empty for result %s", mr.ID)
				}
			},
		},
		{
			name: "success-stores-qualifier-results-json",
			setup: func(th *testsuite.Helper, matchLib *matching.Logic) *matching.MatchSet {
				ingestor := testhelper.NewPopulationIngestor(th.T, th, matchLib)
				parseResult, populateResult, err := ingestor.IngestFromCSVFile("./testdata/population_2_fail_all.csv")
				require.NoError(th.T, err, "ingesting population data")
				require.Equal(th.T, 2, parseResult.ValidRows, "expected 2 valid rows")
				require.Equal(th.T, 2, populateResult.BackendAppUsers, "expected 2 backend users")

				matchSet, err := matchLib.Ingest(context.Background(), th.BackendAppDb(), &matching.QueryFilterUser{
					IsActive: null.BoolFrom(true),
				})
				require.NoError(th.T, err, "creating match set")
				return matchSet
			},
			assertions: func(th *testsuite.Helper, matchSetID string, err error) {
				require.NoError(th.T, err, "RunIngestionSet should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				matchResults, err := pgmodel.MatchResults(
					pgmodel.MatchResultWhere.MatchSetRefID.EQ(matchSetID),
				).All(ctx, exec)
				require.NoError(th.T, err, "fetching match results")
				require.Len(th.T, matchResults, 1, "should have 1 match result for 2 users")

				// Verify QualifierResults JSON contains qualifier names
				jsonStr := string(matchResults[0].QualifierResults.JSON)
				assert.Contains(th.T, jsonStr, "age_window_qualifier", "should contain age qualifier")
				assert.Contains(th.T, jsonStr, "distance_qualifier", "should contain distance qualifier")
			},
		},
		{
			name: "success-hard-qualifier-failures-stored-not-thrown",
			setup: func(th *testsuite.Helper, matchLib *matching.Logic) *matching.MatchSet {
				// Use population that fails all hard qualifiers
				ingestor := testhelper.NewPopulationIngestor(th.T, th, matchLib)
				parseResult, populateResult, err := ingestor.IngestFromCSVFile("./testdata/population_2_fail_all.csv")
				require.NoError(th.T, err, "ingesting population data")
				require.Equal(th.T, 2, parseResult.ValidRows, "expected 2 valid rows")
				require.Equal(th.T, 2, populateResult.BackendAppUsers, "expected 2 backend users")

				matchSet, err := matchLib.Ingest(context.Background(), th.BackendAppDb(), &matching.QueryFilterUser{
					IsActive: null.BoolFrom(true),
				})
				require.NoError(th.T, err, "creating match set")
				return matchSet
			},
			assertions: func(th *testsuite.Helper, matchSetID string, err error) {
				require.NoError(th.T, err, "RunIngestionSet should succeed even with qualifier failures")

				ctx := context.Background()
				exec := th.BackendAppDb()

				matchResults, err := pgmodel.MatchResults(
					pgmodel.MatchResultWhere.MatchSetRefID.EQ(matchSetID),
				).All(ctx, exec)
				require.NoError(th.T, err, "fetching match results")
				require.Len(th.T, matchResults, 1, "should have 1 match result")

				mr := matchResults[0]
				jsonStr := string(mr.QualifierResults.JSON)

				// Verify hard qualifier errors are stored in JSON
				expectedErrors := []string{
					matching.ErrAgeGapMaleExceeds.Error(),
					matching.ErrDistanceExceeds.Error(),
				}
				for _, expectedErr := range expectedErrors {
					assert.Contains(th.T, jsonStr, expectedErr, "should contain error: %s", expectedErr)
				}

				// MatchedQualitatively should be false since hard qualifiers failed
				assert.False(th.T, mr.MatchedQualitatively, "MatchedQualitatively should be false")
			},
		},
		{
			name: "success-qualitative-match-runs-when-hard-pass",
			setup: func(th *testsuite.Helper, matchLib *matching.Logic) *matching.MatchSet {
				// Use population that passes all hard qualifiers
				// CSV now includes population_details which creates profiles automatically
				ingestor := testhelper.NewPopulationIngestor(th.T, th, matchLib)
				parseResult, populateResult, err := ingestor.IngestFromCSVFile("./testdata/population_3_pass_all.csv")
				require.NoError(th.T, err, "ingesting population data")
				require.Equal(th.T, 2, parseResult.ValidRows, "expected 2 valid rows")
				require.Equal(th.T, 2, populateResult.BackendAppUsers, "expected 2 backend users")
				require.Equal(th.T, 2, populateResult.AIBackendProfiles, "expected 2 profiles")
				require.Empty(th.T, populateResult.Errors, "expected no populate errors")

				// Mock qualitative quantifier to always pass
				mockQM := testhelper.MockSuccessQualitativeMatch(th.T)
				matchLib.SetQualitativeQuantifier(mockQM)

				matchSet, err := matchLib.Ingest(context.Background(), th.BackendAppDb(), &matching.QueryFilterUser{
					IsActive: null.BoolFrom(true),
				})
				require.NoError(th.T, err, "creating match set")
				return matchSet
			},
			assertions: func(th *testsuite.Helper, matchSetID string, err error) {
				require.NoError(th.T, err, "RunIngestionSet should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				matchResults, err := pgmodel.MatchResults(
					pgmodel.MatchResultWhere.MatchSetRefID.EQ(matchSetID),
				).All(ctx, exec)
				require.NoError(th.T, err, "fetching match results")
				require.Len(th.T, matchResults, 1, "should have 1 match result")

				mr := matchResults[0]
				// When hard qualifiers pass and qualitative runs, MatchedQualitatively should be true
				assert.True(th.T, mr.MatchedQualitatively, "MatchedQualitatively should be true")
				assert.True(th.T, mr.IsPossibleMatch, "IsPossibleMatch should be true")
			},
		},
		{
			name: "success-empty-match-set-no-error",
			setup: func(th *testsuite.Helper, matchLib *matching.Logic) *matching.MatchSet {
				// Create a MatchSet with no users (manually insert empty one)
				exec := th.BackendAppDb()

				matchSet := factory.NewEntity[*wingedFactory.MatchSet](&wingedFactory.MatchSet{
					Subject: &pgmodel.MatchSet{
						Name:                 "empty-test",
						NumberOfParticipants: 0,
					},
				}).New(th.T, exec)

				parsedID, err := uuid.Parse(matchSet.Subject.ID)
				require.NoError(th.T, err, "parsing match set ID")

				return &matching.MatchSet{
					ID:                   parsedID,
					Name:                 matchSet.Subject.Name,
					NumberOfParticipants: matchSet.Subject.NumberOfParticipants,
				}
			},
			assertions: func(th *testsuite.Helper, matchSetID string, err error) {
				require.NoError(th.T, err, "RunIngestionSet should succeed with empty match set")
			},
		},
		{
			name: "success-updates-matched-qualitatively-flag-false-on-hard-fail",
			setup: func(th *testsuite.Helper, matchLib *matching.Logic) *matching.MatchSet {
				ingestor := testhelper.NewPopulationIngestor(th.T, th, matchLib)
				parseResult, populateResult, err := ingestor.IngestFromCSVFile("./testdata/population_2_fail_all.csv")
				require.NoError(th.T, err, "ingesting population data")
				require.Equal(th.T, 2, parseResult.ValidRows, "expected 2 valid rows")
				require.Equal(th.T, 2, populateResult.BackendAppUsers, "expected 2 backend users")
				require.Empty(th.T, populateResult.Errors, "expected no populate errors")

				matchSet, err := matchLib.Ingest(context.Background(), th.BackendAppDb(), &matching.QueryFilterUser{
					IsActive: null.BoolFrom(true),
				})
				require.NoError(th.T, err, "creating match set")
				return matchSet
			},
			assertions: func(th *testsuite.Helper, matchSetID string, err error) {
				require.NoError(th.T, err)

				ctx := context.Background()
				exec := th.BackendAppDb()

				matchResults, err := pgmodel.MatchResults(
					pgmodel.MatchResultWhere.MatchSetRefID.EQ(matchSetID),
				).All(ctx, exec)
				require.NoError(th.T, err)
				require.NotEmpty(th.T, matchResults)

				for _, mr := range matchResults {
					// MatchedQualitatively is a bool (not null.Bool) - false means hard qualifiers failed
					assert.False(th.T, mr.MatchedQualitatively, "MatchedQualitatively should be false for hard qualifier failures")
				}
			},
		},
		{
			name: "success-multiple-pairs-all-processed",
			setup: func(th *testsuite.Helper, matchLib *matching.Logic) *matching.MatchSet {
				// Use population with 6 users = 15 pairs
				ingestor := testhelper.NewPopulationIngestor(th.T, th, matchLib)
				parseResult, populateResult, err := ingestor.IngestFromCSVFile("./testdata/population_1.csv")
				require.NoError(th.T, err, "ingesting population data")
				require.Equal(th.T, 6, parseResult.ValidRows, "expected 6 valid rows")
				require.Equal(th.T, 6, populateResult.BackendAppUsers, "expected 6 backend users")
				require.Equal(th.T, 6, populateResult.AIBackendProfiles, "expected 6 profiles")
				require.Empty(th.T, populateResult.Errors, "expected no populate errors")

				// Mock qualitative quantifier to avoid hitting real AI endpoint
				mockQM := testhelper.MockSuccessQualitativeMatch(th.T)
				matchLib.SetQualitativeQuantifier(mockQM)

				matchSet, err := matchLib.Ingest(context.Background(), th.BackendAppDb(), &matching.QueryFilterUser{
					IsActive: null.BoolFrom(true),
				})
				require.NoError(th.T, err, "creating match set")
				return matchSet
			},
			assertions: func(th *testsuite.Helper, matchSetID string, err error) {
				require.NoError(th.T, err, "RunIngestionSet should succeed")

				ctx := context.Background()
				exec := th.BackendAppDb()

				matchResults, err := pgmodel.MatchResults(
					pgmodel.MatchResultWhere.MatchSetRefID.EQ(matchSetID),
				).All(ctx, exec)
				require.NoError(th.T, err, "fetching match results")

				// 6 users = 6*5/2 = 15 pairs
				require.Len(th.T, matchResults, 15, "should have 15 match results for 6 users")

				// All should have QualifierResults populated
				processedCount := 0
				for _, mr := range matchResults {
					if mr.QualifierResults.Valid && len(mr.QualifierResults.JSON) > 0 {
						processedCount++
					}
				}
				assert.Equal(th.T, 15, processedCount, "all 15 pairs should be processed")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotEmpty(t, tc.name, "test case name required")
			require.NotNil(t, tc.setup, "setup function required")
			require.NotNil(t, tc.assertions, "assertions function required")

			testSuite := testsuite.New(t)
			t.Cleanup(testSuite.UseBackendDB())
			t.Cleanup(testSuite.UseAiDB())
			t.Cleanup(testSuite.UseSupabaseAuthDB())

			matchLib := testSuite.FakeContainer().GetLibMatching()
			matchSet := tc.setup(testSuite, matchLib)

			aiExec := testSuite.AiBackendDb()
			runErr := matchLib.RunIngestionSet(context.Background(), testSuite.BackendAppDb(), aiExec, matchSet.ID)

			tc.assertions(testSuite, matchSet.ID.String(), runErr)
		})
	}
}
