package economy

import (
	"context"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/sysparam"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// messageStorer enables message CRUD (domain specific).
type messageStorer interface {
	Exists(ctx context.Context, exec boil.ContextExecutor, uuid string) (bool, error)
}

// userStorer enables user queries (general — applies to all entrypoints).
type userStorer interface {
	Exists(ctx context.Context, exec boil.ContextExecutor, uuid string) (bool, error)
	User(ctx context.Context, exec boil.ContextExecutor, filter *QueryFilterUser) (*User, error)
}

// inviteCodeStorer enables invite code queries.
type inviteCodeStorer interface {
	InviteCode(ctx context.Context, exec boil.ContextExecutor, filter *QueryFilterInviteCode) (*InviteCode, error)
}

// subscriptionPlanStorer fetches subscription info.
type subscriptionPlanStorer interface {
	SubscriptionPlan(ctx context.Context, exec boil.ContextExecutor, subscriptionType, subscriptionName string) (*SubscriptionPlan, error)
}

type transactionStorer interface {
	Insert(ctx context.Context, exec boil.ContextExecutor, inserter *InsertTransaction) error
	Transactions(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterTransactions) ([]Transaction, error)
	Transaction(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterTransactions) (*Transaction, error)
	Delete(ctx context.Context, exec boil.ContextExecutor, id string) error
}

// userTotalsStorer enables user totals CRUD.
type userTotalsStorer interface {
	Totals(ctx context.Context, exec boil.ContextExecutor, uuid string) (*UserTotals, error)
	Create(ctx context.Context, exec boil.ContextExecutor, uuid string) (*UserTotals, error)
	Update(ctx context.Context, exec boil.ContextExecutor, updater *UpdateUserTotals) error
}

// actionLogStorer enables user action CRUD.
type actionLogStorer interface {
	Insert(ctx context.Context, exec boil.ContextExecutor, actionLogTypeRefID string, insert *InsertActionLog) (*ActionLog, error)
	ActionLogs(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterActionLog) ([]ActionLog, error)
	ActionLog(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterActionLog) (*ActionLog, error)
	Delete(ctx context.Context, exec boil.ContextExecutor, id string) error
}

// settingGetter is an interface for getting settings.
// Injected via biz layer - settings. Leveraging already existing packages
// in Wingedapp.
type settingGetter interface {
	Settings(ctx context.Context) (*sysparam.Settings, error)
}

// dailyCheckinStorer enables daily check-in transaction CRUD.
type dailyCheckinStorer interface {
	Transactions(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterCheckinTransaction) ([]CheckinTransaction, error)
	Transaction(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterCheckinTransaction) (*CheckinTransaction, error)
	TransactionCount(ctx context.Context, exec boil.ContextExecutor, f *QueryFilterCheckinTransaction) (int, error)
	Insert(ctx context.Context, exec boil.ContextExecutor, inserter *InsertCheckinTransaction) (*pgmodel.WingsEcnTransaction, error)
	Update(ctx context.Context, exec boil.ContextExecutor, updater *UpdateCheckinTransaction) (int, error)
}
