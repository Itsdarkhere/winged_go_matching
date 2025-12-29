package repo_test

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
)

type testCaseUserElevenLabs struct {
	name       string
	filter     *repo.UserElevenLabsQueryFilter
	cleanupFn  []func()
	mutations  func(th *testsuite.Helper, tc *testCaseUserElevenLabs)
	assertions func(th *testsuite.Helper, tc *testCaseUserElevenLabs, userElevenLabs pgmodel.UserElevenLabSlice, err error)
}

func getTestCasesUserElevenLabs() []testCaseUserElevenLabs {
	return []testCaseUserElevenLabs{
		{
			name: "fail (lets see what you come up here)",
			filter: &repo.UserElevenLabsQueryFilter{
				ID:     null.String{},
				UserID: null.String{},
			},
			mutations: func(th *testsuite.Helper, tc *testCaseUserElevenLabs) {
				// do mutations - factories to spawn data
			},
			assertions: func(th *testsuite.Helper, tc *testCaseUserElevenLabs, userElevenLabs pgmodel.UserElevenLabSlice, err error) {
				// do assertions - of things you want to expect
			},
		},
		{
			name: "success",
			filter: &repo.UserElevenLabsQueryFilter{
				ID:     null.String{},
				UserID: null.String{},
			},
			mutations: func(th *testsuite.Helper, tc *testCaseUserElevenLabs) {
				// do mutations - factories to spawn data
			},
			assertions: func(th *testsuite.Helper, tc *testCaseUserElevenLabs, userElevenLabs pgmodel.UserElevenLabSlice, err error) {
				// do assertions - of things you want to expect
			},
		},
	}
}

func TestRepo_UserElevenLabs(t *testing.T) {
	for _, tc := range getTestCasesUserElevenLabs() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tSuite := testsuite.New(t)
			db, cleanup := tSuite.BackendAppDB()
			defer cleanup()

			store := repo.Store{}
			userElevenLabs, err := store.UserElevenLabs(context.Background(), db, tc.filter)
			tc.assertions(tSuite, &tc, userElevenLabs, err)
		})
	}
}
