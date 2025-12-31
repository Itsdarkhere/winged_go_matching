package apprepo

import (
	"context"
	"fmt"
	"strings"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	supabasepgmodel "wingedapp/pgtester/internal/wingedapp/supabase/db/supabasepgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
)

var (
	// toDeleteBackendApp - ordered list of tables to delete (FK order matters)
	// Format: {table, column} where column is the user FK column
	toDeleteBackendApp = []struct{ table, column string }{
		// 1. Tables with FK to match_result (must delete before match_result)
		{pgmodel.TableNames.MatchChatMessage, "sender_id"},
		// 2. match_result (initiator and receiver user refs)
		{pgmodel.TableNames.MatchResult, "initiator_user_ref_id"},
		{pgmodel.TableNames.MatchResult, "receiver_user_ref_id"},
		// 3. Other tables with user_ref_id
		{pgmodel.TableNames.WingsEcnTransaction, "user_ref_id"},
		{pgmodel.TableNames.WingsEcnActionLog, "user_ref_id"},
		{pgmodel.TableNames.WingsEcnUserTotals, "user_ref_id"},
		{pgmodel.TableNames.AgentLog, "user_ref_id"},
		// 4. Tables with user_id
		{pgmodel.TableNames.UserPhoto, "user_id"},
		{pgmodel.TableNames.UserAiConvo, "user_id"},
		{pgmodel.TableNames.UserBlockedContact, "user_id"},
		{pgmodel.TableNames.UserDatingPreferences, "user_id"},
		{pgmodel.TableNames.UserAiContext, "user_id"},
		{pgmodel.TableNames.UserAvailability, "user_id"},
		{pgmodel.TableNames.UserElevenLabs, "user_id"},
		{pgmodel.TableNames.WingsEcnUserSubscriptionPlan, "user_id"},
		// 5. Users table last
		{pgmodel.TableNames.Users, "id"},
	}

	toDeleteAIBackend = []string{
		aipgmodel.TableNames.AudioFiles,
		aipgmodel.TableNames.Profiles,
		aipgmodel.TableNames.Transcripts,
	}
)

type Store struct {
}

// DeleteUserData deletes all user data associated with the given userID (backend_app.users.id).
// Order matters: FK-dependent tables are deleted first.
// ai_backend tables use supabase_id (not backend_app user_id) since audio_files.user_id FK points to supabase.auth.users.
// supabaseExec is optional - pass nil to skip supabase auth.users deletion.
// Users without a valid supabase_id (e.g., test match users) skip ai_backend and supabase deletion.
func (s *Store) DeleteUserData(ctx context.Context,
	beExec boil.ContextExecutor,
	aiExec boil.ContextExecutor,
	supabaseExec boil.ContextExecutor,
	userID string,
) error {
	fmt.Println("==== hello hello 1")
	// Look up supabase_id for ai_backend deletion
	user, err := pgmodel.FindUser(ctx, beExec, userID)
	if err != nil {
		return fmt.Errorf("finding user %s: %w", userID, err)
	}

	// Only delete ai_backend and supabase data if user has a valid supabase_id
	// (test match users may not have supabase accounts)
	if user.SupabaseID.Valid {
		for _, tbl := range toDeleteAIBackend {
			query := fmt.Sprintf("DELETE FROM %s WHERE user_id = $1", tbl)
			if _, err := aiExec.ExecContext(ctx, query, user.SupabaseID.String); err != nil {
				return fmt.Errorf("deleting ai backend supabase ID from table %s: %w", tbl, err)
			}
		}
	}

	fmt.Println("==== hello hello")
	for _, tbl := range toDeleteAIBackend {
		fmt.Println("=== gonna delete:", user.ID, " from table ", tbl)
		query := fmt.Sprintf("DELETE FROM %s WHERE user_id = $1", tbl)
		if _, err := aiExec.ExecContext(ctx, query, user.ID); err != nil {
			return fmt.Errorf("deleting ai backend with user-id data from table %s: %w", tbl, err)
		}
	}

	// delete backend_app data in FK order
	for _, spec := range toDeleteBackendApp {
		query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", spec.table, spec.column)
		if _, err := beExec.ExecContext(ctx, query, userID); err != nil {
			// Skip if table/column doesn't exist (schema mismatch)
			if strings.Contains(err.Error(), "does not exist") {
				continue
			}
			return fmt.Errorf("deleting from %s: %w", spec.table, err)
		}
	}

	if user.SupabaseID.Valid {
		// delete from supabase auth.users
		if _, err := supabasepgmodel.Users(
			supabasepgmodel.UserWhere.ID.EQ(user.SupabaseID.String),
		).DeleteAll(ctx, supabaseExec); err != nil {
			return fmt.Errorf("deleting supabase user %s: %w", user.SupabaseID.String, err)
		}
	}

	return nil
}

// DeleteUserDataByEmail deletes all user data for users matching the given emails.
// Returns the count of deleted users.
// This is used by population ingestion for cleanup before re-inserting.
// supabaseExec is optional - pass nil to skip supabase auth.users deletion.
func (s *Store) DeleteUserDataByEmail(ctx context.Context,
	beExec boil.ContextExecutor,
	aiExec boil.ContextExecutor,
	supabaseExec boil.ContextExecutor,
	emails []string,
) (int64, error) {
	if len(emails) == 0 {
		return 0, nil
	}

	// delete backend users
	beUsers, err := pgmodel.Users(
		pgmodel.UserWhere.Email.IN(emails),
	).All(ctx, beExec)
	if err != nil {
		return 0, fmt.Errorf("finding users by email: %w", err)
	}

	for _, user := range beUsers {
		// delete backend users and all associated data to ai_backend
		if err := s.DeleteUserData(ctx, beExec, aiExec, supabaseExec, user.ID); err != nil {
			return 0, fmt.Errorf("deleting user %s: %w", user.ID, err)
		}
	}

	// delete supabase users (in case the email collides with something on supabase
	// but not on backend_app)
	supabaseUsers, err := supabasepgmodel.Users(
		supabasepgmodel.UserWhere.Email.IN(emails),
	).All(ctx, supabaseExec)
	if err != nil {
		return 0, fmt.Errorf("finding supabase users by email for debug: %w", err)
	}
	tables := []string{
		aipgmodel.TableNames.AudioFiles,
		aipgmodel.TableNames.Profiles,
		aipgmodel.TableNames.Transcripts,
	}

	for i := range supabaseUsers {
		sBaseUser := supabaseUsers[i]

		// delete from ai_backend tables
		for _, tbl := range tables {
			query := fmt.Sprintf("DELETE FROM %s WHERE user_id = $1", tbl)
			if _, err := aiExec.ExecContext(ctx, query, sBaseUser.ID); err != nil {
				return 0, fmt.Errorf("deleting table %s: %w", tbl, err)
			}
		}
	}

	if _, err := supabasepgmodel.Users(
		supabasepgmodel.UserWhere.Email.IN(emails),
	).DeleteAll(ctx, supabaseExec); err != nil {
		return 0, fmt.Errorf("deleting supabase users by email: %w", err)
	}

	return int64(len(beUsers) + len(supabaseUsers)), nil
}
