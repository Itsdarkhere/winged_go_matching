package agentlog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wingedapp/pgtester/internal/wingedapp/business/sdk"
	agentLogLib "wingedapp/pgtester/internal/wingedapp/lib/agentlog"

	"github.com/aarondl/null/v8"
	"github.com/google/uuid"
)

type Business struct {
	transactor       transactor
	agentLogGetter   agentLogGetter
	agentLogInserter agentLogInserter
}

func NewBusiness(
	transactor transactor,
	agentLogGetter agentLogGetter,
	agentLogInserter agentLogInserter,
) (*Business, error) {
	if transactor == nil {
		return nil, errors.New("transactor is required")
	}
	if agentLogGetter == nil {
		return nil, errors.New("agentLogGetter is required")
	}
	if agentLogInserter == nil {
		return nil, errors.New("agentLogInserter is required")
	}

	return &Business{
		transactor:       transactor,
		agentLogGetter:   agentLogGetter,
		agentLogInserter: agentLogInserter,
	}, nil
}

// AgentLogs returns unseen agent logs for a user and marks them as seen.
func (b *Business) AgentLogs(ctx context.Context, userID uuid.UUID, pagination *sdk.Pagination) (*agentLogLib.AgentLogPaginated, error) {
	exec := b.transactor.DB()
	result, err := b.agentLogGetter.AgentLogs(ctx, exec, userID, pagination)
	if err != nil {
		return nil, fmt.Errorf("get agent logs: %w", err)
	}
	return result, nil
}

// AgentLog returns a single agent log by ID for the user and marks it as seen.
func (b *Business) AgentLog(ctx context.Context, userID uuid.UUID, logID uuid.UUID) (*agentLogLib.AgentLog, error) {
	exec := b.transactor.DB()
	result, err := b.agentLogGetter.AgentLog(ctx, exec, userID, logID)
	if err != nil {
		return nil, fmt.Errorf("get agent log: %w", err)
	}
	return result, nil
}

// AllLogs returns all agent logs for a user (seen and unseen).
func (b *Business) AllLogs(ctx context.Context, userID uuid.UUID, pagination *sdk.Pagination) (*agentLogLib.AgentLogPaginated, error) {
	exec := b.transactor.DB()
	result, err := b.agentLogGetter.AllLogs(ctx, exec, userID, pagination)
	if err != nil {
		return nil, fmt.Errorf("get all agent logs: %w", err)
	}
	return result, nil
}

// UnseenCount returns the count of unseen displayable logs for a user.
func (b *Business) UnseenCount(ctx context.Context, userID uuid.UUID) (int, error) {
	exec := b.transactor.DB()
	count, err := b.agentLogGetter.UnseenCount(ctx, exec, userID)
	if err != nil {
		return 0, fmt.Errorf("get unseen count: %w", err)
	}
	return count, nil
}

// Notify creates an agent log for a user with optional delay.
// This is the simplified entry point for other modules to inject logs.
func (b *Business) Notify(ctx context.Context, userID string, message string, delay time.Duration) error {
	displayBy := time.Now()
	if delay > 0 {
		displayBy = displayBy.Add(delay)
	}

	_, err := b.agentLogInserter.Insert(ctx, b.transactor.DB(), &agentLogLib.InsertAgentLog{
		UserRefID: userID,
		Log:       message,
		DisplayBy: null.TimeFrom(displayBy),
	})
	if err != nil {
		return fmt.Errorf("insert agent log: %w", err)
	}

	return nil
}
