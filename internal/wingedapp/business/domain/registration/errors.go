package registration

import "errors"

var (
	ErrUserNotFound                   = errors.New("no user found")
	ErrTranscriptsEmpty               = errors.New("transcripts empty")
	ErrUserMaxPhotoCountExceeded      = errors.New("user max photo count exceeded")
	ErrUserPhoto                      = errors.New("user max photo count exceeded")
	ErrUserPhotoInvalidOrder          = errors.New("user photo invalid order")
	ErrUserPhotoAlreadyExists         = errors.New("user photo already exists")
	ErrUserPhotoNotFound              = errors.New("no user photo found")
	ErrRegistrationCodeNotFound       = errors.New("no registration code found")
	ErrUserUnregistered               = errors.New("user unregistered")
	ErrUserAlreadyRegistered          = errors.New("user already registered")
	ErrUserHasNoValidRegistrationCode = errors.New("user has no valid registration code")
	ErrUserRegistrationCodeMismatch   = errors.New("user registration code mismatch")
	ErrProfileNotFound                = errors.New("no user profile found")
	ErrUserAudioFileNotFound          = errors.New("no user audio file found")
	ErrUserTranscriptNotFound         = errors.New("no transcript found")
	ErrUserNoPhotos                   = errors.New("no user photos found")
	ErrUserNoSelectedIntroID          = errors.New("no user selected voice id")
	ErrSysParamIntLessThanOne         = errors.New("sys param less than one")
	ErrSysParamShouldBeInt            = errors.New("sys param should be int")
	ErrUserInviteCodeUsageExceeded    = errors.New("user invite code usage exceeded")
	ErrUserInviteCodeExpired          = errors.New("user invite code expired")
	ErrUserInviteExclusive            = errors.New("user invite code exclusive")
)
