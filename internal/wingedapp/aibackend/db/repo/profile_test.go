package repo_test

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"
	aiBackendFactory "wingedapp/pgtester/internal/wingedapp/aibackend/db/factory"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/repo"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type testCaseProfile struct {
	name           string
	filter         *repo.ProfileQueryFilter
	mutations      func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseProfile)
	assertions     func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseProfile, profiles aipgmodel.ProfileSlice, err error)
	expectedID     string
	expectedUserID string
	expectedIDs    []string
}

func getTestCasesProfile() []testCaseProfile {
	return []testCaseProfile{
		{
			name: "fail-no-profile-found)",
			filter: &repo.ProfileQueryFilter{
				ID:     uuid.New().String(),
				UserID: null.String{},
			},
			mutations: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseProfile) {
				// no profiles inserted
			},
			assertions: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseProfile, profiles aipgmodel.ProfileSlice, err error) {
				require.NoError(th.T, err)
				require.Empty(th.T, profiles)
			},
		},
		{
			name:   "success-filter-by-user-id",
			filter: &repo.ProfileQueryFilter{},
			mutations: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseProfile) {
				fProfile := factory.Entity[*aiBackendFactory.Profile]{}
				profile := fProfile.New(th.T, db)

				tc.filter.UserID = profile.Subject.UserID
			},
			assertions: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseProfile, profiles aipgmodel.ProfileSlice, err error) {
				require.NoError(th.T, err)
				require.Len(th.T, profiles, 1)
				require.False(th.T, profiles[0].UserID.Valid)
				require.Equal(th.T, tc.expectedUserID, profiles[0].UserID.String)
			},
		},
		{
			name:   "success-filter-by-id",
			filter: &repo.ProfileQueryFilter{},
			mutations: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseProfile) {
				fProfile := factory.Entity[*aiBackendFactory.Profile]{}
				profile := fProfile.New(th.T, db)

				tc.filter.ID = profile.Subject.ID

				tc.expectedID = profile.Subject.ID
			},
			assertions: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseProfile, profiles aipgmodel.ProfileSlice, err error) {
				require.NoError(th.T, err)
				require.Len(th.T, profiles, 1)
				require.Equal(th.T, tc.expectedID, profiles[0].ID)
			},
		},
		{
			name:   "success-multiple-profiles",
			filter: &repo.ProfileQueryFilter{},
			mutations: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseProfile) {
				fProfile := factory.Entity[*aiBackendFactory.Profile]{}
				profile1 := fProfile.New(th.T, db)

				p2Factory := factory.Entity[*aiBackendFactory.Profile]{}
				profile2 := p2Factory.New(th.T, db)

				tc.expectedIDs = []string{profile1.Subject.ID, profile2.Subject.ID}
			},
			assertions: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseProfile, profiles aipgmodel.ProfileSlice, err error) {
				require.NoError(th.T, err)
				require.Len(th.T, profiles, 2)

				foundIDs := []string{profiles[0].ID, profiles[1].ID}

				require.ElementsMatch(th.T, tc.expectedIDs, foundIDs)
			},
		},
	}
}

func TestRepo_ProfileList(t *testing.T) {
	for _, tc := range getTestCasesProfile() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tSuite := testsuite.New(t)
			db, cleanup := tSuite.AIDb()
			defer cleanup()

			store := repo.Store{}
			tc.mutations(tSuite, db, &tc)

			profiles, err := store.Profiles(context.Background(), db, tc.filter)
			tc.assertions(tSuite, db, &tc, profiles, err)
		})
	}
}

func TestRepo_ProfileDetails(t *testing.T) {
	tSuite := testsuite.New(t)
	db, cleanup := tSuite.AIDb()
	defer cleanup()

	store := repo.Store{}

	profileID := uuid.New().String()
	userID := uuid.New().String()

	p := &aipgmodel.Profile{
		ID:     profileID,
		UserID: null.StringFrom(userID),
	}
	err := p.Insert(context.Background(), db, boil.Infer())
	require.NoError(t, err)

	filter := &repo.ProfileQueryFilter{ID: profileID}
	profile, err := store.Profile(context.Background(), db, filter)
	require.NoError(t, err)
	require.NotNil(t, profile)
	require.Equal(t, profileID, profile.ID)

	filter = &repo.ProfileQueryFilter{ID: uuid.New().String()}
	notFoundProfile, err := store.Profile(context.Background(), db, filter)
	require.NoError(t, err, "no error for not found profile")
	require.Nil(t, notFoundProfile, "profile should be nil when not found")
}
