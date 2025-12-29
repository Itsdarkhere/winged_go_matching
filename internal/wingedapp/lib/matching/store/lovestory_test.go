package store_test

import (
	"context"
	"testing"
	"wingedapp/pgtester/internal/db/factory"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"
	aiFactory "wingedapp/pgtester/internal/wingedapp/aibackend/db/factory"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/matching/store"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newLovestoryStore() *store.LovestoryStore {
	return store.NewLovestoryStore(applog.NewLogrus("test"))
}

// ============ LovestoryURLsByMatchIDs Tests ============

type testCaseLovestoryURLs struct {
	name            string
	setup           func(th *testsuite.Helper) []uuid.UUID
	extraAssertions func(th *testsuite.Helper, result map[uuid.UUID]null.String, err error)
}

func lovestoryURLsTestCases() []testCaseLovestoryURLs {
	return []testCaseLovestoryURLs{
		{
			name: "success-returns-lovestory-url-for-match",
			setup: func(th *testsuite.Helper) []uuid.UUID {
				exec := th.AiBackendDb()
				matchID := uuid.New()

				// Create lovestory record
				factory.NewEntity[*aiFactory.Lovestory](&aiFactory.Lovestory{
					Subject: &aipgmodel.Lovestory{
						MatchResultRefID:         matchID.String(),
						FirstDateSimulationAudio: null.StringFrom("https://example.com/lovestory-audio.mp3"),
					},
				}).New(th.T, exec)

				return []uuid.UUID{matchID}
			},
			extraAssertions: func(th *testsuite.Helper, result map[uuid.UUID]null.String, err error) {
				require.NoError(th.T, err, "should not error")
				require.Len(th.T, result, 1, "should return 1 result")

				for _, url := range result {
					assert.True(th.T, url.Valid, "URL should be valid")
					assert.Equal(th.T, "https://example.com/lovestory-audio.mp3", url.String)
				}
			},
		},
		{
			name: "success-returns-empty-for-no-matches",
			setup: func(th *testsuite.Helper) []uuid.UUID {
				return []uuid.UUID{uuid.New(), uuid.New()}
			},
			extraAssertions: func(th *testsuite.Helper, result map[uuid.UUID]null.String, err error) {
				require.NoError(th.T, err, "should not error")
				assert.Empty(th.T, result, "should return empty map for non-existent matches")
			},
		},
		{
			name: "success-returns-empty-for-empty-input",
			setup: func(th *testsuite.Helper) []uuid.UUID {
				return []uuid.UUID{}
			},
			extraAssertions: func(th *testsuite.Helper, result map[uuid.UUID]null.String, err error) {
				require.NoError(th.T, err, "should not error")
				assert.Empty(th.T, result, "should return empty map for empty input")
			},
		},
		{
			name: "success-skips-null-audio-urls",
			setup: func(th *testsuite.Helper) []uuid.UUID {
				exec := th.AiBackendDb()
				matchID := uuid.New()

				// Create lovestory record with NULL audio
				factory.NewEntity[*aiFactory.Lovestory](&aiFactory.Lovestory{
					Subject: &aipgmodel.Lovestory{
						MatchResultRefID:         matchID.String(),
						FirstDateSimulationAudio: null.String{}, // NULL
					},
				}).New(th.T, exec)

				return []uuid.UUID{matchID}
			},
			extraAssertions: func(th *testsuite.Helper, result map[uuid.UUID]null.String, err error) {
				require.NoError(th.T, err, "should not error")
				assert.Empty(th.T, result, "should skip records with NULL audio")
			},
		},
		{
			name: "success-batch-returns-multiple-urls",
			setup: func(th *testsuite.Helper) []uuid.UUID {
				exec := th.AiBackendDb()
				matchID1 := uuid.New()
				matchID2 := uuid.New()
				matchID3 := uuid.New() // No lovestory for this one

				factory.NewEntity[*aiFactory.Lovestory](&aiFactory.Lovestory{
					Subject: &aipgmodel.Lovestory{
						MatchResultRefID:         matchID1.String(),
						FirstDateSimulationAudio: null.StringFrom("https://example.com/audio1.mp3"),
					},
				}).New(th.T, exec)

				factory.NewEntity[*aiFactory.Lovestory](&aiFactory.Lovestory{
					Subject: &aipgmodel.Lovestory{
						MatchResultRefID:         matchID2.String(),
						FirstDateSimulationAudio: null.StringFrom("https://example.com/audio2.mp3"),
					},
				}).New(th.T, exec)

				return []uuid.UUID{matchID1, matchID2, matchID3}
			},
			extraAssertions: func(th *testsuite.Helper, result map[uuid.UUID]null.String, err error) {
				require.NoError(th.T, err, "should not error")
				assert.Len(th.T, result, 2, "should return 2 results (matchID3 has no lovestory)")
			},
		},
	}
}

func TestLovestoryStore_LovestoryURLsByMatchIDs(t *testing.T) {
	for _, tt := range lovestoryURLsTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tSuite := testsuite.New(t)
			t.Cleanup(tSuite.UseAiDB())

			matchIDs := tt.setup(tSuite)

			stor := newLovestoryStore()
			result, err := stor.LovestoryURLsByMatchIDs(context.Background(), tSuite.AiBackendDb(), matchIDs)

			tt.extraAssertions(tSuite, result, err)
		})
	}
}
