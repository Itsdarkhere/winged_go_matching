package repo_test

import (
	"context"
	"testing"

	"wingedapp/pgtester/internal/db/factory"
	wingedFactory "wingedapp/pgtester/internal/wingedapp/db/factory"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/enums"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newStore() *repo.Store {
	return &repo.Store{}
}

// --- SyncDietaryRestrictions tests ---

func TestStore_SyncDietaryRestrictions_InsertNew(t *testing.T) {
	tSuite := testsuite.New(t)
	t.Cleanup(tSuite.UseBackendDB())

	ctx := context.Background()
	db := tSuite.BackendAppDb()
	stor := newStore()

	// Create user
	user := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(t, db)

	// Sync with 2 dietary restrictions
	result, err := stor.SyncDietaryRestrictions(ctx, db, &repo.SyncUserPreference{
		UserID: user.Subject.ID,
		Values: []string{string(enums.DietaryRestrictionVegan), string(enums.DietaryRestrictionVegetarian)},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(0), result.Deleted)
	assert.Equal(t, int64(2), result.Inserted)

	// Verify in DB
	dbRestrictions, err := pgmodel.UserDietaryRestrictions(
		pgmodel.UserDietaryRestrictionWhere.UserID.EQ(user.Subject.ID),
	).All(ctx, db)
	require.NoError(t, err)
	assert.Len(t, dbRestrictions, 2)
}

func TestStore_SyncDietaryRestrictions_ReplaceExisting(t *testing.T) {
	tSuite := testsuite.New(t)
	t.Cleanup(tSuite.UseBackendDB())

	ctx := context.Background()
	db := tSuite.BackendAppDb()
	stor := newStore()

	// Create user
	user := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(t, db)

	// Insert initial restriction
	factory.NewEntity[*wingedFactory.UserDietaryRestriction](&wingedFactory.UserDietaryRestriction{
		Subject: &pgmodel.UserDietaryRestriction{
			UserID:             user.Subject.ID,
			DietaryRestriction: string(enums.DietaryRestrictionVegan),
		},
	}).New(t, db)

	// Sync with different restrictions (should delete old, insert new)
	result, err := stor.SyncDietaryRestrictions(ctx, db, &repo.SyncUserPreference{
		UserID: user.Subject.ID,
		Values: []string{string(enums.DietaryRestrictionGlutenFree), string(enums.DietaryRestrictionHalal)},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Deleted)
	assert.Equal(t, int64(2), result.Inserted)

	// Verify old is gone, new are present
	dbRestrictions, err := pgmodel.UserDietaryRestrictions(
		pgmodel.UserDietaryRestrictionWhere.UserID.EQ(user.Subject.ID),
	).All(ctx, db)
	require.NoError(t, err)
	assert.Len(t, dbRestrictions, 2)

	restrictions := make([]string, len(dbRestrictions))
	for i, r := range dbRestrictions {
		restrictions[i] = r.DietaryRestriction
	}
	assert.Contains(t, restrictions, string(enums.DietaryRestrictionGlutenFree))
	assert.Contains(t, restrictions, string(enums.DietaryRestrictionHalal))
	assert.NotContains(t, restrictions, string(enums.DietaryRestrictionVegan))
}

func TestStore_SyncDietaryRestrictions_ClearAll(t *testing.T) {
	tSuite := testsuite.New(t)
	t.Cleanup(tSuite.UseBackendDB())

	ctx := context.Background()
	db := tSuite.BackendAppDb()
	stor := newStore()

	// Create user
	user := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(t, db)

	// Insert initial restriction
	factory.NewEntity[*wingedFactory.UserDietaryRestriction](&wingedFactory.UserDietaryRestriction{
		Subject: &pgmodel.UserDietaryRestriction{
			UserID:             user.Subject.ID,
			DietaryRestriction: string(enums.DietaryRestrictionVegetarian),
		},
	}).New(t, db)

	// Sync with empty list (should clear all)
	result, err := stor.SyncDietaryRestrictions(ctx, db, &repo.SyncUserPreference{
		UserID: user.Subject.ID,
		Values: []string{},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Deleted)
	assert.Equal(t, int64(0), result.Inserted)

	// Verify empty
	dbRestrictions, err := pgmodel.UserDietaryRestrictions(
		pgmodel.UserDietaryRestrictionWhere.UserID.EQ(user.Subject.ID),
	).All(ctx, db)
	require.NoError(t, err)
	assert.Len(t, dbRestrictions, 0)
}

func TestStore_SyncDietaryRestrictions_ErrorEmptyUserID(t *testing.T) {
	tSuite := testsuite.New(t)
	t.Cleanup(tSuite.UseBackendDB())

	ctx := context.Background()
	db := tSuite.BackendAppDb()
	stor := newStore()

	_, err := stor.SyncDietaryRestrictions(ctx, db, &repo.SyncUserPreference{
		UserID: "",
		Values: []string{string(enums.DietaryRestrictionVegan)},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_id is required")
}

// --- UserDietaryRestrictions query tests ---

func TestStore_UserDietaryRestrictions_WithData(t *testing.T) {
	tSuite := testsuite.New(t)
	t.Cleanup(tSuite.UseBackendDB())

	ctx := context.Background()
	db := tSuite.BackendAppDb()
	stor := newStore()

	// Create user and restriction
	user := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(t, db)

	factory.NewEntity[*wingedFactory.UserDietaryRestriction](&wingedFactory.UserDietaryRestriction{
		Subject: &pgmodel.UserDietaryRestriction{
			UserID:             user.Subject.ID,
			DietaryRestriction: string(enums.DietaryRestrictionVegetarian),
		},
	}).New(t, db)

	// Query
	values, err := stor.UserDietaryRestrictions(ctx, db, user.Subject.ID)
	require.NoError(t, err)
	require.Len(t, values, 1)
	assert.Equal(t, string(enums.DietaryRestrictionVegetarian), values[0])
}

func TestStore_UserDietaryRestrictions_Empty(t *testing.T) {
	tSuite := testsuite.New(t)
	t.Cleanup(tSuite.UseBackendDB())

	ctx := context.Background()
	db := tSuite.BackendAppDb()
	stor := newStore()

	// Create user with no restrictions
	user := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(t, db)

	// Query
	values, err := stor.UserDietaryRestrictions(ctx, db, user.Subject.ID)
	require.NoError(t, err)
	assert.Len(t, values, 0)
}

// --- SyncDateTypePreferences tests ---

func TestStore_SyncDateTypePreferences_InsertNew(t *testing.T) {
	tSuite := testsuite.New(t)
	t.Cleanup(tSuite.UseBackendDB())

	ctx := context.Background()
	db := tSuite.BackendAppDb()
	stor := newStore()

	user := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(t, db)

	result, err := stor.SyncDateTypePreferences(ctx, db, &repo.SyncUserPreference{
		UserID: user.Subject.ID,
		Values: []string{string(enums.DateTypeCoreCoffee)},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(0), result.Deleted)
	assert.Equal(t, int64(1), result.Inserted)

	dbPrefs, err := pgmodel.UserDateTypePreferences(
		pgmodel.UserDateTypePreferenceWhere.UserID.EQ(user.Subject.ID),
	).All(ctx, db)
	require.NoError(t, err)
	assert.Len(t, dbPrefs, 1)
}

// --- SyncMobilityConstraints tests ---

func TestStore_SyncMobilityConstraints_InsertNew(t *testing.T) {
	tSuite := testsuite.New(t)
	t.Cleanup(tSuite.UseBackendDB())

	ctx := context.Background()
	db := tSuite.BackendAppDb()
	stor := newStore()

	user := factory.NewEntity[*wingedFactory.User](&wingedFactory.User{}).New(t, db)

	result, err := stor.SyncMobilityConstraints(ctx, db, &repo.SyncUserPreference{
		UserID: user.Subject.ID,
		Values: []string{string(enums.MobilityConstraintWheelchairAccessible)},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(0), result.Deleted)
	assert.Equal(t, int64(1), result.Inserted)

	dbConstraints, err := pgmodel.UserMobilityConstraints(
		pgmodel.UserMobilityConstraintWhere.UserID.EQ(user.Subject.ID),
	).All(ctx, db)
	require.NoError(t, err)
	assert.Len(t, dbConstraints, 1)
}
