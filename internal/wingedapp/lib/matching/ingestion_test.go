package matching_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"
	"wingedapp/pgtester/internal/wingedapp/lib/matching/store"
	"wingedapp/pgtester/internal/wingedapp/lib/matching/testhelper"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCaseIngestion struct {
	name           string
	users          any // do I seed with specific users with X parameters? hmmm
	qResults       *matching.QualifierResults
	populationFile string
}

func TestIngestion(t *testing.T) {

	testCases := []testCaseIngestion{
		{
			name:           "basic match test",
			populationFile: "./testdata/population_1.csv",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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

			ms := store.NewMatchingStores(applog.NewLogrus("test"))
			matchSets, err := ms.MatchSetStore.MatchSets(context.Background(),
				testSuite.BackendAppDb(),
				&matching.QueryFilterMatchSet{},
			)
			require.NoError(t, err, "fetching match sets")
			require.Len(t, matchSets.Data, 1, "expecting single match set")

			paginateMatchResults, err := ms.MatchResultStore.MatchResults(context.Background(),
				testSuite.BackendAppDb(),
				&matching.QueryFilterMatchResult{
					MatchSetID: null.StringFrom(matchSets.Data[0].ID.String()),
				},
			)

			users, err := ms.UserStore.Users(context.Background(),
				testSuite.BackendAppDb(),
				&matching.QueryFilterUser{},
			)
			require.NoError(t, err, "fetching users for permutation calc")
			require.NotEmpty(t, users, "users exist for permutation calc")

			require.Len(t, paginateMatchResults.Data, len(matching.UserPairsUniqPerm(users)), "expecting single match set")
		})
	}
}

// TestIngestion_CSVParsingAllFields verifies that CSV parsing correctly extracts
// all fields from the new flat column format (Niina's "test group v1").
func TestIngestion_CSVParsingAllFields(t *testing.T) {

	// Read and parse the CSV file directly
	file, err := os.Open("./testdata/population_1.csv")
	require.NoError(t, err, "opening test CSV file")
	defer file.Close()

	result, err := matching.ParsePopulationCSV(file)
	require.NoError(t, err, "parsing CSV file")
	require.NotNil(t, result, "parse result should not be nil")
	require.Equal(t, 6, result.ValidRows, "expected 6 valid rows")
	require.Empty(t, result.ParsingErrors, "expected no parsing errors")
	require.Empty(t, result.ValidationErrs, "expected no validation errors")

	// Verify first row has all expected fields
	row := result.Rows[0]
	assert.Equal(t, "Demby", row.FirstName, "first_name")
	assert.Equal(t, "Abella", row.LastName, "last_name")
	assert.Equal(t, "Los Angeles, CA", row.Address, "address")
	assert.Equal(t, 1998, row.Birthday.Year(), "birthday year")
	// Age is calculated from birthday - verify it's reasonable (not hardcoded to avoid date drift)
	expectedAge := calculateTestAge(row.Birthday)
	assert.Equal(t, expectedAge, row.Age, "age (calculated from birthday)")
	assert.Equal(t, "Male", row.Gender, "gender")
	assert.Equal(t, 175, row.HeightCM, "height_cm")
	assert.InDelta(t, 34.0522, row.Latitude, 0.0001, "latitude")
	assert.InDelta(t, -118.2437, row.Longitude, 0.0001, "longitude")
	assert.Equal(t, "Female", row.DatingPreference, "dating_preference")
	assert.Equal(t, []string{"Female"}, row.DatingPreferences, "dating preferences expanded")

	// Verify profile_details is parsed correctly from flat columns
	require.NotNil(t, row.ProfileDetails, "profile_details should not be nil")
	pd := row.ProfileDetails

	// Qualitative fields
	assert.Equal(t, "Thoughtful and tech-driven", pd.SelfPortrait.String, "self_portrait")
	assert.Equal(t, "Tech and Programming", pd.Interests.String, "interests")
	assert.Equal(t, "Exercise and Morning routine", pd.WellbeingHabits.String, "wellbeing_habits")
	assert.Equal(t, "Gym and Meditation", pd.SelfCareHabits.String, "self_care_habits")
	assert.Equal(t, "Budgeting and Saving", pd.MoneyManagement.String, "money_management")
	assert.Equal(t, "Daily reflection", pd.SelfReflectionCapabilities.String, "self_reflection_capabilities")
	assert.Equal(t, "Utilitarian", pd.MoralFrameworks.String, "moral_frameworks")
	assert.Equal(t, "Career growth and Innovation", pd.LifeGoals.String, "life_goals")
	assert.Equal(t, "Trust and Communication", pd.PartnershipValues.String, "partnership_values")
	assert.Equal(t, "Long-term commitment", pd.MutualCommitment.String, "mutual_commitment")
	assert.Equal(t, "Open and Growth-minded", pd.SpiritualityGrowthMindset.String, "spirituality_growth_mindset")
	assert.Equal(t, "Multicultural", pd.CulturalValues.String, "cultural_values")
	assert.Equal(t, "Open to children", pd.FamilyPlanning.String, "family_planning")
	assert.Equal(t, "Coffee shop conversation", pd.IdealDate.String, "ideal_date")
	assert.Equal(t, "Green: humor, Red: dishonesty", pd.RedGreenFlags.String, "red_green_flags")

	// Quantitative fields (0-1 floats in new format)
	assert.InDelta(t, 0.7, pd.ExtroversionSocialEnergy.Float64, 0.01, "extroversion_social_energy")
	assert.InDelta(t, 0.5, pd.RoutineVsSpontaneity.Float64, 0.01, "routine_vs_spontaneity")
	assert.InDelta(t, 0.8, pd.Agreeableness.Float64, 0.01, "agreeableness")
	assert.InDelta(t, 0.7, pd.Conscientiousness.Float64, 0.01, "conscientiousness")
	assert.InDelta(t, 0.3, float64(pd.Neuroticism.Float32), 0.01, "neuroticism")
	assert.InDelta(t, 0.4, pd.DominanceLevel.Float64, 0.01, "dominance_level")
	assert.InDelta(t, 0.6, pd.EmotionalExpressiveness.Float64, 0.01, "emotional_expressiveness")
	assert.InDelta(t, 0.6, pd.SexDrive.Float64, 0.01, "sex_drive")
	assert.InDelta(t, 0.5, pd.GeographicalMobility.Float64, 0.01, "geographical_mobility")

	// Categorical fields
	assert.Equal(t, "collaborative", pd.ConflictResolutionStyle.String, "conflict_resolution_style")
	assert.Equal(t, "heterosexual", pd.SexualityPreferences.String, "sexuality_preferences")
	assert.Equal(t, "agnostic", pd.Religion.String, "religion")

	// Verify all rows have different dating preferences
	datingPrefs := make(map[string]bool)
	for _, r := range result.Rows {
		datingPrefs[r.DatingPreference] = true
	}
	assert.True(t, len(datingPrefs) >= 2, "should have variety of dating preferences")
}

// TestIngestion_UserFieldsPersisted verifies that user fields are correctly
// stored in the database after ingestion.
func TestIngestion_UserFieldsPersisted(t *testing.T) {

	testSuite := testsuite.New(t)
	t.Cleanup(testSuite.UseBackendDB())
	t.Cleanup(testSuite.UseAiDB())
	t.Cleanup(testSuite.UseSupabaseAuthDB())

	matchLib := testSuite.FakeContainer().GetLibMatching()

	// Ingest population data
	ingestor := testhelper.NewPopulationIngestor(t, testSuite, matchLib)
	parseResult, populateResult, err := ingestor.IngestFromCSVFile("./testdata/population_1.csv")
	require.NoError(t, err, "ingesting population data")
	require.Empty(t, populateResult.Errors, "expected no populate errors")

	ms := store.NewMatchingStores(applog.NewLogrus("test"))
	ctx := context.Background()

	// Fetch all users
	users, err := ms.UserStore.Users(ctx, testSuite.BackendAppDb(), &matching.QueryFilterUser{})
	require.NoError(t, err, "fetching users")
	require.Len(t, users, parseResult.ValidRows, "expected %d users", parseResult.ValidRows)

	// Verify user fields for a user born in 1998 (age calculated dynamically)
	expectedAge := calculateTestAge(time.Date(1998, 6, 15, 0, 0, 0, 0, time.UTC))
	var firstUser *matching.User
	for i := range users {
		// We can't guarantee order, so we verify at least one user has expected values
		if users[i].Gender.String == "Male" && users[i].Age.Int == expectedAge {
			firstUser = &users[i]
			break
		}
	}
	require.NotNil(t, firstUser, "should find a %d-year-old Male user (born 1998)", expectedAge)

	assert.Equal(t, expectedAge, firstUser.Age.Int, "age")
	assert.Equal(t, "Male", firstUser.Gender.String, "gender")
	assert.InDelta(t, 175.0, firstUser.Height.Float64, 0.1, "height")
	assert.InDelta(t, 34.0522, firstUser.Latitude.Float64, 0.0001, "latitude")
	assert.InDelta(t, -118.2437, firstUser.Longitude.Float64, 0.0001, "longitude")

	// Verify gender distribution
	genderCounts := make(map[string]int)
	for _, u := range users {
		genderCounts[u.Gender.String]++
	}
	assert.Equal(t, 6, genderCounts["Male"], "should have 6 Male users")

	// Verify FirstName/LastName persisted in DB using pgmodel (look up by first+last name pattern)
	dbUsers, err := pgmodel.Users(
		pgmodel.UserWhere.FirstName.EQ(null.StringFrom("Demby")),
		pgmodel.UserWhere.LastName.EQ(null.StringFrom("Abella")),
	).All(ctx, testSuite.BackendAppDb())
	require.NoError(t, err, "fetching user by name from pgmodel")
	require.Len(t, dbUsers, 1, "should find exactly one Demby Abella")
	assert.Equal(t, "Demby", dbUsers[0].FirstName.String, "firstname should be persisted in DB")
	assert.Equal(t, "Abella", dbUsers[0].LastName.String, "lastname should be persisted in DB")

	// Verify another user's FirstName/LastName
	dbUsers2, err := pgmodel.Users(
		pgmodel.UserWhere.FirstName.EQ(null.StringFrom("Alice")),
		pgmodel.UserWhere.LastName.EQ(null.StringFrom("Smith")),
	).All(ctx, testSuite.BackendAppDb())
	require.NoError(t, err, "fetching second user by name from pgmodel")
	require.Len(t, dbUsers2, 1, "should find exactly one Alice Smith")
	assert.Equal(t, "Alice", dbUsers2[0].FirstName.String, "firstname should be persisted in DB")
	assert.Equal(t, "Smith", dbUsers2[0].LastName.String, "lastname should be persisted in DB")
}

// TestIngestion_DatingPreferencesPersisted verifies that dating preferences
// are correctly stored in the database.
func TestIngestion_DatingPreferencesPersisted(t *testing.T) {

	testSuite := testsuite.New(t)
	t.Cleanup(testSuite.UseBackendDB())
	t.Cleanup(testSuite.UseAiDB())
	t.Cleanup(testSuite.UseSupabaseAuthDB())

	matchLib := testSuite.FakeContainer().GetLibMatching()

	// Ingest population data
	ingestor := testhelper.NewPopulationIngestor(t, testSuite, matchLib)
	parseResult, populateResult, err := ingestor.IngestFromCSVFile("./testdata/population_1.csv")
	require.NoError(t, err, "ingesting population data")
	require.Empty(t, populateResult.Errors, "expected no populate errors")

	// Verify dating preferences were created
	// The CSV has 6 users, first user has 3 preferences (Male,Female,Non-Binary)
	expectedMinPreferences := 6 // at least 1 per user
	assert.GreaterOrEqual(t, populateResult.DatingPreferences, expectedMinPreferences,
		"expected at least %d dating preferences", expectedMinPreferences)

	ms := store.NewMatchingStores(applog.NewLogrus("test"))
	ctx := context.Background()

	// Fetch all users
	users, err := ms.UserStore.Users(ctx, testSuite.BackendAppDb(), &matching.QueryFilterUser{})
	require.NoError(t, err, "fetching users")
	require.Len(t, users, parseResult.ValidRows, "expected users")

	// Verify at least one user has multiple dating preferences
	var foundMultiPrefs bool
	for _, u := range users {
		prefs, err := ms.UserDatingPrefsStore.UserDatingPreferences(ctx,
			testSuite.BackendAppDb(),
			&matching.QueryFilterUserDatingPrefs{
				UserID: null.StringFrom(u.ID.String()),
			},
		)
		require.NoError(t, err, "fetching dating preferences for user %s", u.ID)

		if len(prefs) >= 3 {
			foundMultiPrefs = true
			// Verify the preferences are valid genders
			for _, pref := range prefs {
				assert.Contains(t, []string{"Male", "Female", "Non-Binary"}, pref.DatingPreference,
					"dating preference should be a valid gender")
			}
		}
	}
	assert.True(t, foundMultiPrefs, "should find at least one user with 3+ dating preferences")
}

// TestIngestion_ProfileDetailsPersisted verifies that profile details (qualitative,
// quantitative, categorical) are correctly stored in the ai_backend database.
func TestIngestion_ProfileDetailsPersisted(t *testing.T) {

	testSuite := testsuite.New(t)
	t.Cleanup(testSuite.UseBackendDB())
	t.Cleanup(testSuite.UseAiDB())
	t.Cleanup(testSuite.UseSupabaseAuthDB())

	matchLib := testSuite.FakeContainer().GetLibMatching()

	// Ingest population data
	ingestor := testhelper.NewPopulationIngestor(t, testSuite, matchLib)
	parseResult, populateResult, err := ingestor.IngestFromCSVFile("./testdata/population_1.csv")
	require.NoError(t, err, "ingesting population data")
	require.Empty(t, populateResult.Errors, "expected no populate errors")
	require.Equal(t, parseResult.ValidRows, populateResult.AIBackendProfiles, "should create profiles for all users")

	ms := store.NewMatchingStores(applog.NewLogrus("test"))
	ctx := context.Background()
	aiExec := testSuite.AiBackendDb()

	// Fetch all users
	users, err := ms.UserStore.Users(ctx, testSuite.BackendAppDb(), &matching.QueryFilterUser{})
	require.NoError(t, err, "fetching users")

	// Verify profile for each user
	profilesFound := 0
	for _, u := range users {
		profile, err := ms.ProfileStore.Profile(ctx, aiExec, u.ID)
		require.NoError(t, err, "fetching profile for user %s", u.ID)

		if profile != nil {
			profilesFound++

			// Verify qualitative section is populated (fields that exist in PersonProfile)
			assert.NotEmpty(t, profile.Qualitative.WellbeingHabits, "wellbeing_habits should not be empty")
			assert.NotEmpty(t, profile.Qualitative.Interests, "interests should not be empty")
			assert.NotEmpty(t, profile.Qualitative.SelfCareHabits, "self_care_habits should not be empty")
			assert.NotEmpty(t, profile.Qualitative.SelfPortrait, "self_portrait should not be empty")
			assert.NotEmpty(t, profile.Qualitative.LifeGoals, "life_goals should not be empty")
			assert.NotEmpty(t, profile.Qualitative.PartnershipValues, "partnership_values should not be empty")
			assert.NotEmpty(t, profile.Qualitative.MutualCommitment, "mutual_commitment should not be empty")
			assert.NotEmpty(t, profile.Qualitative.FamilyPlanning, "family_planning should not be empty")
			assert.NotEmpty(t, profile.Qualitative.IdealDate, "ideal_date should not be empty")
			assert.NotEmpty(t, profile.Qualitative.RedGreenFlags, "red_green_flags should not be empty")

			// Verify quantitative section has valid ranges (0-1 floats)
			assert.GreaterOrEqual(t, profile.Quantitative.ExtroversionSocialEnergy, 0.0, "extroversion >= 0")
			assert.LessOrEqual(t, profile.Quantitative.ExtroversionSocialEnergy, 1.0, "extroversion <= 1")
			assert.GreaterOrEqual(t, profile.Quantitative.Agreeableness, 0.0, "agreeableness >= 0")
			assert.LessOrEqual(t, profile.Quantitative.Agreeableness, 1.0, "agreeableness <= 1")
			assert.GreaterOrEqual(t, profile.Quantitative.Conscientiousness, 0.0, "conscientiousness >= 0")
			assert.LessOrEqual(t, profile.Quantitative.Conscientiousness, 1.0, "conscientiousness <= 1")
			assert.GreaterOrEqual(t, profile.Quantitative.Neuroticism, 0.0, "neuroticism >= 0")
			assert.LessOrEqual(t, profile.Quantitative.Neuroticism, 1.0, "neuroticism <= 1")
			assert.GreaterOrEqual(t, profile.Quantitative.DominanceLevel, 0.0, "dominance_level >= 0")
			assert.LessOrEqual(t, profile.Quantitative.DominanceLevel, 1.0, "dominance_level <= 1")
			assert.GreaterOrEqual(t, profile.Quantitative.EmotionalExpressiveness, 0.0, "emotional_expressiveness >= 0")
			assert.LessOrEqual(t, profile.Quantitative.EmotionalExpressiveness, 1.0, "emotional_expressiveness <= 1")
			assert.GreaterOrEqual(t, profile.Quantitative.SexDrive, 0.0, "sex_drive >= 0")
			assert.LessOrEqual(t, profile.Quantitative.SexDrive, 1.0, "sex_drive <= 1")
			assert.GreaterOrEqual(t, profile.Quantitative.GeographicalMobility, 0.0, "geographical_mobility >= 0")
			assert.LessOrEqual(t, profile.Quantitative.GeographicalMobility, 1.0, "geographical_mobility <= 1")

			// Verify categorical section is populated
			assert.NotEmpty(t, profile.Categorical.ConflictResolutionStyle, "conflict_resolution_style should not be empty")
			assert.NotEmpty(t, profile.Categorical.Religion, "religion should not be empty")
		}
	}

	assert.Equal(t, parseResult.ValidRows, profilesFound, "should find profiles for all users")
}

// TestIngestion_SupabaseUsersPersisted verifies that supabase auth users
// are correctly created during ingestion with all OTP-required fields.
func TestIngestion_SupabaseUsersPersisted(t *testing.T) {

	testSuite := testsuite.New(t)
	t.Cleanup(testSuite.UseBackendDB())
	t.Cleanup(testSuite.UseAiDB())
	t.Cleanup(testSuite.UseSupabaseAuthDB())

	matchLib := testSuite.FakeContainer().GetLibMatching()

	// Ingest population data
	ingestor := testhelper.NewPopulationIngestor(t, testSuite, matchLib)
	parseResult, populateResult, err := ingestor.IngestFromCSVFile("./testdata/population_1.csv")
	require.NoError(t, err, "ingesting population data")
	require.Empty(t, populateResult.Errors, "expected no populate errors")

	// Verify supabase users were created
	assert.Equal(t, parseResult.ValidRows, populateResult.SupabaseAuthUsers,
		"should create supabase users for all parsed rows")

	ctx := context.Background()
	db := testSuite.SupabaseAuthDb()

	// ========== VERIFY ALL OTP-REQUIRED FIELDS FOR EACH USER ==========
	var userCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&userCount)
	require.NoError(t, err, "counting supabase users")
	assert.Equal(t, parseResult.ValidRows, userCount, "should have correct number of supabase users")

	// Query all users and verify OTP fields
	rows, err := db.QueryContext(ctx, `
		SELECT id, instance_id, aud, role, raw_app_meta_data, raw_user_meta_data
		FROM users
	`)
	require.NoError(t, err, "querying supabase users")
	defer rows.Close()

	verifiedUsers := 0
	for rows.Next() {
		var userID, instanceID, aud, role string
		var rawAppMetaData, rawUserMetaData []byte

		err = rows.Scan(&userID, &instanceID, &aud, &role, &rawAppMetaData, &rawUserMetaData)
		require.NoError(t, err, "scanning user row")

		// ASSERT: instance_id
		assert.Equal(t, "00000000-0000-0000-0000-000000000000", instanceID,
			"user %s: instance_id must be default Supabase value", userID)

		// ASSERT: aud and role
		assert.Equal(t, "authenticated", aud, "user %s: aud must be 'authenticated'", userID)
		assert.Equal(t, "authenticated", role, "user %s: role must be 'authenticated'", userID)

		// ASSERT: raw_app_meta_data contains provider info (REQUIRED FOR OTP)
		require.NotNil(t, rawAppMetaData, "user %s: raw_app_meta_data must not be NULL", userID)
		var appMeta map[string]interface{}
		err = json.Unmarshal(rawAppMetaData, &appMeta)
		require.NoError(t, err, "user %s: raw_app_meta_data must be valid JSON", userID)
		assert.Equal(t, "email", appMeta["provider"], "user %s: raw_app_meta_data.provider must be 'email'", userID)
		providers, ok := appMeta["providers"].([]interface{})
		require.True(t, ok, "user %s: raw_app_meta_data.providers must be an array", userID)
		assert.Contains(t, providers, "email", "user %s: raw_app_meta_data.providers must contain 'email'", userID)

		// ASSERT: raw_user_meta_data is valid JSON
		require.NotNil(t, rawUserMetaData, "user %s: raw_user_meta_data must not be NULL", userID)
		var userMeta map[string]interface{}
		err = json.Unmarshal(rawUserMetaData, &userMeta)
		require.NoError(t, err, "user %s: raw_user_meta_data must be valid JSON", userID)

		// ASSERT: Identity record exists (REQUIRED FOR LOGIN)
		var identityCount int
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM identities WHERE user_id = $1", userID).Scan(&identityCount)
		require.NoError(t, err, "counting identities for user %s", userID)
		assert.Equal(t, 1, identityCount, "user %s: must have exactly 1 identity record", userID)

		// ASSERT: Identity has correct provider and provider_id
		var provider, providerID string
		err = db.QueryRowContext(ctx,
			"SELECT provider, provider_id FROM identities WHERE user_id = $1",
			userID,
		).Scan(&provider, &providerID)
		require.NoError(t, err, "querying identity for user %s", userID)
		assert.Equal(t, "email", provider, "user %s: identity.provider must be 'email'", userID)
		assert.Equal(t, userID, providerID, "user %s: identity.provider_id must equal user_id", userID)

		// ASSERT: Identity identity_data contains sub
		var identityDataSub string
		err = db.QueryRowContext(ctx,
			"SELECT identity_data->>'sub' FROM identities WHERE user_id = $1",
			userID,
		).Scan(&identityDataSub)
		require.NoError(t, err, "querying identity_data.sub for user %s", userID)
		assert.Equal(t, userID, identityDataSub, "user %s: identity_data.sub must equal user_id", userID)

		// ASSERT: Identity identity_data contains email_verified and phone_verified (REQUIRED FOR OTP)
		var emailVerified, phoneVerified bool
		err = db.QueryRowContext(ctx,
			"SELECT (identity_data->>'email_verified')::boolean, (identity_data->>'phone_verified')::boolean FROM identities WHERE user_id = $1",
			userID,
		).Scan(&emailVerified, &phoneVerified)
		require.NoError(t, err, "querying identity_data verified flags for user %s", userID)
		assert.True(t, emailVerified, "user %s: identity_data.email_verified must be true", userID)
		assert.False(t, phoneVerified, "user %s: identity_data.phone_verified must be false", userID)

		verifiedUsers++
	}
	require.NoError(t, rows.Err(), "iterating user rows")
	assert.Equal(t, parseResult.ValidRows, verifiedUsers, "should verify all users")
}

// TestIngestion_MatchResultsCreated verifies that match results are correctly
// created for all unique user pairs after ingestion.
func TestIngestion_MatchResultsCreated(t *testing.T) {

	testSuite := testsuite.New(t)
	t.Cleanup(testSuite.UseBackendDB())
	t.Cleanup(testSuite.UseAiDB())
	t.Cleanup(testSuite.UseSupabaseAuthDB())

	matchLib := testSuite.FakeContainer().GetLibMatching()

	// Ingest population data
	ingestor := testhelper.NewPopulationIngestor(t, testSuite, matchLib)
	parseResult, populateResult, err := ingestor.IngestFromCSVFile("./testdata/population_1.csv")
	require.NoError(t, err, "ingesting population data")
	require.Empty(t, populateResult.Errors, "expected no populate errors")

	// Create match set and run ingestion
	matchSet, err := matchLib.IngestAll(context.Background(), testSuite.BackendAppDb())
	require.NoError(t, err, "ingesting all matches")
	require.NotNil(t, matchSet, "match set should not be nil")
	require.Equal(t, parseResult.ValidRows, matchSet.NumberOfParticipants, "participants count")

	ms := store.NewMatchingStores(applog.NewLogrus("test"))
	ctx := context.Background()

	// Verify match results count
	paginatedResults, err := ms.MatchResultStore.MatchResults(ctx,
		testSuite.BackendAppDb(),
		&matching.QueryFilterMatchResult{
			MatchSetID: null.StringFrom(matchSet.ID.String()),
		},
	)
	require.NoError(t, err, "fetching match results")

	// Calculate expected unique pairs: n*(n-1)/2 for n users
	n := parseResult.ValidRows
	expectedPairs := n * (n - 1) / 2
	assert.Equal(t, expectedPairs, len(paginatedResults.Data),
		"should have %d unique match results for %d users", expectedPairs, n)

	// Verify each match result has valid user IDs
	for _, mr := range paginatedResults.Data {
		assert.NotEqual(t, uuid.Nil, mr.UserAID, "user_a_id should not be nil")
		assert.NotEqual(t, uuid.Nil, mr.UserBID, "user_b_id should not be nil")
		assert.NotEqual(t, mr.UserAID, mr.UserBID, "user_a_id and user_b_id should be different")
		assert.Equal(t, matchSet.ID, mr.MatchSetID, "match_set_id should match")
	}
}

// calculateTestAge calculates age from birthday to avoid hardcoded values
// that break when the year changes.
func calculateTestAge(birthday time.Time) int {
	now := time.Now()
	age := now.Year() - birthday.Year()
	// Adjust if birthday hasn't occurred yet this year
	if now.YearDay() < birthday.YearDay() {
		age--
	}
	return age
}

// TestPopulate_IsTestUserOption verifies that PopulateOptions.IsTestUser
// correctly marks all populated users as test users in the database.
func TestPopulate_IsTestUserOption(t *testing.T) {
	testSuite := testsuite.New(t)
	t.Cleanup(testSuite.UseBackendDB())
	t.Cleanup(testSuite.UseAiDB())
	t.Cleanup(testSuite.UseSupabaseAuthDB())

	matchLib := testSuite.FakeContainer().GetLibMatching()

	// Parse CSV (standard population file - no is_test_user column)
	file, err := os.Open("./testdata/population_1.csv")
	require.NoError(t, err, "opening test CSV file")
	defer file.Close()

	parseResult, err := matching.ParsePopulationCSV(file)
	require.NoError(t, err, "parsing CSV")
	require.NotEmpty(t, parseResult.Rows, "should have rows")

	ctx := context.Background()

	// Populate WITH IsTestUser option = true
	execs := &matching.PopulationExecutors{
		BackendApp:   testSuite.BackendAppDb(),
		AIBackend:    testSuite.AiBackendDb(),
		SupabaseAuth: testSuite.SupabaseAuthDb(),
	}
	populateResult, err := matchLib.Populate(ctx, execs, parseResult.Rows, &matching.PopulateOptions{
		IsTestUser: true,
	})
	require.NoError(t, err, "populating users")
	require.Empty(t, populateResult.Errors, "expected no populate errors")

	// Verify ALL users are marked as test users in DB
	testUsers, err := pgmodel.Users(
		pgmodel.UserWhere.IsTestUser.EQ(null.BoolFrom(true)),
	).All(ctx, testSuite.BackendAppDb())
	require.NoError(t, err, "fetching test users")
	assert.Len(t, testUsers, parseResult.ValidRows, "all users should be marked as test users")

	// Verify NO users are marked as non-test users
	regularUsers, err := pgmodel.Users(
		pgmodel.UserWhere.IsTestUser.EQ(null.BoolFrom(false)),
	).All(ctx, testSuite.BackendAppDb())
	require.NoError(t, err, "fetching regular users")
	assert.Len(t, regularUsers, 0, "no users should be marked as regular users")
}

// TestPopulate_DefaultIsNotTestUser verifies that by default (nil options),
// users are NOT marked as test users.
func TestPopulate_DefaultIsNotTestUser(t *testing.T) {
	testSuite := testsuite.New(t)
	t.Cleanup(testSuite.UseBackendDB())
	t.Cleanup(testSuite.UseAiDB())
	t.Cleanup(testSuite.UseSupabaseAuthDB())

	matchLib := testSuite.FakeContainer().GetLibMatching()

	// Ingest population data with default options (nil)
	ingestor := testhelper.NewPopulationIngestor(t, testSuite, matchLib)
	parseResult, populateResult, err := ingestor.IngestFromCSVFile("./testdata/population_1.csv")
	require.NoError(t, err, "ingesting population data")
	require.Empty(t, populateResult.Errors, "expected no populate errors")

	ctx := context.Background()

	// Verify ALL users are marked as NOT test users (default)
	regularUsers, err := pgmodel.Users(
		pgmodel.UserWhere.IsTestUser.EQ(null.BoolFrom(false)),
	).All(ctx, testSuite.BackendAppDb())
	require.NoError(t, err, "fetching regular users")
	assert.Len(t, regularUsers, parseResult.ValidRows, "all users should be regular by default")

	// Verify NO users are marked as test users
	testUsers, err := pgmodel.Users(
		pgmodel.UserWhere.IsTestUser.EQ(null.BoolFrom(true)),
	).All(ctx, testSuite.BackendAppDb())
	require.NoError(t, err, "fetching test users")
	assert.Len(t, testUsers, 0, "no users should be test users by default")
}

// TestIngestion_OnlyTestUsers verifies that batch matching can filter to
// include ONLY test users when is_test_user=true option is set.
func TestIngestion_OnlyTestUsers(t *testing.T) {
	testSuite := testsuite.New(t)
	t.Cleanup(testSuite.UseBackendDB())
	t.Cleanup(testSuite.UseAiDB())
	t.Cleanup(testSuite.UseSupabaseAuthDB())

	matchLib := testSuite.FakeContainer().GetLibMatching()
	ctx := context.Background()

	// Parse CSV
	file, err := os.Open("./testdata/population_1.csv")
	require.NoError(t, err, "opening test CSV file")
	defer file.Close()

	parseResult, err := matching.ParsePopulationCSV(file)
	require.NoError(t, err, "parsing CSV")

	execs := &matching.PopulationExecutors{
		BackendApp:   testSuite.BackendAppDb(),
		AIBackend:    testSuite.AiBackendDb(),
		SupabaseAuth: testSuite.SupabaseAuthDb(),
	}

	// Populate first 3 users as regular users
	regularRows := parseResult.Rows[:3]
	_, err = matchLib.Populate(ctx, execs, regularRows, nil)
	require.NoError(t, err, "populating regular users")

	// Populate remaining 3 users as test users
	testRows := parseResult.Rows[3:]
	_, err = matchLib.Populate(ctx, execs, testRows, &matching.PopulateOptions{IsTestUser: true})
	require.NoError(t, err, "populating test users")

	// Run batch matching for ONLY test users (is_test_user=true)
	isTestUser := true
	batchOptions := &matching.BatchIngestOptions{
		IsTestUser: &isTestUser,
	}
	matchSet, err := matchLib.IngestWithOptions(ctx, testSuite.BackendAppDb(), batchOptions)
	require.NoError(t, err, "ingesting only test users")
	require.NotNil(t, matchSet, "match set should not be nil")

	// Should only have 3 test users
	assert.Equal(t, 3, matchSet.NumberOfParticipants, "should only include 3 test users")

	// Verify match results count: 3 users = 3*(3-1)/2 = 3 pairs
	ms := store.NewMatchingStores(applog.NewLogrus("test"))
	paginatedResults, err := ms.MatchResultStore.MatchResults(ctx,
		testSuite.BackendAppDb(),
		&matching.QueryFilterMatchResult{
			MatchSetID: null.StringFrom(matchSet.ID.String()),
		},
	)
	require.NoError(t, err, "fetching match results")
	assert.Len(t, paginatedResults.Data, 3, "should have 3 match pairs for 3 test users")
}

// TestIngestion_OnlyNonTestUsers verifies that batch matching can filter to
// include ONLY non-test users when is_test_user=false option is set.
func TestIngestion_OnlyNonTestUsers(t *testing.T) {
	testSuite := testsuite.New(t)
	t.Cleanup(testSuite.UseBackendDB())
	t.Cleanup(testSuite.UseAiDB())
	t.Cleanup(testSuite.UseSupabaseAuthDB())

	matchLib := testSuite.FakeContainer().GetLibMatching()
	ctx := context.Background()

	// Parse CSV
	file, err := os.Open("./testdata/population_1.csv")
	require.NoError(t, err, "opening test CSV file")
	defer file.Close()

	parseResult, err := matching.ParsePopulationCSV(file)
	require.NoError(t, err, "parsing CSV")

	execs := &matching.PopulationExecutors{
		BackendApp:   testSuite.BackendAppDb(),
		AIBackend:    testSuite.AiBackendDb(),
		SupabaseAuth: testSuite.SupabaseAuthDb(),
	}

	// Populate first 3 users as regular users
	regularRows := parseResult.Rows[:3]
	_, err = matchLib.Populate(ctx, execs, regularRows, nil)
	require.NoError(t, err, "populating regular users")

	// Populate remaining 3 users as test users
	testRows := parseResult.Rows[3:]
	_, err = matchLib.Populate(ctx, execs, testRows, &matching.PopulateOptions{IsTestUser: true})
	require.NoError(t, err, "populating test users")

	// Run batch matching for ONLY non-test users (is_test_user=false)
	isTestUser := false
	batchOptions := &matching.BatchIngestOptions{
		IsTestUser: &isTestUser,
	}
	matchSet, err := matchLib.IngestWithOptions(ctx, testSuite.BackendAppDb(), batchOptions)
	require.NoError(t, err, "ingesting only non-test users")
	require.NotNil(t, matchSet, "match set should not be nil")

	// Should only have 3 regular users
	assert.Equal(t, 3, matchSet.NumberOfParticipants, "should only include 3 regular users")

	// Verify match results count: 3 users = 3*(3-1)/2 = 3 pairs
	ms := store.NewMatchingStores(applog.NewLogrus("test"))
	paginatedResults, err := ms.MatchResultStore.MatchResults(ctx,
		testSuite.BackendAppDb(),
		&matching.QueryFilterMatchResult{
			MatchSetID: null.StringFrom(matchSet.ID.String()),
		},
	)
	require.NoError(t, err, "fetching match results")
	assert.Len(t, paginatedResults.Data, 3, "should have 3 match pairs for 3 regular users")
}
