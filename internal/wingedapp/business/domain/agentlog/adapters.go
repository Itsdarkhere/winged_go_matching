package agentlog

import (
	"context"

	"wingedapp/pgtester/internal/wingedapp/business/sdk"
	agentLogLib "wingedapp/pgtester/internal/wingedapp/lib/agentlog"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
)

// transactor is an interface for handling transactions.
type transactor interface {
	TX() (boil.ContextTransactor, error)
	Rollback(boil.ContextTransactor)
	DB() boil.ContextExecutor
}

// agentLogGetter contains methods to get agent logs for users.
type agentLogGetter interface {
	AgentLogs(ctx context.Context, exec boil.ContextExecutor, userID uuid.UUID, pagination *sdk.Pagination) (*agentLogLib.AgentLogPaginated, error)
	AgentLog(ctx context.Context, exec boil.ContextExecutor, userID uuid.UUID, logID uuid.UUID) (*agentLogLib.AgentLog, error)
	AllLogs(ctx context.Context, exec boil.ContextExecutor, userID uuid.UUID, pagination *sdk.Pagination) (*agentLogLib.AgentLogPaginated, error)
	UnseenCount(ctx context.Context, exec boil.ContextExecutor, userID uuid.UUID) (int, error)
}

// agentLogInserter contains methods to insert agent logs.
type agentLogInserter interface {
	Insert(ctx context.Context, exec boil.ContextExecutor, inserter *agentLogLib.InsertAgentLog) (*agentLogLib.AgentLog, error)
}
