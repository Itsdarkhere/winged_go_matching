package registration

import (
	"context"
	"wingedapp/pgtester/internal/wingedapp/lib/economy"
	"wingedapp/pgtester/internal/wingedapp/sysparam"

	"github.com/aarondl/sqlboiler/v4/boil"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// texter is an interface for sending text messages.
//
//counterfeiter:generate . texter
type texter interface {
	SendMessage(ctx context.Context, to string, msg string) error
}

// settingGetter is an interface for getting text message settings.
//
//counterfeiter:generate . settingGetter
type settingGetter interface {
	Settings(ctx context.Context) (*sysparam.Settings, error)
}

// uploader is an interface for uploading files.
// main use-case is for uploading user profile pictures.
//
//counterfeiter:generate . uploader
type uploader interface {
	Upload(ctx context.Context, key string, file []byte) (*FileDetails, error)
	PublicURL(ctx context.Context, key string) (string, error)
}

// beUploader is an interface for uploading files.
// main use-case is for uploading user profile pictures.
//
//counterfeiter:generate . beUploader
type beUploader interface {
	uploader
}

// aiUploader is an interface for uploading files.
// main use-case is for uploading user profile pictures.
//
//counterfeiter:generate . aiUploader
type aiUploader interface {
	uploader
}

// sysParamStorer is an interface for managing system parameters.
//
// counterfeiter:generate . sysParamStorer
type sysParamStorer interface {
	SysParam(ctx context.Context, exec boil.ContextExecutor, key string) (string, error)
}

// userStorer is an interface for managing user data.
type userStorer interface {
	Users(ctx context.Context, execBE boil.ContextExecutor, execAI boil.ContextExecutor, filter *QueryFilterUser) ([]User, error)
	User(ctx context.Context, execBE boil.ContextExecutor, execAI boil.ContextExecutor, filter *QueryFilterUser) (*User, error)
	UpdateUser(ctx context.Context, txBE boil.ContextTransactor, execAI boil.ContextExecutor, user *UpdateUser) (*User, error)
	UpsertUser(ctx context.Context, tx boil.ContextTransactor, user *UpsertUser) (*User, error)
	CountUser(ctx context.Context, exec boil.ContextExecutor, filter *QueryFilterUser) (int64, error)
}

// userInviteCodeStorer is an interface for managing user invite codes.
type userInviteCodeStorer interface {
	UserInviteCode(ctx context.Context, exec boil.ContextExecutor, filter *UserInviteCodeQueryFilter) (*UserInviteCode, error)
	DeleteUserInviteCode(ctx context.Context, exec boil.ContextExecutor, id string) (int64, error)
	UpdateUserInviteCode(ctx context.Context, exec boil.ContextExecutor, params *UpdateUserInviteCode) error
}

// userElevenLabsStorer is an interface for managing user ElevenLabs data.
type userElevenLabsStorer interface {
	InsertUserElevenLabs(ctx context.Context, db boil.ContextExecutor, params *InsertUserElevenLabs) error
}

// userContactsStorer is an interface for managing user contacts such as blocked contacts.
type userContactsStorer interface {
	InsertUserBlockedContact(ctx context.Context, db boil.ContextExecutor, params *InsertUserBlockedContact) error
	UserBlockedContacts(ctx context.Context, exec boil.ContextExecutor, filter *UserBlockedContactQueryFilter) ([]UserBlockedContact, error)
	UserBlockedContact(ctx context.Context, exec boil.ContextExecutor, filter *UserBlockedContactQueryFilter) (*UserBlockedContact, error)
	UserUnblockContact(ctx context.Context, exec boil.ContextExecutor, filter *UserBlockedContactQueryFilter) error
	UserUnblockAll(ctx context.Context, db boil.ContextExecutor, ids []string) error
}

// userPhotoStorer is an interface for managing user photos.
type userPhotoStorer interface {
	InsertUserPhoto(ctx context.Context, exec boil.ContextExecutor, userID, bucket, key string, orderNo int) (*UserPhoto, error)
	UserPhoto(ctx context.Context, exec boil.ContextExecutor, filter *UserPhotoQueryFilter) (*UserPhoto, error)
	UserPhotos(ctx context.Context, exec boil.ContextExecutor, filter *UserPhotoQueryFilter) ([]UserPhoto, error)
	DeletePhoto(ctx context.Context, exec boil.ContextExecutor, id string) error
}

// profilerStorer is an interface for managing user profile.
type profilerStorer interface {
	Profile(ctx context.Context, exec boil.ContextExecutor, filter *ProfileQueryFilter) (*Profile, error)
	Profiles(ctx context.Context, exec boil.ContextExecutor) ([]Profile, error)
}

// audioFileStorer is an interface for managing user audio file.
type audioFileStorer interface {
	AudioFile(ctx context.Context, exec boil.ContextExecutor, filter *AudioFileQueryFilter) (*UserAudio, error)
	AudioFiles(ctx context.Context, exec boil.ContextExecutor, filter *AudioFileQueryFilter) ([]UserAudio, error)
}

// transcriptStorer is an interface for managing user transcript.
type transcriptStorer interface {
	Transcript(ctx context.Context, exec boil.ContextExecutor, filter *TranscriptQueryFilters) (*UserTranscript, error)
	Transcripts(ctx context.Context, exec boil.ContextExecutor, filter *TranscriptQueryFilters) ([]UserTranscript, error)
}

// voiceStorer
type voiceStorer interface {
	Voices(ctx context.Context, exec boil.ContextExecutor, filter *VoiceQueryFilter) ([]Voice, error)
	Voice(ctx context.Context, exec boil.ContextExecutor, filter *VoiceQueryFilter) (*Voice, error)
}

type deleter interface {
	DeleteUserData(ctx context.Context,
		beExec boil.ContextExecutor,
		aiExec boil.ContextExecutor,
		supabaseExec boil.ContextExecutor,
		userID string,
	) error
}

// storer is an interface for user data storage operations.
type storer interface {
	userStorer
	userInviteCodeStorer
	userElevenLabsStorer
	userContactsStorer
	userPhotoStorer
	profilerStorer
	audioFileStorer
	transcriptStorer
	sysParamStorer
}

// transactor is an interface for handling transactions.
type transactor interface {
	TX() (boil.ContextTransactor, error)
	Rollback(boil.ContextTransactor)
	DB() boil.ContextExecutor
}

// actionLogger is an interface for economy actions (referral bonuses).
type actionLogger interface {
	CreateActionLog(ctx context.Context, exec boil.ContextExecutor, inserter *economy.InsertActionLog) error
}
