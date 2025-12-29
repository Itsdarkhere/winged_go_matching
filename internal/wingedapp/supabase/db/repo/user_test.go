package repo_test

import (
	"context"
	"encoding/json"
	"testing"

	"wingedapp/pgtester/internal/wingedapp/supabase/db/repo"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCaseInsertUser struct {
	name            string
	setup           func(th *testsuite.Helper) *repo.InsertUserParams
	extraAssertions func(th *testsuite.Helper, userID string, err error)
}

func insertUserTestCases() []testCaseInsertUser {
	return []testCaseInsertUser{
		{
			name: "success-creates-otp-compatible-user-with-all-required-fields",
			setup: func(th *testsuite.Helper) *repo.InsertUserParams {
				return &repo.InsertUserParams{
					Email: "otp-test@winged-test.local",
				}
			},
			extraAssertions: func(th *testsuite.Helper, userID string, err error) {
				require.NoError(th.T, err, "should not error on insert")
				assert.NotEmpty(th.T, userID, "should return user ID")

				ctx := context.Background()
				db := th.SupabaseAuthDb()

				// ========== ASSERT: User record exists ==========
				var userCount int
				err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id = $1", userID).Scan(&userCount)
				require.NoError(th.T, err, "should query users table")
				assert.Equal(th.T, 1, userCount, "should have 1 user record")

				// ========== ASSERT: instance_id (required for Supabase) ==========
				var instanceID string
				err = db.QueryRowContext(ctx, "SELECT instance_id FROM users WHERE id = $1", userID).Scan(&instanceID)
				require.NoError(th.T, err, "should query instance_id")
				assert.Equal(th.T, "00000000-0000-0000-0000-000000000000", instanceID, "instance_id must be default Supabase value")

				// ========== ASSERT: raw_app_meta_data (REQUIRED FOR OTP) ==========
				var rawAppMetaDataBytes []byte
				err = db.QueryRowContext(ctx, "SELECT raw_app_meta_data FROM users WHERE id = $1", userID).Scan(&rawAppMetaDataBytes)
				require.NoError(th.T, err, "should query raw_app_meta_data")
				require.NotNil(th.T, rawAppMetaDataBytes, "raw_app_meta_data must not be NULL")

				var rawAppMetaData map[string]interface{}
				err = json.Unmarshal(rawAppMetaDataBytes, &rawAppMetaData)
				require.NoError(th.T, err, "raw_app_meta_data must be valid JSON")
				assert.Equal(th.T, "email", rawAppMetaData["provider"], "raw_app_meta_data.provider must be 'email'")
				providers, ok := rawAppMetaData["providers"].([]interface{})
				require.True(th.T, ok, "raw_app_meta_data.providers must be an array")
				assert.Contains(th.T, providers, "email", "raw_app_meta_data.providers must contain 'email'")

				// ========== ASSERT: raw_user_meta_data (must be valid JSON) ==========
				var rawUserMetaDataBytes []byte
				err = db.QueryRowContext(ctx, "SELECT raw_user_meta_data FROM users WHERE id = $1", userID).Scan(&rawUserMetaDataBytes)
				require.NoError(th.T, err, "should query raw_user_meta_data")
				require.NotNil(th.T, rawUserMetaDataBytes, "raw_user_meta_data must not be NULL")

				var rawUserMetaData map[string]interface{}
				err = json.Unmarshal(rawUserMetaDataBytes, &rawUserMetaData)
				require.NoError(th.T, err, "raw_user_meta_data must be valid JSON")

				// ========== ASSERT: aud and role ==========
				var aud, role string
				err = db.QueryRowContext(ctx, "SELECT aud, role FROM users WHERE id = $1", userID).Scan(&aud, &role)
				require.NoError(th.T, err, "should query aud and role")
				assert.Equal(th.T, "authenticated", aud, "aud must be 'authenticated'")
				assert.Equal(th.T, "authenticated", role, "role must be 'authenticated'")

				// ========== ASSERT: Identity record exists (REQUIRED FOR LOGIN) ==========
				var identityCount int
				err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM identities WHERE user_id = $1", userID).Scan(&identityCount)
				require.NoError(th.T, err, "should query identities table")
				assert.Equal(th.T, 1, identityCount, "must have exactly 1 identity record")

				// ========== ASSERT: Identity provider and provider_id ==========
				var provider, providerID string
				err = db.QueryRowContext(ctx, "SELECT provider, provider_id FROM identities WHERE user_id = $1", userID).Scan(&provider, &providerID)
				require.NoError(th.T, err, "should query identity provider details")
				assert.Equal(th.T, "email", provider, "identity.provider must be 'email'")
				assert.Equal(th.T, userID, providerID, "identity.provider_id must equal user_id for email provider")

				// ========== ASSERT: Identity identity_data contains sub and email ==========
				var identityDataSub, identityDataEmail string
				err = db.QueryRowContext(ctx, "SELECT identity_data->>'sub', identity_data->>'email' FROM identities WHERE user_id = $1", userID).Scan(&identityDataSub, &identityDataEmail)
				require.NoError(th.T, err, "should query identity_data")
				assert.Equal(th.T, userID, identityDataSub, "identity_data.sub must equal user_id")
				assert.Equal(th.T, "otp-test@winged-test.local", identityDataEmail, "identity_data.email must match input email")

				// ========== ASSERT: Identity identity_data contains email_verified and phone_verified (REQUIRED FOR OTP) ==========
				var emailVerified, phoneVerified bool
				err = db.QueryRowContext(ctx, "SELECT (identity_data->>'email_verified')::boolean, (identity_data->>'phone_verified')::boolean FROM identities WHERE user_id = $1", userID).Scan(&emailVerified, &phoneVerified)
				require.NoError(th.T, err, "should query identity_data verified flags")
				assert.True(th.T, emailVerified, "identity_data.email_verified must be true")
				assert.False(th.T, phoneVerified, "identity_data.phone_verified must be false")
			},
		},
		{
			name: "error-nil-params",
			setup: func(th *testsuite.Helper) *repo.InsertUserParams {
				return nil
			},
			extraAssertions: func(th *testsuite.Helper, userID string, err error) {
				require.Error(th.T, err, "should error on nil params")
				assert.Empty(th.T, userID, "should not return user ID")
				assert.Contains(th.T, err.Error(), "params cannot be nil")
			},
		},
	}
}

func TestStore_Insert(t *testing.T) {
	for _, tt := range insertUserTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tSuite := testsuite.New(t)
			t.Cleanup(tSuite.UseSupabaseAuthDB())

			params := tt.setup(tSuite)

			store := repo.NewStore()
			userID, err := store.Insert(context.Background(), tSuite.SupabaseAuthDb(), params)

			tt.extraAssertions(tSuite, userID, err)
		})
	}
}

type testCaseDeleteByEmails struct {
	name            string
	setup           func(th *testsuite.Helper) []string
	extraAssertions func(th *testsuite.Helper, deleted int64, err error)
}

func deleteByEmailsTestCases() []testCaseDeleteByEmails {
	return []testCaseDeleteByEmails{
		{
			name: "success-deletes-user-and-cascades-identity",
			setup: func(th *testsuite.Helper) []string {
				store := repo.NewStore()
				email := "delete-test@winged-test.local"

				_, err := store.Insert(context.Background(), th.SupabaseAuthDb(), &repo.InsertUserParams{
					Email: email,
				})
				require.NoError(th.T, err, "setup: should insert user")

				return []string{email}
			},
			extraAssertions: func(th *testsuite.Helper, deleted int64, err error) {
				require.NoError(th.T, err, "should not error on delete")
				assert.Equal(th.T, int64(1), deleted, "should delete 1 user")

				// Verify no users remain
				var userCount int
				err = th.SupabaseAuthDb().QueryRowContext(
					context.Background(),
					"SELECT COUNT(*) FROM users WHERE email = $1",
					"delete-test@winged-test.local",
				).Scan(&userCount)
				require.NoError(th.T, err, "should query users table")
				assert.Equal(th.T, 0, userCount, "should have 0 user records")

				// Verify identities are cascade deleted
				var identityCount int
				err = th.SupabaseAuthDb().QueryRowContext(
					context.Background(),
					"SELECT COUNT(*) FROM identities",
				).Scan(&identityCount)
				require.NoError(th.T, err, "should query identities table")
				assert.Equal(th.T, 0, identityCount, "should have 0 identity records (cascade delete)")
			},
		},
		{
			name: "success-empty-emails-returns-zero",
			setup: func(th *testsuite.Helper) []string {
				return []string{}
			},
			extraAssertions: func(th *testsuite.Helper, deleted int64, err error) {
				require.NoError(th.T, err, "should not error on empty emails")
				assert.Equal(th.T, int64(0), deleted, "should delete 0 users")
			},
		},
		{
			name: "success-non-existent-email-returns-zero",
			setup: func(th *testsuite.Helper) []string {
				return []string{"nonexistent@winged-test.local"}
			},
			extraAssertions: func(th *testsuite.Helper, deleted int64, err error) {
				require.NoError(th.T, err, "should not error on non-existent email")
				assert.Equal(th.T, int64(0), deleted, "should delete 0 users")
			},
		},
	}
}

func TestStore_DeleteByEmails(t *testing.T) {
	for _, tt := range deleteByEmailsTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tSuite := testsuite.New(t)
			t.Cleanup(tSuite.UseSupabaseAuthDB())

			emails := tt.setup(tSuite)

			store := repo.NewStore()
			deleted, err := store.DeleteByEmails(context.Background(), tSuite.SupabaseAuthDb(), emails)

			tt.extraAssertions(tSuite, deleted, err)
		})
	}
}
