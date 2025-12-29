package repo_test

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	aiBackendFactory "wingedapp/pgtester/internal/wingedapp/aibackend/db/factory"
	"wingedapp/pgtester/internal/wingedapp/business/domain/registration"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/stretchr/testify/require"

	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/repo"
	"wingedapp/pgtester/internal/wingedapp/testsuite"
)

type testCaseAnonymizedContacts struct {
	name           string
	filter         *registration.AudioFileQueryFilter
	mutations      func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseAnonymizedContacts)
	assertions     func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseAnonymizedContacts, audioFile aipgmodel.AudioFileSlice, err error)
	expectedID     string
	expectedUserID string
	expectedIDs    []string
}

func getTestCasesAudioFile() []testCaseAnonymizedContacts {
	return []testCaseAnonymizedContacts{
		{
			name:   "success-filter-by-user-id",
			filter: &registration.AudioFileQueryFilter{},
			mutations: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseAnonymizedContacts) {
				fAudioFile := factory.Entity[*aiBackendFactory.AudioFile]{}
				audioFile := fAudioFile.New(th.T, db)
				tc.filter.UserID = null.StringFrom(audioFile.Subject.UserID)
				tc.expectedUserID = audioFile.Subject.UserID
			},
			assertions: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseAnonymizedContacts, audioFile aipgmodel.AudioFileSlice, err error) {
				require.NoError(th.T, err)
				require.Len(th.T, audioFile, 1)
				require.Equal(th.T, tc.expectedUserID, audioFile[0].UserID)
			},
		},
		{
			name:   "success-filter-by-id",
			filter: &registration.AudioFileQueryFilter{},
			mutations: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseAnonymizedContacts) {
				fAudioFile := factory.Entity[*aiBackendFactory.AudioFile]{}
				audioFile := fAudioFile.New(th.T, db)

				tc.filter.ID = audioFile.Subject.ID

				tc.expectedID = audioFile.Subject.ID
			},
			assertions: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseAnonymizedContacts, audioFile aipgmodel.AudioFileSlice, err error) {
				require.NoError(th.T, err)
				require.Len(th.T, audioFile, 1)
				require.Equal(th.T, tc.expectedID, audioFile[0].ID)
			},
		},
		{
			name:   "success-multiple-audio-files",
			filter: &registration.AudioFileQueryFilter{},
			mutations: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseAnonymizedContacts) {
				fAudioFile := factory.Entity[*aiBackendFactory.AudioFile]{}
				audioFile1 := fAudioFile.New(th.T, db)

				fAudioFile2 := factory.Entity[*aiBackendFactory.AudioFile]{}
				audioFile2 := fAudioFile2.New(th.T, db)

				tc.expectedIDs = []string{audioFile1.Subject.ID, audioFile2.Subject.ID}
			},
			assertions: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseAnonymizedContacts, audioFile aipgmodel.AudioFileSlice, err error) {
				require.NoError(th.T, err)
				require.Len(th.T, audioFile, 2)

				foundIDs := []string{audioFile[0].ID, audioFile[1].ID}

				require.ElementsMatch(th.T, tc.expectedIDs, foundIDs)
			},
		},
	}
}

func TestRepo_AudioFile(t *testing.T) {
	for _, tc := range getTestCasesAudioFile() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tSuite := testsuite.New(t)
			db, cleanup := tSuite.AIDb()
			defer cleanup()

			tc.mutations(tSuite, db, &tc)

			store := repo.Store{}
			audioFile, err := store.AudioFiles(context.Background(), db, tc.filter)
			tc.assertions(tSuite, db, &tc, audioFile, err)
		})
	}
}
