package setting

import (
	"errors"
)

var (
	ErrBelowMinAgeRange            = errors.New("below minimum age range")
	ErrDatePrefAgeEndGtStart       = errors.New("dating pref age: end greater than start")
	ErrUnknownSysParam             = errors.New("unknown system parameter")
	ErrBothSettersDefined          = errors.New("both int and string setters defined")
	ErrInviteUserCodeAlreadyExists = errors.New("code already exists")
	ErrInviteUserAlreadyAMember    = errors.New("user is already a member")
	ErrInviteUserAlreadyInvited    = errors.New("user is already invited")
)
