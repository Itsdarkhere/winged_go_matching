package matching_test

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/business/sdk"
	wingedFactory "wingedapp/pgtester/internal/wingedapp/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"
	"wingedapp/pgtester/internal/wingedapp/lib/matching/testhelper"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/ericlagergren/decimal"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustDecimal creates a types.Decimal from a string, panics on error.
func mustDecimal(s string) types.Decimal {
	d := new(decimal.Big)
	d, ok := d.SetString(s)
	if !ok {
		panic("invalid decimal: " + s)
	}
	return types.NewDecimal(d)
}

type testCaseMatchResult struct {
	name            string
	filter          matching.QueryFilterMatchResult
	populationFile  string
	extraAssertions func(th *testsuite.Helper, paginated *matching.MatchResultPaginated, err error)
}

func assertMatchResultsSuccess(th *testsuite.Helper, matchResults []matching.MatchResult, err error) []matching.User {
	require.NotEmpty(th.T, matchResults, "expecting match results")

	ctx := context.Background()
	exec := th.BackendAppDb()

	userStore := th.FakeContainer().GetStoreMatching().UserStore
	users, err := userStore.Users(ctx, exec, &matching.QueryFilterUser{})
	require.NoError(th.T, err, "fetching users for validation")
	require.NotEmpty(th.T, users, "expecting users")

	for _, mr := range matchResults {
		require.NotNil(th.T, mr.UserADetails, "expecting enriched UserADetails")
		require.NotNil(th.T, mr.UserBDetails, "expecting enriched UserBDetails")
		require.NotNil(th.T, mr.UserAProfile, "expecting enriched UserAProfile")
		require.NotNil(th.T, mr.UserBProfile, "expecting enriched UserBProfile")

		// Assert UserADetails has all fields populated
		assert.NotEqual(th.T, uuid.Nil, mr.UserADetails.ID, "UserADetails.ID should not be nil")
		assert.True(th.T, mr.UserADetails.Age.Valid, "UserADetails.Age should be valid")
		assert.NotZero(th.T, mr.UserADetails.Age.Int, "UserADetails.Age should not be zero")
		assert.True(th.T, mr.UserADetails.Gender.Valid, "UserADetails.Gender should be valid")
		assert.NotEmpty(th.T, mr.UserADetails.Gender.String, "UserADetails.Gender should not be empty")
		assert.True(th.T, mr.UserADetails.Height.Valid, "UserADetails.Height should be valid")
		assert.NotZero(th.T, mr.UserADetails.Height.Float64, "UserADetails.Height should not be zero")
		assert.True(th.T, mr.UserADetails.Latitude.Valid, "UserADetails.Latitude should be valid")
		assert.NotZero(th.T, mr.UserADetails.Latitude.Float64, "UserADetails.Latitude should not be zero")
		assert.True(th.T, mr.UserADetails.Longitude.Valid, "UserADetails.Longitude should be valid")
		assert.NotZero(th.T, mr.UserADetails.Longitude.Float64, "UserADetails.Longitude should not be zero")
		assert.NotEmpty(th.T, mr.UserADetails.DatingPreferences, "UserADetails.DatingPreferences should not be empty")

		// Assert UserBDetails has all fields populated
		assert.NotEqual(th.T, uuid.Nil, mr.UserBDetails.ID, "UserBDetails.ID should not be nil")
		assert.True(th.T, mr.UserBDetails.Age.Valid, "UserBDetails.Age should be valid")
		assert.NotZero(th.T, mr.UserBDetails.Age.Int, "UserBDetails.Age should not be zero")
		assert.True(th.T, mr.UserBDetails.Gender.Valid, "UserBDetails.Gender should be valid")
		assert.NotEmpty(th.T, mr.UserBDetails.Gender.String, "UserBDetails.Gender should not be empty")
		assert.True(th.T, mr.UserBDetails.Height.Valid, "UserBDetails.Height should be valid")
		assert.NotZero(th.T, mr.UserBDetails.Height.Float64, "UserBDetails.Height should not be zero")
		assert.True(th.T, mr.UserBDetails.Latitude.Valid, "UserBDetails.Latitude should be valid")
		assert.NotZero(th.T, mr.UserBDetails.Latitude.Float64, "UserBDetails.Latitude should not be zero")
		assert.True(th.T, mr.UserBDetails.Longitude.Valid, "UserBDetails.Longitude should be valid")
		assert.NotZero(th.T, mr.UserBDetails.Longitude.Float64, "UserBDetails.Longitude should not be zero")
		assert.NotEmpty(th.T, mr.UserBDetails.DatingPreferences, "UserBDetails.DatingPreferences should not be empty")

		// Assert UserAProfile has all sections populated
		assert.NotEmpty(th.T, mr.UserAProfile.Qualitative.Interests, "UserAProfile.Qualitative.Interests should not be empty")
		assert.NotEmpty(th.T, mr.UserAProfile.Qualitative.WellbeingHabits, "UserAProfile.Qualitative.WellbeingHabits should not be empty")
		assert.NotEmpty(th.T, mr.UserAProfile.Qualitative.SelfPortrait, "UserAProfile.Qualitative.SelfPortrait should not be empty")
		assert.NotZero(th.T, mr.UserAProfile.Quantitative.ExtroversionSocialEnergy, "UserAProfile.Quantitative.ExtroversionSocialEnergy should not be zero")
		assert.NotZero(th.T, mr.UserAProfile.Quantitative.Agreeableness, "UserAProfile.Quantitative.Agreeableness should not be zero")
		assert.NotEmpty(th.T, mr.UserAProfile.Categorical.ConflictResolutionStyle, "UserAProfile.Categorical.ConflictResolutionStyle should not be empty")

		// Assert UserBProfile has all sections populated
		assert.NotEmpty(th.T, mr.UserBProfile.Qualitative.Interests, "UserBProfile.Qualitative.Interests should not be empty")
		assert.NotEmpty(th.T, mr.UserBProfile.Qualitative.WellbeingHabits, "UserBProfile.Qualitative.WellbeingHabits should not be empty")
		assert.NotEmpty(th.T, mr.UserBProfile.Qualitative.SelfPortrait, "UserBProfile.Qualitative.SelfPortrait should not be empty")
		assert.NotZero(th.T, mr.UserBProfile.Quantitative.ExtroversionSocialEnergy, "UserBProfile.Quantitative.ExtroversionSocialEnergy should not be zero")
		assert.NotZero(th.T, mr.UserBProfile.Quantitative.Agreeableness, "UserBProfile.Quantitative.Agreeableness should not be zero")
		assert.NotEmpty(th.T, mr.UserBProfile.Categorical.ConflictResolutionStyle, "UserBProfile.Categorical.ConflictResolutionStyle should not be empty")
	}

	return users
}

func TestLogic_MatchConfig_Success(t *testing.T) {

	testCases := []struct {
		name            string
		filter          matching.QueryFilterMatchConfig
		extraAssertions func(th *testsuite.Helper, config *matching.Config, err error)
	}{
		{
			name:   "returns-config-with-no-filter",
			filter: matching.QueryFilterMatchConfig{},
			extraAssertions: func(th *testsuite.Helper, config *matching.Config, err error) {
				require.NoError(th.T, err, "fetching config")
				require.NotNil(th.T, config, "expecting config")

				// Default values from migration 21
				assert.NotEmpty(th.T, config.DropHours, "expecting drop hours")
				assert.NotEmpty(th.T, config.DropHoursUTC, "expecting drop hours UTC")
				assert.NotZero(th.T, config.ID, "expecting config ID")
			},
		},
		{
			name: "filter-by-drop-hour-existing",
			filter: matching.QueryFilterMatchConfig{
				DropHour: null.StringFrom("19:00"),
			},
			extraAssertions: func(th *testsuite.Helper, config *matching.Config, err error) {
				require.NoError(th.T, err, "fetching config")
				require.NotNil(th.T, config, "expecting config")

				// Verify the config has the drop hour we filtered by
				assert.Contains(th.T, config.DropHours, "19:00", "expecting 19:00 in drop hours")
			},
		},
		{
			name: "filter-by-drop-hour-non-existing",
			filter: matching.QueryFilterMatchConfig{
				DropHour: null.StringFrom("03:00"), // not in default
			},
			extraAssertions: func(th *testsuite.Helper, config *matching.Config, err error) {
				require.Error(th.T, err, "expecting error for non-existing drop hour")
				require.Nil(th.T, config, "expecting nil config")
			},
		},
		{
			name: "filter-by-drop-hour-utc",
			filter: matching.QueryFilterMatchConfig{
				DropHourUTC: null.StringFrom("GMT+3"),
			},
			extraAssertions: func(th *testsuite.Helper, config *matching.Config, err error) {
				require.NoError(th.T, err, "fetching config")
				require.NotNil(th.T, config, "expecting config")

				assert.Contains(th.T, config.DropHoursUTC, "GMT+3", "expecting GMT+3 in drop hours UTC")
			},
		},
		{
			name: "filter-has-drop-hours-true",
			filter: matching.QueryFilterMatchConfig{
				HasDropHours: null.BoolFrom(true),
			},
			extraAssertions: func(th *testsuite.Helper, config *matching.Config, err error) {
				require.NoError(th.T, err, "fetching config")
				require.NotNil(th.T, config, "expecting config")

				assert.NotEmpty(th.T, config.DropHours, "expecting non-empty drop hours")
			},
		},
		{
			name: "filter-location-radius-range",
			filter: matching.QueryFilterMatchConfig{
				LocationRadiusKMMin: null.Float64From(100),
				LocationRadiusKMMax: null.Float64From(300),
			},
			extraAssertions: func(th *testsuite.Helper, config *matching.Config, err error) {
				require.NoError(th.T, err, "fetching config")
				require.NotNil(th.T, config, "expecting config")

				// Default is 200km, should be in range [100, 300]
				assert.GreaterOrEqual(th.T, config.LocationRadiusKM, float64(100), "location radius >= 100")
				assert.LessOrEqual(th.T, config.LocationRadiusKM, float64(300), "location radius <= 300")
			},
		},
		{
			name: "filter-score-range",
			filter: matching.QueryFilterMatchConfig{
				ScoreRangeStartMin: null.Float64From(0.5),
				ScoreRangeEndMax:   null.Float64From(0.7),
			},
			extraAssertions: func(th *testsuite.Helper, config *matching.Config, err error) {
				require.NoError(th.T, err, "fetching config")
				require.NotNil(th.T, config, "expecting config")

				// Default is 0.52-0.60
				assert.GreaterOrEqual(th.T, config.ScoreRangeStart, float64(0.5), "score range start >= 0.5")
				assert.LessOrEqual(th.T, config.ScoreRangeEnd, float64(0.7), "score range end <= 0.7")
			},
		},
		{
			name: "filter-match-expiration-hours",
			filter: matching.QueryFilterMatchConfig{
				MatchExpirationHoursMin: null.IntFrom(48),
				MatchExpirationHoursMax: null.IntFrom(96),
			},
			extraAssertions: func(th *testsuite.Helper, config *matching.Config, err error) {
				require.NoError(th.T, err, "fetching config")
				require.NotNil(th.T, config, "expecting config")

				// Default is 72 hours
				assert.GreaterOrEqual(th.T, config.MatchExpirationHours, 48, "match expiration >= 48")
				assert.LessOrEqual(th.T, config.MatchExpirationHours, 96, "match expiration <= 96")
			},
		},
		{
			name: "filter-stale-chat-nudge-range",
			filter: matching.QueryFilterMatchConfig{
				StaleChatNudgeMin: null.IntFrom(12),
				StaleChatNudgeMax: null.IntFrom(48),
			},
			extraAssertions: func(th *testsuite.Helper, config *matching.Config, err error) {
				require.NoError(th.T, err, "fetching config")
				require.NotNil(th.T, config, "expecting config")

				// Default is 24 hours
				assert.GreaterOrEqual(th.T, config.StaleChatNudge, 12, "stale chat nudge >= 12")
				assert.LessOrEqual(th.T, config.StaleChatNudge, 48, "stale chat nudge <= 48")
			},
		},
		{
			name: "filter-combined-drop-hour-and-location",
			filter: matching.QueryFilterMatchConfig{
				DropHour:            null.StringFrom("20:00"),
				LocationRadiusKMMin: null.Float64From(150),
			},
			extraAssertions: func(th *testsuite.Helper, config *matching.Config, err error) {
				require.NoError(th.T, err, "fetching config")
				require.NotNil(th.T, config, "expecting config")

				assert.Contains(th.T, config.DropHours, "20:00", "expecting 20:00 in drop hours")
				assert.GreaterOrEqual(th.T, config.LocationRadiusKM, float64(150), "location radius >= 150")
			},
		},
		{
			name: "filter-out-of-range-returns-no-config",
			filter: matching.QueryFilterMatchConfig{
				LocationRadiusKMMin: null.Float64From(1000), // default is 200
			},
			extraAssertions: func(th *testsuite.Helper, config *matching.Config, err error) {
				require.Error(th.T, err, "expecting error for out-of-range filter")
				require.Nil(th.T, config, "expecting nil config")
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotEmpty(t, tc.name, "test case name required")
			require.NotNil(t, tc.extraAssertions, "extraAssertions function required")

			testSuite := testsuite.New(t)
			t.Cleanup(testSuite.UseBackendDB())

			matchLib := testSuite.FakeContainer().GetLibMatching()

			ctx := context.Background()
			exec := testSuite.BackendAppDb()

			config, err := matchLib.MatchConfig(ctx, exec, &tc.filter)
			tc.extraAssertions(testSuite, config, err)
		})
	}
}

func TestLogic_MatchConfig_AllFieldsFilter(t *testing.T) {

	testSuite := testsuite.New(t)
	t.Cleanup(testSuite.UseBackendDB())

	ctx := context.Background()
	exec := testSuite.BackendAppDb()

	// Delete existing config to start fresh
	_, err := exec.ExecContext(ctx, "DELETE FROM match_config")
	require.NoError(t, err, "deleting existing configs")

	// Seed a config with ALL filterable fields set to specific values
	factory.NewEntity[*wingedFactory.MatchConfig](&wingedFactory.MatchConfig{
		Subject: &pgmodel.MatchConfig{
			// Age range fields
			AgeRangeStart:        null.IntFrom(21),
			AgeRangeEnd:          null.IntFrom(35),
			AgeRangeWomanOlderBy: 5,
			AgeRangeManOlderBy:   10,
			// Height field - uses types.Decimal
			HeightMaleGreaterByCM: mustDecimal("8.5"),
			// Location fields
			LocationRadiusKM:          mustDecimal("175"),
			LocationAdaptiveExpansion: types.Int64Array{200, 350, 500},
			// Drop hours fields
			DropHours:    types.StringArray{"18:00", "19:00", "20:00"},
			DropHoursUtc: types.StringArray{"GMT+5"},
			// Stale chat fields
			StaleChatNudge:      36,
			StaleChatAgentSetup: 96,
			// Match expiration/block fields
			MatchExpirationHours: 48,
			MatchBlockDeclined:   168,
			MatchBlockIgnored:    168,
			MatchBlockClosed:     168,
			// Score range fields
			ScoreRangeStart: mustDecimal("0.52"),
			ScoreRangeEnd:   mustDecimal("0.58"),
		},
	}).New(t, exec)

	matchLib := testSuite.FakeContainer().GetLibMatching()

	// Query using ALL available filter fields that match our seeded config
	filter := &matching.QueryFilterMatchConfig{
		// Age range filters (seeded: start=21, end=35)
		AgeRangeStartMin: null.IntFrom(20),
		AgeRangeStartMax: null.IntFrom(25),
		AgeRangeEndMin:   null.IntFrom(30),
		AgeRangeEndMax:   null.IntFrom(40),

		// Height filter (seeded: 8.5)
		HeightMaleGreaterByCMMin: null.Float64From(8.0),
		HeightMaleGreaterByCMMax: null.Float64From(9.0),

		// Location radius filter (seeded: 175)
		LocationRadiusKMMin: null.Float64From(170),
		LocationRadiusKMMax: null.Float64From(180),

		// Drop hours filters (seeded: ["18:00","19:00","20:00"], ["GMT+5"])
		DropHour:     null.StringFrom("19:00"),
		DropHourUTC:  null.StringFrom("GMT+5"),
		HasDropHours: null.BoolFrom(true),

		// Stale chat filters (seeded: nudge=36, agent_setup=96)
		StaleChatNudgeMin:      null.IntFrom(35),
		StaleChatNudgeMax:      null.IntFrom(40),
		StaleChatAgentSetupMin: null.IntFrom(90),
		StaleChatAgentSetupMax: null.IntFrom(100),

		// Match expiration filter (seeded: 48)
		MatchExpirationHoursMin: null.IntFrom(45),
		MatchExpirationHoursMax: null.IntFrom(50),

		// Score range filters (seeded: start=0.52, end=0.58)
		ScoreRangeStartMin: null.Float64From(0.50),
		ScoreRangeStartMax: null.Float64From(0.55),
		ScoreRangeEndMin:   null.Float64From(0.55),
		ScoreRangeEndMax:   null.Float64From(0.60),
	}

	config, err := matchLib.MatchConfig(ctx, exec, filter)
	require.NoError(t, err, "fetching config with all filters")
	require.NotNil(t, config, "expecting exactly one config to match")

	// Assert ALL seeded values match
	// Age range
	assert.Equal(t, 21, config.AgeRangeStart)
	assert.Equal(t, 35, config.AgeRangeEnd)
	// Height
	assert.Equal(t, 8.5, config.HeightMaleGreaterByCM)
	// Location
	assert.Equal(t, float64(175), config.LocationRadiusKM)
	// Drop hours
	assert.Contains(t, config.DropHours, "18:00")
	assert.Contains(t, config.DropHours, "19:00")
	assert.Contains(t, config.DropHours, "20:00")
	assert.Contains(t, config.DropHoursUTC, "GMT+5")
	// Stale chat
	assert.Equal(t, 36, config.StaleChatNudge)
	assert.Equal(t, 96, config.StaleChatAgentSetup)
	// Match expiration/block
	assert.Equal(t, 48, config.MatchExpirationHours)
	assert.Equal(t, 168, config.MatchBlockDeclined)
	assert.Equal(t, 168, config.MatchBlockIgnored)
	assert.Equal(t, 168, config.MatchBlockClosed)
	// Score range
	assert.Equal(t, 0.52, config.ScoreRangeStart)
	assert.Equal(t, 0.58, config.ScoreRangeEnd)

	// Verify configs count is exactly 1
	configs, err := matchLib.MatchConfigs(ctx, exec, filter)
	require.NoError(t, err, "fetching configs with all filters")
	require.Len(t, configs, 1, "expecting exactly one config to match all filters")
}

func TestLogic_MatchConfigs_Success(t *testing.T) {

	testCases := []struct {
		name            string
		filter          matching.QueryFilterMatchConfig
		extraAssertions func(th *testsuite.Helper, configs []matching.Config, err error)
	}{
		{
			name:   "returns-all-configs-with-no-filter",
			filter: matching.QueryFilterMatchConfig{},
			extraAssertions: func(th *testsuite.Helper, configs []matching.Config, err error) {
				require.NoError(th.T, err, "fetching configs")
				require.NotEmpty(th.T, configs, "expecting at least one config")

				// Verify first config has expected fields populated
				config := configs[0]
				assert.NotZero(th.T, config.ID, "expecting config ID")
				assert.NotEmpty(th.T, config.DropHours, "expecting drop hours")
			},
		},
		{
			name: "returns-configs-sorted-by-location-radius-asc",
			filter: matching.QueryFilterMatchConfig{
				OrderBy: null.StringFrom("location_radius_km"),
				Sort:    null.StringFrom("+"),
			},
			extraAssertions: func(th *testsuite.Helper, configs []matching.Config, err error) {
				require.NoError(th.T, err, "fetching configs")
				require.NotEmpty(th.T, configs, "expecting configs")

				// With single config, just verify it returns successfully
				assert.NotZero(th.T, configs[0].LocationRadiusKM, "expecting location radius")
			},
		},
		{
			name: "returns-configs-sorted-by-score-range-desc",
			filter: matching.QueryFilterMatchConfig{
				OrderBy: null.StringFrom("score_range_start"),
				Sort:    null.StringFrom("-"),
			},
			extraAssertions: func(th *testsuite.Helper, configs []matching.Config, err error) {
				require.NoError(th.T, err, "fetching configs")
				require.NotEmpty(th.T, configs, "expecting configs")

				assert.NotZero(th.T, configs[0].ScoreRangeStart, "expecting score range start")
			},
		},
		{
			name: "filter-configs-with-drop-hours-returns-results",
			filter: matching.QueryFilterMatchConfig{
				HasDropHours: null.BoolFrom(true),
			},
			extraAssertions: func(th *testsuite.Helper, configs []matching.Config, err error) {
				require.NoError(th.T, err, "fetching configs")
				require.NotEmpty(th.T, configs, "expecting configs with drop hours")

				for _, cfg := range configs {
					assert.NotEmpty(th.T, cfg.DropHours, "all configs should have drop hours")
				}
			},
		},
		{
			name: "filter-configs-without-drop-hours-returns-empty",
			filter: matching.QueryFilterMatchConfig{
				HasDropHours: null.BoolFrom(false),
			},
			extraAssertions: func(th *testsuite.Helper, configs []matching.Config, err error) {
				require.NoError(th.T, err, "fetching configs")
				// Default seeded config has drop hours, so should be empty
				require.Empty(th.T, configs, "expecting no configs without drop hours")
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotEmpty(t, tc.name, "test case name required")
			require.NotNil(t, tc.extraAssertions, "extraAssertions function required")

			testSuite := testsuite.New(t)
			t.Cleanup(testSuite.UseBackendDB())

			matchLib := testSuite.FakeContainer().GetLibMatching()

			ctx := context.Background()
			exec := testSuite.BackendAppDb()

			configs, err := matchLib.MatchConfigs(ctx, exec, &tc.filter)
			tc.extraAssertions(testSuite, configs, err)
		})
	}
}

func TestLogic_MatchResults_Success(t *testing.T) {

	testCases := []testCaseMatchResult{
		{
			name:           "paginate-all",
			populationFile: "./testdata/population_4_match_result_listing.csv",
			filter: matching.QueryFilterMatchResult{
				EnrichUsers: true,
			},
			extraAssertions: func(th *testsuite.Helper, matchResults *matching.MatchResultPaginated, err error) {
				users := assertMatchResultsSuccess(th, matchResults.Data, err)

				expected := matching.UserPairsUniqPerm(users)
				actual := len(matchResults.Data)
				require.Len(th.T, expected, actual, "matching users permuation")
			},
		},
		{
			name:           "paginate-1",
			populationFile: "./testdata/population_4_match_result_listing.csv",
			filter: matching.QueryFilterMatchResult{
				EnrichUsers: true,
				Pagination: &sdk.Pagination{
					Rows: null.IntFrom(1),
				},
			},
			extraAssertions: func(th *testsuite.Helper, matchResults *matching.MatchResultPaginated, err error) {
				assertMatchResultsSuccess(th, matchResults.Data, err)
				require.Len(th.T, matchResults.Data, 1, "expect only 1 match result due to pagination")
			},
		},
		{
			name:           "filter-is-approved-false",
			populationFile: "./testdata/population_4_match_result_listing.csv",
			filter: matching.QueryFilterMatchResult{
				IsApproved: null.BoolFrom(false),
			},
			extraAssertions: func(th *testsuite.Helper, matchResults *matching.MatchResultPaginated, err error) {
				require.NoError(th.T, err, "fetching match results")
				// All match results should have is_approved=false since none are approved yet
				for _, mr := range matchResults.Data {
					assert.False(th.T, mr.IsApproved, "expecting is_approved to be false")
				}
			},
		},
		{
			name:           "filter-is-approved-true-empty",
			populationFile: "./testdata/population_4_match_result_listing.csv",
			filter: matching.QueryFilterMatchResult{
				IsApproved: null.BoolFrom(true),
			},
			extraAssertions: func(th *testsuite.Helper, matchResults *matching.MatchResultPaginated, err error) {
				require.NoError(th.T, err, "fetching match results")
				// No match results should be approved (default is false)
				require.Empty(th.T, matchResults.Data, "expecting no approved match results")
			},
		},
		{
			name:           "filter-is-expired-false",
			populationFile: "./testdata/population_4_match_result_listing.csv",
			filter: matching.QueryFilterMatchResult{
				IsExpired: null.BoolFrom(false),
			},
			extraAssertions: func(th *testsuite.Helper, matchResults *matching.MatchResultPaginated, err error) {
				require.NoError(th.T, err, "fetching match results")
				// All match results should have is_expired=false (default)
				for _, mr := range matchResults.Data {
					assert.False(th.T, mr.IsExpired, "expecting is_expired to be false")
				}
			},
		},
		{
			name:           "filter-is-possible-match-false",
			populationFile: "./testdata/population_4_match_result_listing.csv",
			filter: matching.QueryFilterMatchResult{
				IsPossibleMatch: null.BoolFrom(false),
			},
			extraAssertions: func(th *testsuite.Helper, matchResults *matching.MatchResultPaginated, err error) {
				require.NoError(th.T, err, "fetching match results")
				// All match results should have is_possible_match=false (default)
				for _, mr := range matchResults.Data {
					assert.False(th.T, mr.IsPossibleMatch, "expecting is_possible_match to be false")
				}
			},
		},
		{
			name:           "filter-combined-for-drop-empty",
			populationFile: "./testdata/population_4_match_result_listing.csv",
			filter: matching.QueryFilterMatchResult{
				IsApproved:      null.BoolFrom(true),
				IsPossibleMatch: null.BoolFrom(true),
				IsExpired:       null.BoolFrom(false),
			},
			extraAssertions: func(th *testsuite.Helper, matchResults *matching.MatchResultPaginated, err error) {
				require.NoError(th.T, err, "fetching match results")
				// No results expected since none are approved yet
				require.Empty(th.T, matchResults.Data, "expecting no match results ready for drop")
			},
		},
		{
			name:           "match-results-for-drop-after-update",
			populationFile: "./testdata/population_4_match_result_listing.csv",
			filter:         matching.QueryFilterMatchResult{}, // not used directly, we use MatchResultsForDrop
			extraAssertions: func(th *testsuite.Helper, _ *matching.MatchResultPaginated, _ error) {
				ctx := context.Background()
				exec := th.BackendAppDb()
				aiExec := th.AiBackendDb()
				matchLib := th.FakeContainer().GetLibMatching()

				// Initially, no match results should be ready for drop
				dropResults, err := matchLib.MatchResultsForDrop(ctx, exec)
				require.NoError(th.T, err, "fetching match results for drop")
				require.Empty(th.T, dropResults.Data, "expecting no match results ready for drop initially")

				// Get all match results
				allResults, err := matchLib.MatchResults(ctx, exec, aiExec, &matching.QueryFilterMatchResult{})
				require.NoError(th.T, err, "fetching all match results")
				require.NotEmpty(th.T, allResults.Data, "expecting match results")

				// Update one match result to be ready for drop
				storeMatching := th.FakeContainer().GetStoreMatching()
				firstResult := allResults.Data[0]
				_, err = storeMatching.MatchResultStore.Update(ctx, exec, &matching.UpdateMatchResult{
					ID:              firstResult.ID,
					IsApproved:      null.BoolFrom(true),
					IsPossibleMatch: null.BoolFrom(true),
					IsExpired:       null.BoolFrom(false),
				})
				require.NoError(th.T, err, "updating match result for drop")

				// Now we should have one match result ready for drop
				dropResults, err = matchLib.MatchResultsForDrop(ctx, exec)
				require.NoError(th.T, err, "fetching match results for drop")
				require.Len(th.T, dropResults.Data, 1, "expecting one match result ready for drop")
				assert.Equal(th.T, firstResult.ID, dropResults.Data[0].ID, "expecting the updated match result")
				assert.True(th.T, dropResults.Data[0].IsApproved, "expecting is_approved to be true")
				assert.True(th.T, dropResults.Data[0].IsPossibleMatch, "expecting is_possible_match to be true")
				assert.False(th.T, dropResults.Data[0].IsExpired, "expecting is_expired to be false")
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotEmpty(t, tc.name, "test case name required")
			require.NotEmpty(t, tc.populationFile, "population file required")
			require.NotNil(t, tc.extraAssertions, "extraAssertions function required")

			testSuite := testsuite.New(t)
			t.Cleanup(testSuite.UseBackendDB())
			t.Cleanup(testSuite.UseAiDB())
			t.Cleanup(testSuite.UseSupabaseAuthDB())

			matchLib := testSuite.FakeContainer().GetLibMatching()

			// ingest population data
			ingestor := testhelper.NewPopulationIngestor(t, testSuite, matchLib)
			parseResult, populateResult, err := ingestor.IngestFromCSVFile(tc.populationFile)
			require.NoError(t, err, "ingesting population data")
			require.Equal(t, parseResult.ValidRows, populateResult.BackendAppUsers, "populate should create all users")
			require.Equal(t, parseResult.ValidRows, populateResult.AIBackendProfiles, "populate should create all profiles")
			require.Empty(t, populateResult.Errors, "expected no populate errors")

			// ingest population data into a match set
			matchSet, err := matchLib.IngestAll(context.Background(), testSuite.BackendAppDb())
			require.NoError(t, err, "ingesting matches")
			require.NotNil(t, matchSet, "match set returned")
			require.Equal(t, parseResult.ValidRows, matchSet.NumberOfParticipants, "match set participants")

			ctx := context.Background()
			exec := testSuite.BackendAppDb()
			aiExec := testSuite.AiBackendDb()

			paginatedMatchResults, err := matchLib.MatchResults(ctx, exec, aiExec, &tc.filter)
			tc.extraAssertions(testSuite, paginatedMatchResults, err)
		})
	}
}
