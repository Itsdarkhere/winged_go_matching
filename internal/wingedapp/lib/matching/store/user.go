package store

import (
	"context"
	"fmt"
	"time"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/google/uuid"
)

type UserStore struct {
	l    applog.Logger
	repo *repo.Store
}

func (s *UserStore) UserDatingPreferences(ctx context.Context,
	exec boil.ContextExecutor,
	user *matching.QueryFilterUser,
) ([]matching.UserDatingPreference, error) {
	return nil, nil
}

// Insert inserts a user into backend_app.users and returns the user ID.
func (s *UserStore) Insert(
	ctx context.Context,
	exec boil.ContextExecutor,
	params *matching.InsertPopulationUser,
) (string, error) {
	if params == nil {
		return "", fmt.Errorf("params cannot be nil")
	}

	user := &pgmodel.User{
		ID:                     uuid.NewString(),
		Email:                  params.Email,
		FirstName:              null.StringFrom(params.FirstName),
		LastName:               null.StringFrom(params.LastName),
		Gender:                 null.StringFrom(params.Gender),
		HeightCM:               null.IntFrom(params.HeightCM),
		Latitude:               null.Float64From(params.Latitude),
		Longitude:              null.Float64From(params.Longitude),
		Birthday:               null.TimeFrom(params.Birthday), // Directly from CSV
		Address:                null.StringFrom(params.Address),
		IsActive:               null.BoolFrom(true),
		RegisteredSuccessfully: null.BoolFrom(true),
		MobileConfirmed:        null.BoolFrom(true),
		IsTestUser:             null.BoolFrom(params.IsTestUser),
		CreatedAt:              null.TimeFrom(time.Now()),
		UpdatedAt:              null.TimeFrom(time.Now()),
	}

	// Whitelist columns - 'location' column was removed from DB but pgmodel not regenerated yet
	if err := user.Insert(ctx, exec, boil.Whitelist(
		pgmodel.UserColumns.ID,
		pgmodel.UserColumns.Email,
		pgmodel.UserColumns.FirstName,
		pgmodel.UserColumns.LastName,
		pgmodel.UserColumns.Gender,
		pgmodel.UserColumns.HeightCM,
		pgmodel.UserColumns.Latitude,
		pgmodel.UserColumns.Longitude,
		pgmodel.UserColumns.Birthday,
		pgmodel.UserColumns.Address,
		pgmodel.UserColumns.IsActive,
		pgmodel.UserColumns.RegisteredSuccessfully,
		pgmodel.UserColumns.MobileConfirmed,
		pgmodel.UserColumns.IsTestUser,
		pgmodel.UserColumns.CreatedAt,
		pgmodel.UserColumns.UpdatedAt,
	)); err != nil {
		return "", fmt.Errorf("insert backend_app user: %w", err)
	}

	return user.ID, nil
}

func (s *UserStore) Users(ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterUser,
) ([]matching.User, error) {
	var users []matching.User

	// tables
	userTbl := pgmodel.TableNames.Users

	// cols
	userCols := pgmodel.UserColumns

	qMods := append(
		qModsUser(f),
		qm.Select(
			"u."+userCols.ID+" AS id",
			"u."+userCols.Email+" AS email",
			"u."+userCols.FirstName+" AS firstname",
			"u."+userCols.LastName+" AS lastname",
			"CASE WHEN u."+userCols.Birthday+" IS NOT NULL THEN EXTRACT(YEAR FROM AGE(u."+userCols.Birthday+"))::int END AS age",
			"u."+userCols.Gender+" AS gender",
			"u."+userCols.HeightCM+" AS height",
			"u."+userCols.Latitude+" AS latitude",
			"u."+userCols.Longitude+" AS longitude",
			"u."+userCols.UserType+" AS user_type",
			"u."+userCols.IsTestUser+" AS is_test_user",
		),
		qm.From(userTbl+" u"),
	)

	if err := pgmodel.NewQuery(qMods...).Bind(ctx, exec, &users); err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}

	return users, nil
}

func (s *UserStore) User(ctx context.Context,
	exec boil.ContextExecutor,
	f *matching.QueryFilterUser,
) (*matching.User, error) {
	users, err := s.Users(ctx, exec, f)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no user found")
	}

	if len(users) > 1 {
		return nil, fmt.Errorf("user count mismatch, have %d, want 1", len(users))
	}

	return &users[0], nil
}

func qModsUser(f *matching.QueryFilterUser) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)

	if f.ID.Valid {
		qMods = append(qMods, qm.Where("u.id=?", f.ID.String))
	}

	if f.IsActive.Valid {
		qMods = append(qMods, qm.Where("u.is_active=?", f.IsActive.Bool))
	}

	if f.IsTestUser.Valid {
		qMods = append(qMods, qm.Where("u.is_test_user=?", f.IsTestUser.Bool))
	}

	if f.UserType.Valid {
		qMods = append(qMods, qm.Where("u.user_type=?", f.UserType.String))
	}

	return qMods
}

// DeleteByEmails deletes users and their dating preferences by email addresses.
// Returns the number of users deleted.
func (s *UserStore) DeleteByEmails(
	ctx context.Context,
	exec boil.ContextExecutor,
	emails []string,
) (int64, error) {
	if len(emails) == 0 {
		return 0, nil
	}

	// First get user IDs for these emails
	users, err := pgmodel.Users(
		pgmodel.UserWhere.Email.IN(emails),
	).All(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("find users by email: %w", err)
	}

	if len(users) == 0 {
		return 0, nil
	}

	userIDs := make([]string, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	// Delete dating preferences first (FK constraint)
	_, err = pgmodel.UserDatingPreferences(
		pgmodel.UserDatingPreferenceWhere.UserID.IN(userIDs),
	).DeleteAll(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("delete dating preferences: %w", err)
	}

	// Delete users
	deleted, err := pgmodel.Users(
		pgmodel.UserWhere.ID.IN(userIDs),
	).DeleteAll(ctx, exec)
	if err != nil {
		return 0, fmt.Errorf("delete users: %w", err)
	}

	s.l.Debug(ctx, "deleted backend_app users by email", applog.F("deleted_count", deleted), applog.F("emails", emails))

	return deleted, nil
}
