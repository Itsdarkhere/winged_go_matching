package matching_test

import (
	"context"
	"sync"
	"testing"
	"time"
	"wingedapp/pgtester/internal/db/factory"
	wingedFactory "wingedapp/pgtester/internal/wingedapp/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/lib/enums"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProposeMatch_ConcurrentProposals_RaceCondition tests the fix for concurrent proposals.
// This test verifies that when two users propose simultaneously, exactly one detects mutual
// proposal and exactly one date_instance is created (no duplicates, no missed detection).
func TestProposeMatch_ConcurrentProposals_RaceCondition(t *testing.T) {
	t.Parallel()

	testSuite := testsuite.New(t)
	t.Cleanup(testSuite.UseBackendDB())

	ctx := context.Background()
	exec := testSuite.BackendAppDb()

	// Setup: Create users and match
	userA := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{
		Subject: &pgmodel.User{FirstName: null.StringFrom("UserA")},
	}).New(t, exec)

	userB := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{
		Subject: &pgmodel.User{FirstName: null.StringFrom("UserB")},
	}).New(t, exec)

	matchResult := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
		Subject: &pgmodel.MatchResult{
			IsApproved:  true,
			IsDropped:   true,
			UserAAction: string(enums.MatchUserActionPending),
			UserBAction: string(enums.MatchUserActionPending),
		},
		FactoryUserA: userA,
		FactoryUserB: userB,
	}).New(t, exec)

	matchLib := testSuite.FakeContainer().GetLibMatching()

	transactor := testSuite.FakeContainer().GetStoreBackendAppTransactor()

	// Execute: Both users propose simultaneously
	var wg sync.WaitGroup
	var resultA, resultB *matching.ProposeMatchResult
	var errA, errB error

	wg.Add(2)

	// User A proposes
	go func() {
		defer wg.Done()

		// Get separate transaction for user A
		tx, err := transactor.TX()
		if err != nil {
			errA = err
			return
		}
		defer transactor.Rollback(tx)

		resultA, errA = matchLib.ProposeMatch(ctx, tx, &matching.ProposeMatchParams{
			MatchResultID: uuid.MustParse(matchResult.Subject.ID),
			UserID:        uuid.MustParse(userA.Subject.ID),
		})

		if errA == nil {
			errA = tx.Commit()
		}
	}()

	// User B proposes (slight delay to increase race probability)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond) // Stagger slightly

		tx, err := transactor.TX()
		if err != nil {
			errB = err
			return
		}
		defer transactor.Rollback(tx)

		resultB, errB = matchLib.ProposeMatch(ctx, tx, &matching.ProposeMatchParams{
			MatchResultID: uuid.MustParse(matchResult.Subject.ID),
			UserID:        uuid.MustParse(userB.Subject.ID),
		})

		if errB == nil {
			errB = tx.Commit()
		}
	}()

	wg.Wait()

	// Assert: Both should succeed
	require.NoError(t, errA, "User A propose should succeed")
	require.NoError(t, errB, "User B propose should succeed")
	require.NotNil(t, resultA)
	require.NotNil(t, resultB)

	// Critical assertion: EXACTLY ONE should detect mutual proposal
	// (The second one to acquire the lock will see partner already proposed)
	mutualCount := 0
	if resultA.MutualProposal {
		mutualCount++
	}
	if resultB.MutualProposal {
		mutualCount++
	}

	assert.Equal(t, 1, mutualCount, "Exactly one proposal should detect mutual status")

	// Verify database state
	dateInstanceCount, err := pgmodel.DateInstances(
		pgmodel.DateInstanceWhere.MatchResultRefID.EQ(matchResult.Subject.ID),
	).Count(ctx, exec)
	require.NoError(t, err)
	assert.Equal(t, int64(1), dateInstanceCount, "Exactly one date_instance should be created")

	// Verify match_result was updated correctly
	finalMatch, err := pgmodel.FindMatchResult(ctx, exec, matchResult.Subject.ID)
	require.NoError(t, err)
	assert.True(t, finalMatch.CurrentDateInstanceID.Valid, "Match should have date_instance_id")
	assert.True(t, finalMatch.MatchLifecycleStatus.Valid, "Match should have lifecycle status")
}

// TestProposeMatch_SequentialProposals_BothDetect tests normal flow (non-concurrent).
// This verifies the fix doesn't break the normal sequential proposal case.
func TestProposeMatch_SequentialProposals_BothDetect(t *testing.T) {
	t.Parallel()

	testSuite := testsuite.New(t)
	t.Cleanup(testSuite.UseBackendDB())

	ctx := context.Background()
	exec := testSuite.BackendAppDb()

	// Setup (same as concurrent test)
	userA := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{
		Subject: &pgmodel.User{FirstName: null.StringFrom("UserA")},
	}).New(t, exec)

	userB := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{
		Subject: &pgmodel.User{FirstName: null.StringFrom("UserB")},
	}).New(t, exec)

	matchResult := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
		Subject: &pgmodel.MatchResult{
			IsApproved:  true,
			IsDropped:   true,
			UserAAction: string(enums.MatchUserActionPending),
			UserBAction: string(enums.MatchUserActionPending),
		},
		FactoryUserA: userA,
		FactoryUserB: userB,
	}).New(t, exec)

	matchLib := testSuite.FakeContainer().GetLibMatching()
	transactor := testSuite.FakeContainer().GetStoreBackendAppTransactor()

	// Execute: User A proposes first
	tx1, err := transactor.TX()
	require.NoError(t, err)
	defer transactor.Rollback(tx1)

	resultA, err := matchLib.ProposeMatch(ctx, tx1, &matching.ProposeMatchParams{
		MatchResultID: uuid.MustParse(matchResult.Subject.ID),
		UserID:        uuid.MustParse(userA.Subject.ID),
	})
	require.NoError(t, err)
	require.NoError(t, tx1.Commit())

	// Assert: First proposal should NOT be mutual
	assert.False(t, resultA.MutualProposal, "First proposal should not be mutual")
	assert.Empty(t, resultA.DateInstanceID, "No date instance yet")

	// Execute: User B proposes second
	tx2, err := transactor.TX()
	require.NoError(t, err)
	defer transactor.Rollback(tx2)

	resultB, err := matchLib.ProposeMatch(ctx, tx2, &matching.ProposeMatchParams{
		MatchResultID: uuid.MustParse(matchResult.Subject.ID),
		UserID:        uuid.MustParse(userB.Subject.ID),
	})
	require.NoError(t, err)
	require.NoError(t, tx2.Commit())

	// Assert: Second proposal SHOULD be mutual
	assert.True(t, resultB.MutualProposal, "Second proposal should be mutual")
	assert.NotEmpty(t, resultB.DateInstanceID, "Date instance should be created")

	// Verify database state
	dateInstanceCount, err := pgmodel.DateInstances(
		pgmodel.DateInstanceWhere.MatchResultRefID.EQ(matchResult.Subject.ID),
	).Count(ctx, exec)
	require.NoError(t, err)
	assert.Equal(t, int64(1), dateInstanceCount, "Exactly one date_instance created")
}

// TestProposeMatch_Idempotency tests that calling ProposeMatch multiple times
// on the same mutual proposal doesn't create duplicate date_instances.
func TestProposeMatch_Idempotency(t *testing.T) {
	t.Parallel()

	testSuite := testsuite.New(t)
	t.Cleanup(testSuite.UseBackendDB())

	ctx := context.Background()
	exec := testSuite.BackendAppDb()

	// Setup
	userA := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{
		Subject: &pgmodel.User{FirstName: null.StringFrom("UserA")},
	}).New(t, exec)

	userB := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{
		Subject: &pgmodel.User{FirstName: null.StringFrom("UserB")},
	}).New(t, exec)

	matchResult := factory.NewEntity[*wingedFactory.MatchResult](&wingedFactory.MatchResult{
		Subject: &pgmodel.MatchResult{
			IsApproved:  true,
			IsDropped:   true,
			UserAAction: string(enums.MatchUserActionPending),
			UserBAction: string(enums.MatchUserActionPending),
		},
		FactoryUserA: userA,
		FactoryUserB: userB,
	}).New(t, exec)

	matchLib := testSuite.FakeContainer().GetLibMatching()
	transactor := testSuite.FakeContainer().GetStoreBackendAppTransactor()

	// Execute: Both users propose sequentially
	tx1, err := transactor.TX()
	require.NoError(t, err)
	defer transactor.Rollback(tx1)

	_, err = matchLib.ProposeMatch(ctx, tx1, &matching.ProposeMatchParams{
		MatchResultID: uuid.MustParse(matchResult.Subject.ID),
		UserID:        uuid.MustParse(userA.Subject.ID),
	})
	require.NoError(t, err)
	require.NoError(t, tx1.Commit())

	tx2, err := transactor.TX()
	require.NoError(t, err)
	defer transactor.Rollback(tx2)

	result2, err := matchLib.ProposeMatch(ctx, tx2, &matching.ProposeMatchParams{
		MatchResultID: uuid.MustParse(matchResult.Subject.ID),
		UserID:        uuid.MustParse(userB.Subject.ID),
	})
	require.NoError(t, err)
	require.NoError(t, tx2.Commit())

	// Get first date_instance ID
	firstDateInstanceID := result2.DateInstanceID
	require.NotEmpty(t, firstDateInstanceID, "Date instance should be created")

	// Execute: User B proposes AGAIN (simulating duplicate action)
	tx3, err := transactor.TX()
	require.NoError(t, err)
	defer transactor.Rollback(tx3)

	result3, err := matchLib.ProposeMatch(ctx, tx3, &matching.ProposeMatchParams{
		MatchResultID: uuid.MustParse(matchResult.Subject.ID),
		UserID:        uuid.MustParse(userB.Subject.ID),
	})
	require.NoError(t, err)
	require.NoError(t, tx3.Commit())

	// Assert: Should still be mutual, same date_instance
	assert.True(t, result3.MutualProposal, "Should still be mutual")
	assert.Equal(t, firstDateInstanceID, result3.DateInstanceID, "Should return same date_instance ID")

	// Verify database state: STILL only one date_instance
	dateInstanceCount, err := pgmodel.DateInstances(
		pgmodel.DateInstanceWhere.MatchResultRefID.EQ(matchResult.Subject.ID),
	).Count(ctx, exec)
	require.NoError(t, err)
	assert.Equal(t, int64(1), dateInstanceCount, "Still exactly one date_instance (idempotent)")
}
