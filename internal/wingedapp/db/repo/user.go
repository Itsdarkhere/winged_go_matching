package repo

import (
	"context"
	"errors"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jackc/pgx/v4"
)

func NewUserEmailFilter(email string) *QueryFilterUser {
	return &QueryFilterUser{
		Email: null.String{
			String: email,
			Valid:  true,
		},
	}
}

type UpsertUser struct {
	ID                     null.String `json:"id"`
	SupabaseID             null.String `json:"supabase_id"`
	Email                  null.String `json:"email"`
	FirstName              null.String `json:"first_name"`
	LastName               null.String `json:"last_name"`
	Birthday               null.Time   `json:"birthday"`
	Gender                 null.String `json:"gender"`
	Height                 null.Int    `json:"height"`
	MobileNumber           null.String `json:"mobilenumber"`
	RegistrationCode       null.String `json:"registration_code"`
	RegisteredSuccessfully null.Bool   `json:"registered_successfully"`
	RegistrationCodeSentAt null.Time   `json:"registration_code_sent_at"`
	MobileCode             null.String `json:"mobile_code"`
	MobileConfirmed        null.Bool   `json:"mobile_confirmed"`
}

// UpdateUser represents the data needed to update a user.
type UpdateUser struct {
	ID                      string      `json:"id"`
	Email                   null.String `json:"email"`
	FirstName               null.String `json:"first_name"`
	LastName                null.String `json:"last_name"`
	Birthday                null.Time   `json:"birthday"`
	Gender                  null.String `json:"gender"`
	Height                  null.Int    `json:"height"`
	Number                  null.String `json:"number"`
	RegisteredSuccessfully  null.Bool   `json:"registered_successfully"`
	RegistrationCode        null.String `json:"registration_code"`
	RegistrationCodeSentAt  null.Time   `json:"registration_code_sent_at"`
	LastCheckedCallStatus   null.Time   `json:"last_checked_call_status"`
	MobileCode              null.String `json:"mobile_code"`
	Sha256Hash              null.String `json:"sha256_code"`
	MobileConfirmed         null.Bool   `json:"mobile_confirmed"`
	Location                null.String `json:"location"`
	AgentDating             null.Bool   `json:"agent_dating"`
	DatingPrefAgeRangeStart null.Int    `json:"dating_pref_age_range_start"`
	DatingPrefAgeRangeEnd   null.Int    `json:"dating_pref_age_range_end"`
	AgentDeployed           null.Bool   `json:"agent_deployed"`
	SelectedIntroID         null.String `json:"selected_intro_id"`
	UserInviteCodeID        null.String `json:"user_invite_code_id"`
	HasTranscript           null.Bool   `json:"has_transcript"`
	LatestTranscriptID      null.String `json:"latest_transcript_id"`
	LatestTranscriptTS      null.Time   `json:"latest_transcript_ts"`
	Sexuality               null.String `json:"sexuality"` // String enum
	SexualityIsVisible      null.Bool   `json:"sexuality_is_visible"`
	UserType                null.String `json:"user_type"` // String enum
}

// UpsertUser upserts a user in the database.
func (s *Store) UpsertUser(ctx context.Context, upsert *UpsertUser, db boil.ContextExecutor) error {
	user := pgmodel.User{}
	cols := make([]string, 0)

	// registration fields
	if upsert.RegistrationCode.Valid {
		user.RegistrationCode = null.StringFrom(upsert.RegistrationCode.String)
	}
	if upsert.RegistrationCodeSentAt.Valid {
		user.RegistrationCodeSentAt = null.TimeFrom(upsert.RegistrationCodeSentAt.Time)
	}
	if upsert.RegisteredSuccessfully.Valid {
		user.RegisteredSuccessfully = null.BoolFrom(upsert.RegisteredSuccessfully.Bool)
	}
	if upsert.MobileNumber.Valid {
		user.MobileNumber = null.StringFrom(upsert.MobileNumber.String)
	}

	// basic fields
	if upsert.Email.Valid {
		user.Email = upsert.Email.String
		cols = append(cols, pgmodel.UserColumns.Email)
	}
	if upsert.SupabaseID.Valid {
		user.SupabaseID = null.StringFrom(upsert.SupabaseID.String)
	}
	if upsert.FirstName.Valid {
		user.FirstName = null.StringFrom(upsert.FirstName.String)
	}
	if upsert.LastName.Valid {
		user.LastName = null.StringFrom(upsert.LastName.String)
	}
	if upsert.Birthday.Valid {
		user.Birthday = null.TimeFrom(upsert.Birthday.Time)
	}
	if upsert.Gender.Valid {
		user.Gender = null.StringFrom(upsert.Gender.String)
	}
	if upsert.Height.Valid {
		user.HeightCM = null.IntFrom(upsert.Height.Int)
	}
	if upsert.MobileNumber.Valid {
		user.MobileNumber = null.StringFrom(upsert.MobileNumber.String)
	}
	if upsert.MobileCode.Valid {
		user.MobileCode = null.StringFrom(upsert.MobileCode.String)
	}
	if upsert.MobileConfirmed.Valid {
		user.MobileConfirmed = null.BoolFrom(upsert.MobileConfirmed.Bool)
	}

	if err := user.Upsert(ctx, db, true, cols, boil.Infer(), boil.Infer()); err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}

	return nil
}

// UpdateUser updates user details in the database.
func (s *Store) UpdateUser(ctx context.Context, u *UpdateUser, db boil.ContextExecutor) error {
	user, err := pgmodel.FindUser(ctx, db, u.ID)
	if err != nil {
		return fmt.Errorf("find user by UserID: %w", err)
	}

	// registration fields
	if u.Email.Valid {
		user.Email = u.Email.String
	}
	if u.RegistrationCode.Valid {
		user.RegistrationCode = null.StringFrom(u.RegistrationCode.String)
	}
	if u.MobileCode.Valid {
		user.MobileCode = null.StringFrom(u.MobileCode.String)
	}
	if u.MobileConfirmed.Valid {
		user.MobileConfirmed = null.BoolFrom(u.MobileConfirmed.Bool)
	}
	if u.RegisteredSuccessfully.Valid {
		user.RegisteredSuccessfully = null.BoolFrom(u.RegisteredSuccessfully.Bool)
	}
	if u.RegistrationCodeSentAt.Valid {
		user.RegistrationCodeSentAt = null.TimeFrom(u.RegistrationCodeSentAt.Time)
	}
	if u.LastCheckedCallStatus.Valid {
		user.LastCheckedCallStatus = null.TimeFrom(u.LastCheckedCallStatus.Time)
	}
	if u.Sha256Hash.Valid {
		user.Sha256Hash = u.Sha256Hash
	}

	// basic fields
	if u.Email.Valid {
		user.Email = u.Email.String
	}
	if u.FirstName.Valid {
		user.FirstName = u.FirstName
	}
	if u.LastName.Valid {
		user.LastName = u.LastName
	}
	if u.Birthday.Valid {
		user.Birthday = u.Birthday
	}
	if u.Gender.Valid {
		user.Gender = u.Gender
	}
	if u.Height.Valid {
		user.HeightCM = u.Height
	}
	if u.Number.Valid {
		user.MobileNumber = u.Number
	}
	if u.Location.Valid {
		user.Location = u.Location
	}
	if u.AgentDating.Valid {
		user.AgentDating = u.AgentDating
	}
	if u.DatingPrefAgeRangeStart.Valid {
		user.DatingPrefAgeRangeStart = u.DatingPrefAgeRangeStart
	}
	if u.DatingPrefAgeRangeEnd.Valid {
		user.DatingPrefAgeRangeEnd = u.DatingPrefAgeRangeEnd
	}
	if u.AgentDeployed.Valid {
		user.AgentDeployed = u.AgentDeployed
	}
	if u.SelectedIntroID.Valid {
		user.SelectedIntroID = u.SelectedIntroID
	}
	if u.UserInviteCodeID.Valid {
		user.UserInviteCodeRefID = u.UserInviteCodeID
	}
	if u.HasTranscript.Valid {
		user.HasTranscript = u.HasTranscript.Bool
	}
	if u.LatestTranscriptID.Valid {
		user.LatestTranscriptID = u.LatestTranscriptID
	}
	if u.LatestTranscriptTS.Valid {
		user.LatestTranscriptTS = u.LatestTranscriptTS
	}
	if u.Sexuality.Valid {
		user.Sexuality = u.Sexuality
	}
	if u.SexualityIsVisible.Valid {
		user.SexualityIsVisible = u.SexualityIsVisible
	}
	if u.UserType.Valid {
		user.UserType = u.UserType.String
	}

	if _, err := user.Update(ctx, db, boil.Infer()); err != nil {
		return fmt.Errorf("u user: %w", err)
	}

	return nil
}

func qModUsers(f *QueryFilterUser) []qm.QueryMod {
	filters := make([]qm.QueryMod, 0)
	if f.ID.Valid {
		filters = append(filters, pgmodel.UserWhere.ID.EQ(f.ID.String))
	}
	if f.Email.Valid {
		filters = append(filters, pgmodel.UserWhere.Email.EQ(f.Email.String))
	}
	if f.Number.Valid {
		filters = append(filters, pgmodel.UserWhere.MobileNumber.EQ(f.Number))
	}
	if f.Sha256Hash.Valid {
		filters = append(filters, pgmodel.UserWhere.Sha256Hash.EQ(f.Sha256Hash))
	}
	return filters
}

type QueryFilterUser struct {
	ID         null.String `json:"id"`
	Email      null.String `json:"email"`
	Number     null.String `json:"mobile_number"`
	Sha256Hash null.String `json:"sha256_hash"`
}

func (s *Store) Users(ctx context.Context,
	exec boil.ContextExecutor,
	filter *QueryFilterUser,
) (pgmodel.UserSlice, error) {
	users, err := pgmodel.Users(qModUsers(filter)...).All(ctx, exec)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgmodel.UserSlice{}, nil // no results found
		}
		return nil, fmt.Errorf("query users: %w", err)
	}

	return users, nil
}

// User retrieves a single user from the database
// based on the provided filter.
func (s *Store) User(ctx context.Context,
	exec boil.ContextExecutor,
	filter *QueryFilterUser,
) (*pgmodel.User, error) {
	users, err := s.Users(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	if len(users) == 0 {
		return &pgmodel.User{}, nil
	}

	if len(users) != 1 {
		return nil, fmt.Errorf("user count mismatch, have %d, want 1", len(users))
	}

	return users[0], nil
}
