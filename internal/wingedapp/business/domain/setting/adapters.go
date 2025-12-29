package setting

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
)

type sysParamStorer interface {
	SysParams(ctx context.Context, exec boil.ContextExecutor) ([]SysParam, error)
}

type personalInfoStorer interface {
	PersonalInformation(ctx context.Context, exec boil.ContextExecutor, userID string) (*PersonalInformation, error)
	UpdatePersonalInformation(ctx context.Context, exec boil.ContextExecutor, updater *UpdatePersonalInformation) error
}

type anonymizedContactsStorer interface {
	AnonymizedContacts(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterAnonymizedContact) ([]AnonymizedContact, error)
	UpsertAnonymizedContacts(ctx context.Context, exec boil.ContextExecutor, inserter []UpsertAnonymizedContact) ([]UpsertAnonymizedContact, error)
}

// userInviteCodeStorer is an interface for managing user invite codes.
type userInviteCodeStorer interface {
	UserInviteCodes(ctx context.Context, exec boil.ContextExecutor, filter *QueryFilterUserInviteCode) ([]UserInviteCode, error)

	// TODO: add response type
	InsertUserInviteCode(ctx context.Context, exec boil.ContextExecutor, inserter *InsertUserInviteCode) error
}

type userStorer interface {
	Users(ctx context.Context, exec boil.ContextExecutor, filter *QueryFilterUser) ([]User, error)
}

type storer interface {
	userStorer
	personalInfoStorer
	sysParamStorer
	anonymizedContactsStorer
	userInviteCodeStorer
}

// transactor is an interface for handling transactions.
type transactor interface {
	TX() (boil.ContextTransactor, error)
	Rollback(boil.ContextTransactor)
	DB() boil.ContextExecutor
}
