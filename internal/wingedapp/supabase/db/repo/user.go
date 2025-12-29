package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	supabasepgmodel "wingedapp/pgtester/internal/wingedapp/supabase/db/supabasepgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
)

// Default instance ID used by Supabase
const defaultInstanceID = "00000000-0000-0000-0000-000000000000"

// InsertUserParams contains data for inserting a user to supabase auth.users.
type InsertUserParams struct {
	Email string
}

// Insert inserts a user into supabase auth.users with all required metadata
// for OTP/email sign-in, plus the corresponding identity record.
// Returns the user ID.
func (s *Store) Insert(
	ctx context.Context,
	exec boil.ContextExecutor,
	params *InsertUserParams,
) (string, error) {
	if params == nil {
		return "", fmt.Errorf("params cannot be nil")
	}

	now := time.Now()
	userID := uuid.NewString()

	// raw_app_meta_data is required for OTP/email sign-in
	// Format: {"provider":"email","providers":["email"]}
	rawAppMetaData, err := json.Marshal(map[string]interface{}{
		"provider":  "email",
		"providers": []string{"email"},
	})
	if err != nil {
		return "", fmt.Errorf("marshal raw_app_meta_data: %w", err)
	}

	// raw_user_meta_data can be empty but must be valid JSON
	rawUserMetaData, err := json.Marshal(map[string]interface{}{})
	if err != nil {
		return "", fmt.Errorf("marshal raw_user_meta_data: %w", err)
	}

	user := &supabasepgmodel.User{
		InstanceID:       null.StringFrom(defaultInstanceID),
		ID:               userID,
		Email:            null.StringFrom(params.Email),
		Aud:              null.StringFrom("authenticated"),
		Role:             null.StringFrom("authenticated"),
		EmailConfirmedAt: null.TimeFrom(now),
		ConfirmedAt:      null.TimeFrom(now),
		CreatedAt:        null.TimeFrom(now),
		UpdatedAt:        null.TimeFrom(now),
		RawAppMetaData:   null.JSONFrom(rawAppMetaData),
		RawUserMetaData:  null.JSONFrom(rawUserMetaData),
	}

	if err := user.Insert(ctx, exec, boil.Infer()); err != nil {
		return "", fmt.Errorf("insert supabase auth user: %w", err)
	}

	// Insert identity record (required for Supabase auth login to work)
	if err := s.insertIdentity(ctx, exec, userID, params.Email, now); err != nil {
		return "", fmt.Errorf("insert supabase auth identity: %w", err)
	}

	return userID, nil
}

// insertIdentity inserts an identity record for a user.
// This is required for Supabase auth to properly authenticate users.
func (s *Store) insertIdentity(
	ctx context.Context,
	exec boil.ContextExecutor,
	userID string,
	email string,
	timestamp time.Time,
) error {
	// Build identity_data JSON with all required fields for OTP login
	// Must include email_verified and phone_verified for Supabase auth to work
	identityData := map[string]interface{}{
		"sub":            userID,
		"email":          email,
		"email_verified": true,
		"phone_verified": false,
	}
	identityDataJSON, err := json.Marshal(identityData)
	if err != nil {
		return fmt.Errorf("marshal identity data: %w", err)
	}

	identityID := uuid.NewString()
	query := `
		INSERT INTO identities (id, user_id, provider_id, identity_data, provider, last_sign_in_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = exec.ExecContext(ctx, query,
		identityID,       // id
		userID,           // user_id
		userID,           // provider_id (same as user_id for email provider)
		identityDataJSON, // identity_data
		"email",          // provider
		timestamp,        // last_sign_in_at
		timestamp,        // created_at
		timestamp,        // updated_at
	)
	if err != nil {
		return fmt.Errorf("exec insert identity: %w", err)
	}

	return nil
}

// InsertUser inserts a user into supabase auth.users (legacy method).
// Deprecated: Use Insert instead.
func (s *Store) InsertUser(
	ctx context.Context,
	exec boil.ContextExecutor,
	user *supabasepgmodel.User,
) error {
	if err := user.Insert(ctx, exec, boil.Infer()); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

// DeleteByEmails deletes users by email addresses.
// Returns the number of users deleted.
func (s *Store) DeleteByEmails(
	ctx context.Context,
	exec boil.ContextExecutor,
	emails []string,
) (int64, error) {
	if len(emails) == 0 {
		return 0, nil
	}

	deleted, err := supabasepgmodel.Users(
		supabasepgmodel.UserWhere.Email.IN(emails),
	).DeleteAll(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("delete supabase users: %w", err)
	}

	return deleted, nil
}
