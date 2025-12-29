package pg

import (
	"context"
	"errors"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/aibackend/db/aipgmodel"
	"wingedapp/pgtester/internal/wingedapp/business/domain/registration"
	"wingedapp/pgtester/internal/wingedapp/db/boilhelper"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

func usersFilter(f *registration.QueryFilterUser) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)

	if f.ID.Valid {
		qMods = append(qMods, pgmodel.UserWhere.ID.EQ(f.ID.String))
	}

	if f.Email.Valid {
		qMods = append(qMods, pgmodel.UserWhere.Email.EQ(f.Email.String))
	}

	if f.MobileCode.Valid {
		qMods = append(qMods, pgmodel.UserWhere.MobileCode.EQ(f.MobileCode))
	}

	return qMods
}

func toRepoUserQueryFilter(filter *registration.QueryFilterUser) *repo.QueryFilterUser {
	return &repo.QueryFilterUser{
		ID:    filter.ID,
		Email: filter.Email,
	}
}

func toRepoUpsertUser(bizUpserter *registration.UpsertUser) *repo.UpsertUser {
	return &repo.UpsertUser{
		ID:                     bizUpserter.ID,
		SupabaseID:             bizUpserter.SupabaseID,
		Email:                  bizUpserter.Email,
		FirstName:              bizUpserter.FirstName,
		LastName:               bizUpserter.LastName,
		Birthday:               bizUpserter.Birthday,
		Gender:                 bizUpserter.Gender,
		Height:                 bizUpserter.Height,
		MobileNumber:           bizUpserter.Number,
		RegistrationCode:       bizUpserter.RegistrationCode,
		RegistrationCodeSentAt: bizUpserter.RegistrationCodeSentAt,
	}
}

// UpdateUser(ctx context.Context, exec boil.ContextExecutor, user *UpdateUser) (*User, error)

// UpsertUser upserts a user to the store
func (s *Store) UpsertUser(ctx context.Context,
	tx boil.ContextTransactor,
	user *registration.UpsertUser,
) (*registration.User, error) {
	if err := s.repoBackendApp.UpsertUser(ctx, toRepoUpsertUser(user), tx); err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}

	userPG, err := s.repoBackendApp.User(ctx, tx, toRepoUserQueryFilter(&registration.QueryFilterUser{
		Email: user.Email,
	}))
	if err != nil {
		return nil, fmt.Errorf("query user after upsert: %w", err)
	}

	return pgUserToRegistrationUser(userPG), nil
}

func pgUserToRegistrationUser(user *pgmodel.User) *registration.User {
	if user == nil {
		return nil
	}

	return &registration.User{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
}

func (s *Store) CountUser(ctx context.Context,
	exec boil.ContextExecutor,
	filter *registration.QueryFilterUser,
) (int64, error) {
	f := usersFilter(filter)
	return pgmodel.Users(f...).Count(ctx, exec)
}

// User retrieves a single user from the database based on the provided filter.
func (s *Store) User(ctx context.Context,
	execBE boil.ContextExecutor,
	execAI boil.ContextExecutor,
	filter *registration.QueryFilterUser,
) (*registration.User, error) {
	if !filter.HasFilters() {
		return nil, fmt.Errorf("no filters provided")
	}

	users, err := s.Users(ctx, execBE, execAI, filter)
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}

	if len(users) == 0 {
		return nil, registration.ErrUserNotFound
	}

	if len(users) != 1 {
		return nil, fmt.Errorf("user count mismatch, have %d, want 1", len(users))
	}

	return &users[0], nil
}

// toPGUpdateUser converts a business UpdateUser to a repo UpdateUser.
func (s *Store) toPGUpdateUser(ctx context.Context, exec boil.ContextTransactor, bizUpdateUser *registration.UpdateUser) (*repo.UpdateUser, error) {
	if bizUpdateUser == nil {
		return nil, nil
	}

	return &repo.UpdateUser{
		ID:                     bizUpdateUser.ID,
		Email:                  bizUpdateUser.Email,
		FirstName:              bizUpdateUser.FirstName,
		LastName:               bizUpdateUser.LastName,
		Birthday:               bizUpdateUser.Birthday,
		Gender:                 bizUpdateUser.Gender,
		Height:                 bizUpdateUser.Height,
		Number:                 bizUpdateUser.Number,
		RegisteredSuccessfully: bizUpdateUser.RegisteredSuccessfully,
		RegistrationCode:       bizUpdateUser.RegistrationCode,
		MobileCode:             bizUpdateUser.MobileCode,
		Sha256Hash:             bizUpdateUser.Sha256Hash,
		MobileConfirmed:        bizUpdateUser.MobileConfirmed,
		RegistrationCodeSentAt: bizUpdateUser.RegistrationCodeSentAt,
		LastCheckedCallStatus:  bizUpdateUser.LastCheckedCallStatus,
		SelectedIntroID:        bizUpdateUser.SelectedVoiceID,
		UserInviteCodeID:       bizUpdateUser.UserInviteCodeID,
		AgentDeployed:          bizUpdateUser.AgentDeployed,
		HasTranscript:          bizUpdateUser.HasTranscript,
		LatestTranscriptID:     bizUpdateUser.LatestTranscriptID,
		LatestTranscriptTS:     bizUpdateUser.LatestTranscriptTS,
		Sexuality:              bizUpdateUser.SexualityCategoryID, // String enum value
		SexualityIsVisible:     bizUpdateUser.SexualityIsVisible,
	}, nil
}

func (s *Store) UpdateUser(ctx context.Context,
	tx boil.ContextTransactor,
	execAI boil.ContextExecutor,
	updater *registration.UpdateUser,
) (*registration.User, error) {
	repoUpdateUserParams, err := s.toPGUpdateUser(ctx, tx, updater)
	if err != nil {
		return nil, fmt.Errorf("convert to repo update user: %w", err)
	}

	if err = s.repoBackendApp.UpdateUser(ctx, repoUpdateUserParams, tx); err != nil {
		return nil, fmt.Errorf("updater user: %w", err)
	}

	// Update user dating preferences
	if updater.DatingPreferences.Male.Valid {
		if err := s.repoBackendApp.UpsertUserDatingPreference(ctx, tx, updater.ID, pgmodel.DatingPreferencesMale); err != nil {
			return nil, fmt.Errorf("updater user dating pref male: %w", err)
		}
	}
	if updater.DatingPreferences.Female.Valid {
		if err := s.repoBackendApp.UpsertUserDatingPreference(ctx, tx, updater.ID, pgmodel.DatingPreferencesFemale); err != nil {
			return nil, fmt.Errorf("updater user dating pref female: %w", err)
		}
	}
	if updater.DatingPreferences.NonBinary.Valid {
		if err := s.repoBackendApp.UpsertUserDatingPreference(ctx, tx, updater.ID, pgmodel.DatingPreferencesNonBinary); err != nil {
			return nil, fmt.Errorf("updater user dating pref non-binary: %w", err)
		}
	}

	user, err := s.User(ctx, tx, execAI, &registration.QueryFilterUser{
		ID: null.StringFrom(updater.ID),
	})
	if err != nil {
		return nil, fmt.Errorf("query user after update: %w", err)
	}

	return user, nil
}

// convIDFromLatestTranscript gets the conversation ID from the latest transcript ID.
func convIDFromLatestTranscript(
	latestTranscriptID string,
	uTrans []registration.UserTranscript,
) null.String {
	ns := null.String{}
	for _, ut := range uTrans { // uTrans is ordered by created at desc
		if latestTranscriptID == ut.ID {
			ns.Valid = true
			ns.String = ut.ConversationID // get latest conversation ID
			break
		}
	}

	return ns
}

// enrichUserCallStates adds call states to the user object,
// which are: transcripts, and audio files.
func (s *Store) enrichUserCallStates(ctx context.Context,
	execAI boil.ContextExecutor,
	user *registration.User,
) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	// get all failures, and successes for transcripts, ensure ordered by created at desc
	transcripts, err := s.Transcripts(ctx, execAI, &registration.TranscriptQueryFilters{
		UserID:                   user.SupabaseID,
		CategoriesCallSuccessful: []string{registration.TranscriptSuccess, registration.TranscriptFailure},
		OrderedBys:               []string{aipgmodel.TranscriptColumns.CreatedAt + " DESC"},
	})
	if err != nil {
		return fmt.Errorf("transcripts: %w", err)
	}

	audioFiles, err := s.AudioFiles(ctx, execAI, &registration.AudioFileQueryFilter{
		UserID:         user.SupabaseID,
		HasStorage:     null.BoolFrom(true),
		ConversationID: convIDFromLatestTranscript(user.LatestTranscriptID.String, transcripts),
		Categories:     []string{audioFileExciting, audioFileGeneric, audioFileVulnerable},
	})
	if err != nil && !errors.Is(err, registration.ErrUserAudioFileNotFound) {
		return fmt.Errorf("audio files: %w", err)
	}

	user.Transcripts = transcripts
	user.HasTranscript = len(transcripts) > 0
	user.AudioFiles = audioFiles

	return nil
}

// enrichUserPhotos adds photos to the user object.
func (s *Store) enrichUserPhotos(ctx context.Context,
	exec boil.ContextExecutor,
	user *registration.User,
) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	photos, err := s.UserPhotos(ctx, exec, &registration.UserPhotoQueryFilter{
		UserID: null.StringFrom(user.ID),
	})
	if err != nil {
		return fmt.Errorf("user photos: %w", err)
	}

	user.Photos = photos

	return nil
}

/*
// Totals retrieves the totals for a user by their UUID.
func (u *UserTotalsStore) Totals(ctx context.Context, exec boil.ContextExecutor, uuid string) (*economy.UserTotals, error) {
	var t economy.UserTotals

	col := pgmodel.WingsEcnUserTotalColumns
	where := pgmodel.WingsEcnUserTotalWhere
	sel := qm.Select

	if err := pgmodel.WingsEcnUserTotals(
		sel(
			col.ID+" AS id",
			col.TotalWings+" AS wings",
			col.CounterSentMessages+" AS total_sent_messages",
			col.PremiumExpiresIn+" AS premium_expires_in",
		),
		where.UserID.EQ(uuid),
	).Bind(ctx, exec, &t); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("pgmodel single user totals: %w", err)
	}

	return &t, nil
}
*/

type Users struct {
	pgmodel.UserSlice
}

func usersFilterExperiment(f *registration.QueryFilterUser) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)

	if f.ID.Valid {
		qMods = append(qMods, pgmodel.UserWhere.ID.EQ(f.ID.String))
	}

	if f.UserTypeID.Valid {
		qMods = append(qMods, pgmodel.UserWhere.UserType.EQ(f.UserTypeID.String))
	}

	if f.Email.Valid {
		qMods = append(qMods, pgmodel.UserWhere.Email.EQ(f.Email.String))
	}

	if f.MobileCode.Valid {
		qMods = append(qMods, pgmodel.UserWhere.MobileCode.EQ(f.MobileCode))
	}

	return qMods
}

// UsersOld retrieves multiple users from the database based on the provided filter.
// TODO: refactor to be more straightforward.
func (s *Store) UsersOld(ctx context.Context,
	execBE boil.ContextExecutor,
	execAI boil.ContextExecutor,
	f *registration.QueryFilterUser,
) ([]registration.User, error) {
	var users []registration.User

	qMods := func() []qm.QueryMod {
		qModBase := usersFilterExperiment(f)
		qModCols := qm.Select(
			boilhelper.QmSelect([]boilhelper.QmColSet{
				{
					TableName: pgmodel.TableNames.Users,
					Cols: []boilhelper.QmCol{ // local cols
						{Name: pgmodel.UserColumns.ID},
						{Name: pgmodel.UserColumns.SupabaseID},
						{Name: pgmodel.UserColumns.FirstName},
						{Name: pgmodel.UserColumns.LastName},
						{Name: pgmodel.UserColumns.Birthday},
						{Name: pgmodel.UserColumns.Email},
						{Name: pgmodel.UserColumns.Gender},
						{Name: pgmodel.UserColumns.HeightCM, Alias: "height"},
						{Name: pgmodel.UserColumns.RegistrationCode},
						{Name: pgmodel.UserColumns.MobileConfirmed},
						{Name: pgmodel.UserColumns.MobileNumber},
						{Name: pgmodel.UserColumns.MobileCode},
						{Name: pgmodel.UserColumns.RegisteredSuccessfully},
						{Name: pgmodel.UserColumns.RegistrationCodeSentAt},
						{Name: pgmodel.UserColumns.LastCheckedCallStatus},
						{Name: pgmodel.UserColumns.AgentDeployed},
						{Name: pgmodel.UserColumns.SelectedIntroID},
						{Name: pgmodel.UserColumns.LatestTranscriptID},
						{Name: pgmodel.UserColumns.LatestTranscriptTS},
						{Name: pgmodel.UserColumns.HasTranscript},
						{Name: pgmodel.UserColumns.UserType},
						{Name: pgmodel.UserColumns.Sexuality},
					},
				},
			},
			)...,
		)

		return append(
			qModBase,
			qModCols,
		)
	}

	if err := pgmodel.Users(qMods()...).Bind(ctx, execBE, &users); err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}

	for i := range users {
		pgUserDatingPrefs, err := s.repoBackendApp.UserDatingPreferences(ctx, execBE, &repo.UserDatingPreferencesQueryFilter{
			UserID: null.StringFrom(users[i].ID),
		})
		if err != nil {
			return nil, fmt.Errorf("user dating pref repo: %w", err)
		}

		users[i].DatingPreferences = storeUserDatingPrefsToSlice(pgUserDatingPrefs)

		if f.EnrichPhotos {
			if err := s.enrichUserPhotos(ctx, execBE, &users[i]); err != nil {
				return nil, fmt.Errorf("enrich user photos: %w", err)
			}
		}

		if f.EnrichCallStates {
			if err := s.enrichUserCallStates(ctx, execAI, &users[i]); err != nil {
				return nil, fmt.Errorf("enrich user call states: %w", err)
			}
		}
	}

	return users, nil
}

func (s *Store) Users(ctx context.Context,
	execBE boil.ContextExecutor,
	execAI boil.ContextExecutor,
	f *registration.QueryFilterUser,
) ([]registration.User, error) {
	var users []registration.User

	// tables
	usersTbl := pgmodel.TableNames.Users

	// cols
	usrCols := pgmodel.UserColumns

	qMods := append(
		usersFilter2(f),
		qm.Select(
			"u."+usrCols.ID+" AS "+usrCols.ID,
			"u."+usrCols.SupabaseID+" AS "+usrCols.SupabaseID,
			"u."+usrCols.FirstName+" AS "+usrCols.FirstName,
			"u."+usrCols.LastName+" AS "+usrCols.LastName,
			"u."+usrCols.Birthday+" AS "+usrCols.Birthday,
			"u."+usrCols.Email+" AS "+usrCols.Email,
			"u."+usrCols.Gender+" AS "+usrCols.Gender,
			"u."+usrCols.HeightCM+" AS height",
			"u."+usrCols.RegistrationCode+" AS "+usrCols.RegistrationCode,
			"u."+usrCols.MobileConfirmed+" AS "+usrCols.MobileConfirmed,
			"u."+usrCols.MobileNumber+" AS "+usrCols.MobileNumber,
			"u."+usrCols.MobileCode+" AS "+usrCols.MobileCode,
			"u."+usrCols.RegisteredSuccessfully+" AS "+usrCols.RegisteredSuccessfully,
			"u."+usrCols.RegistrationCodeSentAt+" AS "+usrCols.RegistrationCodeSentAt,
			"u."+usrCols.LastCheckedCallStatus+" AS "+usrCols.LastCheckedCallStatus,
			"u."+usrCols.AgentDeployed+" AS "+usrCols.AgentDeployed,
			"u."+usrCols.SelectedIntroID+" AS "+usrCols.SelectedIntroID,
			"u."+usrCols.LatestTranscriptID+" AS "+usrCols.LatestTranscriptID,
			"u."+usrCols.LatestTranscriptTS+" AS "+usrCols.LatestTranscriptTS,
			"u."+usrCols.HasTranscript+" AS "+usrCols.HasTranscript,
			"u."+usrCols.Sexuality+" AS sexuality",
			"u."+usrCols.SexualityIsVisible+" AS "+usrCols.SexualityIsVisible,
			"u."+usrCols.UserType+" AS user_type",
			"u."+usrCols.UserInviteCodeRefID+" AS "+usrCols.UserInviteCodeRefID,
		),
		qm.From(usersTbl+" u"),
	)

	if err := pgmodel.NewQuery(qMods...).Bind(ctx, execBE, &users); err != nil {
		return nil, fmt.Errorf("user bind: %w", err)
	}

	// keep this later
	for i := range users {
		pgUserDatingPrefs, err := s.repoBackendApp.UserDatingPreferences(ctx, execBE, &repo.UserDatingPreferencesQueryFilter{
			UserID: null.StringFrom(users[i].ID),
		})
		if err != nil {
			return nil, fmt.Errorf("user dating pref repo: %w", err)
		}

		users[i].DatingPreferences = storeUserDatingPrefsToSlice(pgUserDatingPrefs)

		if f.EnrichPhotos {
			if err := s.enrichUserPhotos(ctx, execBE, &users[i]); err != nil {
				return nil, fmt.Errorf("enrich user photos: %w", err)
			}
		}

		if f.EnrichCallStates {
			if err = s.enrichUserCallStates(ctx, execAI, &users[i]); err != nil {
				return nil, fmt.Errorf("enrich user call states: %w", err)
			}
		}
	}

	return users, nil
}

func usersFilter2(f *registration.QueryFilterUser) []qm.QueryMod {
	qMods := make([]qm.QueryMod, 0)

	usrCol := pgmodel.UserColumns

	if f.ID.Valid {
		qMods = append(qMods, qm.Where("u"+"."+usrCol.ID+"=?", f.ID.String))
	}

	if f.UserTypeID.Valid {
		qMods = append(qMods, qm.Where("u"+"."+usrCol.UserType+"=?", f.UserTypeID.String))
	}

	if f.Email.Valid {
		qMods = append(qMods, qm.Where("u"+"."+usrCol.Email+"=?", f.Email.String))
	}

	if f.MobileCode.Valid {
		qMods = append(qMods, qm.Where("u"+"."+usrCol.MobileCode+"=?", f.MobileCode.String))
	}

	return qMods
}

func storeUserDatingPrefsToSlice(prefs pgmodel.UserDatingPreferenceSlice) []string {
	result := make([]string, 0, len(prefs))
	for _, pref := range prefs {
		result = append(result, pref.DatingPreference)
	}
	return result
}
