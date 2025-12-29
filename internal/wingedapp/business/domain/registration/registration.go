package registration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
	"wingedapp/pgtester/internal/lib/twilio"
	"wingedapp/pgtester/internal/util/validationlib"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/userhasher"
	"wingedapp/pgtester/internal/wingedapp/sysparam"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
)

type Config struct {
	// UserMaxPhotos is the maximum number of photos a user can upload.
	UserMaxPhotos int `json:"user_max_photos" validate:"required"`

	// LastCallThresholdIntervalSecs is the threshold in minutes to update the last call check.
	// This is to check if certain minutes have "elapsed" since the last call status check, and be able to
	// make a decision if we should retry having a call — or not.
	LastCallThresholdIntervalSecs int `json:"last_call_threshold_interval_secs" validate:"required"`
}

func (c *Config) Validate() error {
	return validationlib.Validate(c)
}

type Business struct {
	logger         applog.Logger  // Logrus
	cfg            *Config        // Config
	texter         texter         // Twilio
	transBE        transactor     // PG tx
	transAI        transactor     // PG tx
	transSupa      transactor     // PG tx (supabase auth)
	storer         storer         // PG repo
	settingGetter  settingGetter  // System parameter getter
	beUploader     beUploader     // Supabase object storage
	aiUploader     aiUploader     // Supabase object storage
	sysParamStorer sysParamStorer // System parameter storer
	deleter        deleter        // Deletes user data
	actionLogger   actionLogger   // Economy action logger for referrals
}

func NewBusiness(
	logger applog.Logger,
	cfg *Config,
	texter texter,
	transBE transactor,
	transAI transactor,
	transSupa transactor,
	storer storer,
	settingGetter settingGetter,
	beUploader uploader,
	aiUploader uploader,
	deleter deleter,
) (*Business, error) {
	if logger == nil {
		return nil, errors.New("nil logger")
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	if texter == nil {
		return nil, errors.New("nil texter")
	}
	if storer == nil {
		return nil, errors.New("nil storer")
	}
	if settingGetter == nil {
		return nil, errors.New("nil settingGetter")
	}
	if transBE == nil {
		return nil, errors.New("nil transBE")
	}
	if transAI == nil {
		return nil, errors.New("nil transAI")
	}
	if transSupa == nil {
		return nil, errors.New("nil transSupa")
	}
	if beUploader == nil {
		return nil, errors.New("nil backend Uploader")
	}
	if aiUploader == nil {
		return nil, errors.New("nil ai Uploader")
	}
	if deleter == nil {
		return nil, errors.New("nil deleter")
	}

	biz := &Business{
		logger:        logger,
		cfg:           cfg,
		texter:        texter,
		storer:        storer,
		settingGetter: settingGetter,
		transBE:       transBE,
		transAI:       transAI,
		transSupa:     transSupa,
		beUploader:    beUploader,
		aiUploader:    aiUploader,
		deleter:       deleter,
	}

	return biz, nil
}

// SetActionLogger sets the action logger for referral bonuses.
// Use setter pattern to avoid breaking existing constructor.
func (b *Business) SetActionLogger(al actionLogger) {
	b.actionLogger = al
}

// updateTestAccountMobileCode updates the mobile code for a test account user.
func (b *Business) updateTestAccountMobileCode(ctx context.Context, user *User, code string) error {
	tx, err := b.transBE.TX()
	if err != nil {
		return fmt.Errorf("tx: %w", err)
	}
	defer b.transBE.Rollback(tx)

	if _, err := b.storer.UpdateUser(ctx, tx, b.dbAI(), &UpdateUser{
		ID:         user.ID,
		MobileCode: null.StringFrom(code),
	}); err != nil {
		return fmt.Errorf("update test user mobile confirmation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit test account mobile code update tx: %w", err)
	}

	return nil
}

// ConfirmMobile confirms a mobile number for a user, by checking against the store.
// func (b *Business) ConfirmMobile(ctx context.Context, userID, mobileCode string) error {
func (b *Business) ConfirmMobile(ctx context.Context, user *User, mobileCode string) error {
	// TODO: remove test account hack later
	const (
		testEmail = "mikaela@wingedapp.com"
		testCode  = "123456"
	)
	if user.Email == testEmail {
		if err := b.updateTestAccountMobileCode(ctx, user, testCode); err != nil {
			return fmt.Errorf("create test account mobile code: %w", err)
		}
	}

	dbBE := b.transBE.DB()
	dbAI := b.transAI.DB()

	_, err := b.storer.User(ctx, dbBE, dbAI,
		&QueryFilterUser{
			ID:         null.StringFrom(user.ID),
			MobileCode: null.StringFrom(mobileCode),
		})
	if err != nil {
		return fmt.Errorf("get mobile code: %w", err)
	}

	tx, err := b.transBE.TX()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer b.transBE.Rollback(tx)

	hash := userhasher.Sha256(user.MobileNumber.String)
	updateParams := &UpdateUser{
		ID:              user.ID,
		Sha256Hash:      null.StringFrom(hash),
		MobileConfirmed: null.BoolFrom(true),
	}

	// mark confirmed if found
	_, err = b.storer.UpdateUser(ctx,
		tx,
		b.dbAI(),
		updateParams,
	)
	if err != nil {
		return fmt.Errorf("update user mobile confirmed true: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// HandleUserBasicInfoUpdate updates the user's basic information
// and sends a verification code via SMS.
func (b *Business) HandleUserBasicInfoUpdate(ctx context.Context, userBasicInfo *UpdateUser) (*User, error) {
	tx, err := b.transBE.TX()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer b.transBE.Rollback(tx)

	user, err := b.storer.UpdateUser(ctx, tx, b.dbAI(), userBasicInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// send a verification after successful update
	code := twilio.RandomDigitCode(RandomMobileCodeLength)
	welcomeMsg := welcomeMsgTmpl(code)
	if err := b.texter.SendMessage(ctx, userBasicInfo.Number.String, welcomeMsg); err != nil {
		return nil, fmt.Errorf("failed to send verification welcomeMsg: %w", err)
	}

	// Update the user mobile num with the registration welcomeMsg
	_, err = b.storer.UpdateUser(ctx, tx, b.dbAI(), &UpdateUser{
		ID:         user.ID,
		MobileCode: null.StringFrom(code),
	})
	if err != nil {
		return nil, fmt.Errorf("update user mobile confirmation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return user, nil
}

func (b *Business) UpdateUserBasicInfoAndSendConfirmMessage(ctx context.Context, userBasicInfo *UpdateUser) (*User, error) {
	// Do we care about validation much here? lol
	return nil, nil
}

// EnterUserCode handles entering a verification code for user registration or login.
func (b *Business) EnterUserCode(ctx context.Context, userUUID uuid.UUID) (*User, error) {
	return nil, nil
}

// upsertUser inserts or updates a user in the local database based on Supabase user data.
// Will add more fields as needed, such as name, birthday, etc. (can scale)
func (b *Business) upsertUser(ctx context.Context, oauthUser *OauthUser) (*User, error) {
	upsertParams := &UpsertUser{
		SupabaseID: null.StringFrom(oauthUser.SupabaseID),
		Email:      null.StringFrom(oauthUser.Email),
	}
	if oauthUser.RegistrationCode.Valid {
		upsertParams.RegistrationCode = null.StringFrom(oauthUser.RegistrationCode.String)
	}

	tx, err := b.transBE.TX()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer b.transBE.Rollback(tx)

	user, err := b.storer.UpsertUser(ctx, tx, upsertParams)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert user: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return user, nil
}

func newQueryFilterEmail(email string) *QueryFilterUser {
	return &QueryFilterUser{
		Email: null.String{
			String: email,
			Valid:  true,
		},
	}
}

func welcomeMsgTmpl(code string) string {
	return fmt.Sprintf(tmplSendCode, code)
}

// UserFromOauthProvider retrieves a user from the OAuth provider.
func (b *Business) UserFromOauthProvider(ctx context.Context, accessToken string) (*OauthUser, error) {
	return nil, nil
}

// dbBE is shorthand for backend app db executor to
// make it less verbose for calls.
func (b *Business) dbBE() boil.ContextExecutor {
	return b.transBE.DB()
}

// dbAI is shorthand for backend app db executor to
// make it less verbose for calls.
func (b *Business) dbAI() boil.ContextExecutor {
	return b.transAI.DB()
}

// dbSupa is shorthand for supabase auth db executor.
func (b *Business) dbSupa() boil.ContextExecutor {
	return b.transSupa.DB()
}

// AuthenticateLocalUser checks if there's a confirmed user given the email.
func (b *Business) AuthenticateLocalUser(ctx context.Context, email string) (*User, error) {
	filter := newQueryFilterEmail(email)
	filter.EnrichPhotos = true

	user, err := b.storer.User(ctx, b.dbBE(), b.dbAI(), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (b *Business) HandleUpdateUser(ctx context.Context, user *UpdateUser) error {
	// Update the user... do the model plumbing to users lol
	// Then ...
	// Message a text message to the user with a verification code.
	// add a new field

	return nil
}

// userInviteCodeMaxUsage retrieves the maximum usage for a user invite code from system parameters.
func (b *Business) userInviteCodeMaxUsage(ctx context.Context, exec boil.ContextExecutor) (int, error) {
	sysParamVal, err := b.storer.SysParam(ctx, exec, sysparam.UserInviteCodeMaxUsage)
	if err != nil {
		return 0, fmt.Errorf("get sys param %s: %w", sysparam.UserInviteCodeMaxUsage, err)
	}

	maxUsage, err := strconv.Atoi(sysParamVal)
	if err != nil {
		err = fmt.Errorf("atoi sys param %s value %q: %w", sysparam.UserInviteCodeMaxUsage, sysParamVal, err)
		return 0, errors.Join(ErrSysParamShouldBeInt, err)
	}
	if maxUsage < 1 {
		err = fmt.Errorf("sys param %s value %q", sysparam.UserInviteCodeMaxUsage, sysParamVal)
		return 0, errors.Join(ErrSysParamIntLessThanOne, err)
	}

	return maxUsage, nil
}

type enterInviteCodeVars struct {
	currTime       time.Time
	userInviteCode *UserInviteCode
	settings       *sysparam.Settings
	tx             boil.ContextTransactor
}

// loadEnterInviteCodeVars loads necessary variables for entering an invite code logic.
// ensures easier readability for the main business func.
// (explore possible unit of work patterns to fully isolate dep loading, and business logic)
// (make biz logic portable)
func (b *Business) loadEnterInviteCodeVars(ctx context.Context, regCode string) (*enterInviteCodeVars, error) {
	v := &enterInviteCodeVars{}

	// call first to avoid being blocked by tx
	settings, err := b.settingGetter.Settings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}

	tx, err := b.transBE.TX()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	userInviteCode, err := b.storer.UserInviteCode(ctx, tx, &UserInviteCodeQueryFilter{
		Code: null.StringFrom(regCode),
	})
	if err != nil {
		return nil, fmt.Errorf("get invite code: %w", err)
	}

	v.currTime = time.Now()
	v.userInviteCode = userInviteCode
	v.settings = settings
	v.tx = tx

	return v, nil
}

// EnterInviteCode enters a registration code for a user.
func (b *Business) EnterInviteCode(ctx context.Context, user *User, regCode string) error {
	if user.RegisteredSuccessfully.Bool {
		return ErrUserAlreadyRegistered // quick guard
	}

	v, err := b.loadEnterInviteCodeVars(ctx, regCode)
	if err != nil {
		return fmt.Errorf("load enter invite code v: %w", err)
	}
	defer b.transBE.Rollback(v.tx)

	// check max usage
	if v.userInviteCode.UsageCount >= v.settings.UserInviteCodeMaxUsage {
		return ErrUserInviteCodeUsageExceeded
	}

	if v.userInviteCode.Category == referralUserInviteCode {
		// check expired
		daysElapsed := math.Abs(v.currTime.Sub(v.userInviteCode.CreatedAt).Hours() / 24)
		if daysElapsed < 1 {
			daysElapsed = 0
		}
		if daysElapsed > float64(v.settings.InviteExpiryDays) {
			return ErrUserInviteCodeExpired
		}

		// check matches user number
		if v.userInviteCode.ForNumber != user.MobileNumber.String {
			return ErrUserInviteExclusive
		}
	}

	// update user, and user invite code
	user, err = b.storer.UpdateUser(ctx, v.tx, b.dbAI(), &UpdateUser{
		ID:                     user.ID,
		RegisteredSuccessfully: null.BoolFrom(true),
		RegistrationCode:       null.StringFrom(regCode),
		RegistrationCodeSentAt: null.TimeFrom(time.Now()),
		UserInviteCodeID:       null.StringFrom(v.userInviteCode.ID),
	})
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// increment user invite code usage
	if err = b.storer.UpdateUserInviteCode(ctx, v.tx, &UpdateUserInviteCode{
		ID:         v.userInviteCode.ID,
		UsageCount: null.IntFrom(v.userInviteCode.UsageCount + 1), // increment
		LastUsed:   null.TimeFrom(v.currTime),
	}); err != nil {
		return fmt.Errorf("delete registration code: %w", err)
	}

	if err = v.tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// AuthenticateOauthUser checks if an oauth user exists in the local database.
// If not, it creates a new user and sends a verification code via SMS.
func (b *Business) AuthenticateOauthUser(ctx context.Context, oAuthUsr *OauthUser) (*User, error) {
	var notFound bool

	user, err := b.storer.User(ctx, b.dbBE(), b.dbAI(), newQueryFilterEmail(oAuthUsr.Email))
	if err != nil {
		if !errors.Is(err, ErrUserNotFound) {
			return nil, fmt.Errorf("failed to query user: %w", err)
		}
		notFound = true
	}

	if notFound { // sync to local db
		user = &User{
			SupabaseID:             null.StringFrom(oAuthUsr.SupabaseID),
			Email:                  oAuthUsr.Email,
			RegisteredSuccessfully: null.BoolFrom(false),
		}

		if err := validationlib.Validate(user); err != nil {
			return nil, fmt.Errorf("validate user: %w", err)
		}

		user, err = b.upsertUser(ctx, oAuthUsr)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert user: %w", err)
		}
	}

	return user, nil
}

// SendCodeToUser enters a verification code for user registration or login.
func (b *Business) SendCodeToUser(ctx context.Context, email string) (*User, error) {
	return nil, nil
}

// UpdateBasicUserDetails updates basic user details such as name, email, etc.
func (b *Business) UpdateBasicUserDetails(ctx context.Context, updated *UpdateUser) error {
	// some table upates here
	return nil
}

// UpdateUserAgentDetails updates user agent details which is collected
// from an external source.
func (b *Business) UpdateUserAgentDetails(ctx context.Context) {
	// jsonB fields? hmmm
}

func (b *Business) AddUserElevenLabsEntry(ctx context.Context, userID string, conversation map[string]any) error {
	bytesConv, err := json.Marshal(conversation)
	if err != nil {
		return fmt.Errorf("marshal conversation: %w", err)
	}

	insert := &InsertUserElevenLabs{
		UserID:       userID,
		Conversation: bytesConv,
	}

	if err := b.storer.InsertUserElevenLabs(ctx, b.transBE.DB(), insert); err != nil {
		return fmt.Errorf("insert user eleven labs: %w", err)
	}

	return nil
}

// AddUserRecording updates clip of user recording.
func (b *Business) AddUserRecording(ctx context.Context) {
	// TODO: next sprint

	// jsonB fields? hmmm
}
