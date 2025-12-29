package youragent

import (
	"context"

	"wingedapp/pgtester/internal/wingedapp/lib/economy"

	"github.com/aarondl/sqlboiler/v4/boil"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// userAIConvoStorer is the minimal interface needed to persist conversations.
//
//counterfeiter:generate . userAIConvoStorer
type userAIConvoStorer interface {
	InsertUserAIConvo(ctx context.Context, db boil.ContextExecutor, inserter *InsertUserAIConvo) (*UserAIConvo, error)
	UserAIConvos(ctx context.Context, db boil.ContextExecutor, f *UserAIConvoQueryFilter) ([]UserAIConvo, error)
}

// generalAIContextStorer is the minimal interface needed to persist conversations.
//
// counterfeiter:generate . generalAIContextStorer
type generalAIContextStorer interface {
	InsertUserAIConvo(ctx context.Context, db boil.ContextExecutor, inserter *InsertUserAIConvo) (*UserAIConvo, error)
	GeneralAIContexts(ctx context.Context, db boil.ContextExecutor, f *GeneralAIContextQueryFilter) ([]GeneralAIContext, error)
}

// prompter is the LLM interface
//
//counterfeiter:generate . prompter
type prompter interface {
	Prompt(ctx context.Context, opts *PromptOpts) (*PromptResp, error)
}

// storer is the minimal interface needed to persist conversations.
//
//counterfeiter:generate . storer
type storer interface {
	userAIConvoStorer
	generalAIContextStorer
}

type transactor interface {
	TX() (boil.ContextTransactor, error)
	Rollback(boil.ContextTransactor)
	DB() boil.ContextExecutor
}

// actionLogger is an interface for economy actions (message spending).
type actionLogger interface {
	CanPerformAction(ctx context.Context, exec boil.ContextExecutor, params *economy.CanPerformActionParams) (bool, error)
	CreateActionLog(ctx context.Context, exec boil.ContextExecutor, inserter *economy.InsertActionLog) error
}
