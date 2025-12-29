package matching

import (
	"context"
	"fmt"
	"strings"
	"time"
	"wingedapp/pgtester/internal/util/strutil"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
)

// InsertPopulationUser contains data for inserting a user to backend_app.users
type InsertPopulationUser struct {
	Email             string    // Generated if not provided
	FirstName         string    // Required
	LastName          string    // Required
	Birthday          time.Time // Directly from CSV (YYYY-MM-DD)
	Gender            string    // Male, Female, Non-binary
	HeightCM          int       // Height in centimeters
	Latitude          float64
	Longitude         float64
	Address           string   // Optional address string
	DatingPreferences []string // Expanded from DatingPreference (Male, Female, Any)
	IsTestUser        bool     // Whether this is a test user (excluded from prod matching)
}

// InsertPopulationProfile contains data for inserting a profile to ai_backend.profiles
type InsertPopulationProfile struct {
	UserID  string
	Details *ProfileDetails
}

// InsertSupabaseUser contains data for inserting a user to supabase auth.users
type InsertSupabaseUser struct {
	Email string
}

// PopulationExecutors holds the database executors for population operations.
type PopulationExecutors struct {
	BackendApp   boil.ContextExecutor
	AIBackend    boil.ContextExecutor
	SupabaseAuth boil.ContextExecutor
}

// PopulateOptions contains optional settings for population.
type PopulateOptions struct {
	IsTestUser bool // Mark all populated users as test users
}

// Populate creates users across databases from parsed CSV rows.
// It first cleans any existing conflicting records by email, then inserts new records.
// It coordinates insertions to:
//   - backend_app.users (required - with dating preferences)
//   - ai_backend.profiles (optional - skipped if AIBackend executor is nil)
//   - supabase auth.users (optional - skipped if SupabaseAuth executor is nil)
//
// If email is provided in CSV, it's used directly. Otherwise, unique emails are generated from names.
// If options.IsTestUser is true, all users will be marked as test users.
// Returns a PopulateResult with counts and any errors encountered.
func (l *Logic) Populate(
	ctx context.Context,
	execs *PopulationExecutors,
	rows []PopulationRow,
	options *PopulateOptions,
) (*PopulateResult, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("no rows to populate")
	}

	result := &PopulateResult{}

	// Use email from CSV if provided, otherwise generate unique emails
	emails := make([]string, len(rows))
	for i, row := range rows {
		if row.Email != "" {
			emails[i] = row.Email
		} else {
			emails[i] = generateEmail(row.FirstName, row.LastName)
		}
	}

	// Clean existing conflicts before inserting
	if err := l.cleanConflicts(ctx, execs, emails, result); err != nil {
		return nil, fmt.Errorf("clean conflicts: %w", err)
	}

	for i, row := range rows {
		rowNum := i + 2 // 1-indexed, account for header row
		email := emails[i]

		// 1. Insert into supabase auth.users first (if executor provided)
		if execs.SupabaseAuth != nil && l.supabaseUserStorer != nil {
			_, err := l.supabaseUserStorer.Insert(ctx, execs.SupabaseAuth, &InsertSupabaseUser{
				Email: email,
			})
			if err != nil {
				return nil, fmt.Errorf("row %d: insert supabase auth user: %w", rowNum, err)
			}
			result.SupabaseAuthUsers++
		}

		// 2. Insert into backend_app.users (required)
		isTestUser := false
		if options != nil {
			isTestUser = options.IsTestUser
		}
		userParams := &InsertPopulationUser{
			Email:             email,
			FirstName:         row.FirstName,
			LastName:          row.LastName,
			Birthday:          row.Birthday,
			Gender:            row.Gender,
			HeightCM:          row.HeightCM,
			Latitude:          row.Latitude,
			Longitude:         row.Longitude,
			Address:           row.Address,
			DatingPreferences: row.DatingPreferences,
			IsTestUser:        isTestUser,
		}

		backendUserID, err := l.userStorer.Insert(ctx, execs.BackendApp, userParams)
		if err != nil {
			return nil, fmt.Errorf("row %d: insert backend_app user: %w", rowNum, err)
		}
		result.BackendAppUsers++

		// 3. Insert dating preferences
		if len(row.DatingPreferences) > 0 {
			if err := l.userDatingPreferenceStorer.Insert(ctx, execs.BackendApp, backendUserID, row.DatingPreferences); err != nil {
				return nil, fmt.Errorf("row %d: insert dating preferences: %w", rowNum, err)
			}
			result.DatingPreferences += len(row.DatingPreferences)
		}

		// 4. Insert into ai_backend.profiles (if executor provided)
		if execs.AIBackend != nil && row.ProfileDetails != nil {
			profileParams := &InsertPopulationProfile{
				UserID:  backendUserID,
				Details: row.ProfileDetails,
			}
			if err := l.profileStorer.Insert(ctx, execs.AIBackend, profileParams); err != nil {
				return nil, fmt.Errorf("row %d: insert ai_backend profile: %w", rowNum, err)
			}
			result.AIBackendProfiles++
		}

		result.TotalProcessed++
	}

	return result, nil
}

// generateEmail creates a unique email from first and last name.
// Format: firstname.lastname.shortUUID@winged-test.local
func generateEmail(firstName, lastName string) string {
	shortID := uuid.New().String()[:8] // First 8 chars of UUID for uniqueness
	firstName = strings.ToLower(strings.ReplaceAll(firstName, " ", ""))
	lastName = strings.ToLower(strings.ReplaceAll(lastName, " ", ""))
	return fmt.Sprintf("%s.%s.%s@winged-test.local", firstName, lastName, shortID)
}

// cleanConflicts removes existing records that would conflict with the new data.
// Uses the centralized userDeleter to cascade delete all user-related tables.
func (l *Logic) cleanConflicts(
	ctx context.Context,
	execs *PopulationExecutors,
	emails []string,
	result *PopulateResult,
) error {
	if len(emails) == 0 {
		return nil
	}

	fmt.Println("==== execs:", strutil.GetAsJson(execs))

	deleted, err := l.userDeleter.DeleteUserDataByEmail(ctx, execs.BackendApp, execs.AIBackend, execs.SupabaseAuth, emails)
	if err != nil {
		return fmt.Errorf("delete user data by email: %w", err)
	}
	result.CleanedBackendAppUsers = int(deleted)

	return nil
}
