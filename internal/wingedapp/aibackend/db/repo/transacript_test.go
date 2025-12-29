package repo_test

import (
	"context"
	"testing"
	"time"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"
	aiBackendFactory "wingedapp/pgtester/internal/wingedapp/aibackend/db/factory"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/repo"
	"wingedapp/pgtester/internal/wingedapp/business/domain/registration"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/types"
	"github.com/stretchr/testify/require"
)

type testCaseTranscript struct {
	name           string
	filter         *registration.TranscriptQueryFilters
	mutations      func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseTranscript)
	assertions     func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseTranscript, transcripts aipgmodel.TranscriptSlice, err error)
	expectedID     string
	expectedUserID string
	expectedIDs    []string
}

func makeTranscript(id, userID string) *aipgmodel.Transcript {
	return &aipgmodel.Transcript{
		ID:             id,
		UserID:         userID,
		WebhookType:    "test_webhook",
		EventTimestamp: int(time.Now().Unix()),
		RawPayload:     types.JSON([]byte(`{}`)),
	}
}

func testCasesTranscript() []testCaseTranscript {
	return []testCaseTranscript{
		{
			name: "fail-not-found",
			filter: &registration.TranscriptQueryFilters{
				ID:     uuid.New().String(),
				UserID: null.NewString("", false),
			},
			mutations: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseTranscript) {
				// no transcripts inserted
			},
			assertions: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseTranscript, transcripts aipgmodel.TranscriptSlice, err error) {
				require.NoError(th.T, err)
				require.Empty(th.T, transcripts)
			},
		},
		{
			name:   "success-filter-by-user-id",
			filter: &registration.TranscriptQueryFilters{},
			mutations: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseTranscript) {
				fTranscript := factory.Entity[*aiBackendFactory.Transcript]{}
				transcript := fTranscript.New(th.T, db)

				tc.filter.UserID = null.StringFrom(transcript.Subject.UserID)

				tc.expectedUserID = transcript.Subject.UserID
			},
			assertions: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseTranscript, transcripts aipgmodel.TranscriptSlice, err error) {
				require.NoError(th.T, err)
				require.Len(th.T, transcripts, 1)
				require.Equal(th.T, tc.expectedUserID, transcripts[0].UserID)
			},
		},
		{
			name:   "success-filter-by-id",
			filter: &registration.TranscriptQueryFilters{},
			mutations: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseTranscript) {
				fTranscript := factory.Entity[*aiBackendFactory.Transcript]{}
				transcript := fTranscript.New(th.T, db)

				tc.filter.ID = transcript.Subject.ID

				tc.expectedID = transcript.Subject.ID
			},
			assertions: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseTranscript, transcripts aipgmodel.TranscriptSlice, err error) {
				require.NoError(th.T, err)
				require.Len(th.T, transcripts, 1)
				require.Equal(th.T, tc.expectedID, transcripts[0].ID)
			},
		},
		{
			name:   "success-multiple-transcripts",
			filter: &registration.TranscriptQueryFilters{},
			mutations: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseTranscript) {
				fTranscript1 := factory.Entity[*aiBackendFactory.Transcript]{}
				transcript1 := fTranscript1.New(th.T, db)

				fTranscript2 := factory.Entity[*aiBackendFactory.Transcript]{}
				transcript2 := fTranscript2.New(th.T, db)

				tc.expectedIDs = []string{transcript1.Subject.ID, transcript2.Subject.ID}
			},
			assertions: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseTranscript, transcripts aipgmodel.TranscriptSlice, err error) {
				require.NoError(th.T, err)
				require.Len(th.T, transcripts, 2)

				foundIDs := []string{transcripts[0].ID, transcripts[1].ID}
				require.ElementsMatch(th.T, tc.expectedIDs, foundIDs)
			},
		},
		{
			name: "success-multiple-entries-ordered-by-created-at-desc-limit-1",
			filter: &registration.TranscriptQueryFilters{
				Limit:      null.IntFrom(1),
				OrderedBys: []string{"created_at DESC"},
			},
			mutations: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseTranscript) {
				fTranscript1 := factory.Entity[*aiBackendFactory.Transcript]{}
				fTranscript1.New(th.T, db)

				fTranscript2 := factory.Entity[*aiBackendFactory.Transcript]{}
				fTranscript2.New(th.T, db)

				fTranscript3 := factory.Entity[*aiBackendFactory.Transcript]{}
				transcript3 := fTranscript3.New(th.T, db)

				tc.expectedIDs = []string{transcript3.Subject.ID}
			},
			assertions: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseTranscript, transcripts aipgmodel.TranscriptSlice, err error) {
				require.NoError(th.T, err)
				require.Len(th.T, transcripts, 1)

				foundIDs := []string{transcripts[0].ID}
				require.ElementsMatch(th.T, tc.expectedIDs, foundIDs)
			},
		},
		{
			name: "success-multiple-entries-ordered-by-created-at-asc-limit-1",
			filter: &registration.TranscriptQueryFilters{
				Limit:      null.IntFrom(1),
				OrderedBys: []string{"created_at ASC"},
			},
			mutations: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseTranscript) {
				fTranscript1 := factory.Entity[*aiBackendFactory.Transcript]{}
				transcript1 := fTranscript1.New(th.T, db)

				fTranscript2 := factory.Entity[*aiBackendFactory.Transcript]{}
				fTranscript2.New(th.T, db)

				fTranscript3 := factory.Entity[*aiBackendFactory.Transcript]{}
				fTranscript3.New(th.T, db)

				tc.expectedIDs = []string{transcript1.Subject.ID}
			},
			assertions: func(th *testsuite.Helper, db boil.ContextExecutor, tc *testCaseTranscript, transcripts aipgmodel.TranscriptSlice, err error) {
				require.NoError(th.T, err)
				require.Len(th.T, transcripts, 1)

				foundIDs := []string{transcripts[0].ID}
				require.ElementsMatch(th.T, tc.expectedIDs, foundIDs)
			},
		},
	}
}

func TestRepo_TranscriptList(t *testing.T) {
	for _, tc := range testCasesTranscript() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tSuite := testsuite.New(t)
			db, cleanup := tSuite.AIDb()
			defer cleanup()

			store := repo.Store{}
			tc.mutations(tSuite, db, &tc)

			transcripts, err := store.Transcripts(context.Background(), db, tc.filter)
			tc.assertions(tSuite, db, &tc, transcripts, err)
		})
	}
}

func TestRepo_TranscriptDetails(t *testing.T) {
	tSuite := testsuite.New(t)
	db, cleanup := tSuite.AIDb()
	defer cleanup()

	store := repo.Store{}

	transcriptID := uuid.New().String()
	userID := uuid.New().String()

	p := makeTranscript(transcriptID, userID)
	err := p.Insert(context.Background(), db, boil.Infer())
	require.NoError(t, err)

	filter := &registration.TranscriptQueryFilters{ID: transcriptID}
	transcript, err := store.Transcript(context.Background(), db, filter)
	require.NoError(t, err)
	require.NotNil(t, transcript)
	require.Equal(t, transcriptID, transcript.ID)

	filter = &registration.TranscriptQueryFilters{ID: uuid.New().String()}
	transcript, err = store.Transcript(context.Background(), db, filter)
	require.NoError(t, err)
	require.Nil(t, transcript)

	// transcript
}
