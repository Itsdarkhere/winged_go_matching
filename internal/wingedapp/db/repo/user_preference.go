package repo

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// SyncUserPreference is a generic sync request for preference tables.
// Values are now string enum values (not UUIDs).
type SyncUserPreference struct {
	UserID string
	Values []string // String enum values
}

// SyncUserPreferenceResult is the result of a sync operation.
type SyncUserPreferenceResult struct {
	Deleted  int64
	Inserted int64
}

// SyncDietaryRestrictions replaces all dietary restrictions for a user.
// Pattern: DELETE all for user → INSERT new rows
func (s *Store) SyncDietaryRestrictions(
	ctx context.Context,
	exec boil.ContextExecutor,
	sync *SyncUserPreference,
) (*SyncUserPreferenceResult, error) {
	if sync.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// Delete existing
	deleted, err := pgmodel.UserDietaryRestrictions(
		pgmodel.UserDietaryRestrictionWhere.UserID.EQ(sync.UserID),
	).DeleteAll(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("delete dietary restrictions: %w", err)
	}

	// Insert new (skip if empty)
	if len(sync.Values) == 0 {
		return &SyncUserPreferenceResult{Deleted: deleted, Inserted: 0}, nil
	}

	// Bulk insert
	cols := pgmodel.UserDietaryRestrictionColumns
	tbl := pgmodel.TableNames.UserDietaryRestriction

	query := fmt.Sprintf(
		"INSERT INTO %s (%s, %s) VALUES ",
		tbl, cols.UserID, cols.DietaryRestriction,
	)

	args := make([]interface{}, 0, len(sync.Values)*2)
	for i, val := range sync.Values {
		if i > 0 {
			query += ", "
		}
		paramIdx := i * 2
		query += fmt.Sprintf("($%d, $%d)", paramIdx+1, paramIdx+2)
		args = append(args, sync.UserID, val)
	}

	result, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("insert dietary restrictions: %w", err)
	}

	inserted, _ := result.RowsAffected()
	return &SyncUserPreferenceResult{Deleted: deleted, Inserted: inserted}, nil
}

// SyncDateTypePreferences replaces all date type preferences for a user.
func (s *Store) SyncDateTypePreferences(
	ctx context.Context,
	exec boil.ContextExecutor,
	sync *SyncUserPreference,
) (*SyncUserPreferenceResult, error) {
	if sync.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	deleted, err := pgmodel.UserDateTypePreferences(
		pgmodel.UserDateTypePreferenceWhere.UserID.EQ(sync.UserID),
	).DeleteAll(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("delete date type preferences: %w", err)
	}

	if len(sync.Values) == 0 {
		return &SyncUserPreferenceResult{Deleted: deleted, Inserted: 0}, nil
	}

	cols := pgmodel.UserDateTypePreferenceColumns
	tbl := pgmodel.TableNames.UserDateTypePreference

	query := fmt.Sprintf(
		"INSERT INTO %s (%s, %s) VALUES ",
		tbl, cols.UserID, cols.DateTypeCore,
	)

	args := make([]interface{}, 0, len(sync.Values)*2)
	for i, val := range sync.Values {
		if i > 0 {
			query += ", "
		}
		paramIdx := i * 2
		query += fmt.Sprintf("($%d, $%d)", paramIdx+1, paramIdx+2)
		args = append(args, sync.UserID, val)
	}

	result, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("insert date type preferences: %w", err)
	}

	inserted, _ := result.RowsAffected()
	return &SyncUserPreferenceResult{Deleted: deleted, Inserted: inserted}, nil
}

// SyncMobilityConstraints replaces all mobility constraints for a user.
func (s *Store) SyncMobilityConstraints(
	ctx context.Context,
	exec boil.ContextExecutor,
	sync *SyncUserPreference,
) (*SyncUserPreferenceResult, error) {
	if sync.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	deleted, err := pgmodel.UserMobilityConstraints(
		pgmodel.UserMobilityConstraintWhere.UserID.EQ(sync.UserID),
	).DeleteAll(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("delete mobility constraints: %w", err)
	}

	if len(sync.Values) == 0 {
		return &SyncUserPreferenceResult{Deleted: deleted, Inserted: 0}, nil
	}

	cols := pgmodel.UserMobilityConstraintColumns
	tbl := pgmodel.TableNames.UserMobilityConstraint

	query := fmt.Sprintf(
		"INSERT INTO %s (%s, %s) VALUES ",
		tbl, cols.UserID, cols.MobilityConstraint,
	)

	args := make([]interface{}, 0, len(sync.Values)*2)
	for i, val := range sync.Values {
		if i > 0 {
			query += ", "
		}
		paramIdx := i * 2
		query += fmt.Sprintf("($%d, $%d)", paramIdx+1, paramIdx+2)
		args = append(args, sync.UserID, val)
	}

	result, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("insert mobility constraints: %w", err)
	}

	inserted, _ := result.RowsAffected()
	return &SyncUserPreferenceResult{Deleted: deleted, Inserted: inserted}, nil
}

// --- Query methods for GET ---

// UserDietaryRestrictions returns all dietary restriction enum values for a user.
func (s *Store) UserDietaryRestrictions(
	ctx context.Context,
	exec boil.ContextExecutor,
	userID string,
) ([]string, error) {
	rows, err := pgmodel.UserDietaryRestrictions(
		pgmodel.UserDietaryRestrictionWhere.UserID.EQ(userID),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("query dietary restrictions: %w", err)
	}

	values := make([]string, len(rows))
	for i, r := range rows {
		values[i] = r.DietaryRestriction
	}
	return values, nil
}

// UserDateTypePreferences returns all date type preference enum values for a user.
func (s *Store) UserDateTypePreferences(
	ctx context.Context,
	exec boil.ContextExecutor,
	userID string,
) ([]string, error) {
	rows, err := pgmodel.UserDateTypePreferences(
		pgmodel.UserDateTypePreferenceWhere.UserID.EQ(userID),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("query date type preferences: %w", err)
	}

	values := make([]string, len(rows))
	for i, r := range rows {
		values[i] = r.DateTypeCore
	}
	return values, nil
}

// UserMobilityConstraints returns all mobility constraint enum values for a user.
func (s *Store) UserMobilityConstraints(
	ctx context.Context,
	exec boil.ContextExecutor,
	userID string,
) ([]string, error) {
	rows, err := pgmodel.UserMobilityConstraints(
		pgmodel.UserMobilityConstraintWhere.UserID.EQ(userID),
	).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("query mobility constraints: %w", err)
	}

	values := make([]string, len(rows))
	for i, r := range rows {
		values[i] = r.MobilityConstraint
	}
	return values, nil
}
