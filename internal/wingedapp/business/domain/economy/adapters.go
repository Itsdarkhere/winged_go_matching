package economy

import (
	"context"

	economyLib "wingedapp/pgtester/internal/wingedapp/lib/economy"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// transactor is an interface for handling transactions.
type transactor interface {
	TX() (boil.ContextTransactor, error)
	Rollback(boil.ContextTransactor)
	DB() boil.ContextExecutor
}

// checkinPerformer handles daily check-in operations.
type checkinPerformer interface {
	PerformCheckin(ctx context.Context, exec boil.ContextExecutor, userID string) (*economyLib.CheckinResult, error)
	GetStatus(ctx context.Context, exec boil.ContextExecutor, userID string) (*economyLib.CheckinStatus, error)
}

// actionLogger handles action log creation (payments, referrals, etc).
type actionLogger interface {
	CreateActionLog(ctx context.Context, exec boil.ContextExecutor, inserter *economyLib.InsertActionLog) error
}
