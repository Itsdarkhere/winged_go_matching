package youragent

import (
	"context"
	"fmt"
	"strings"

	"wingedapp/pgtester/internal/wingedapp/lib/economy"
)

// Prompt sends a message to the AI service and returns the response.
func (b *Business) Prompt(ctx context.Context, userID string, opts *PromptOpts) (*PromptResp, error) {
	// check economy balance
	canPerform, err := b.actionLogger.CanPerformAction(ctx, b.trans.DB(), &economy.CanPerformActionParams{
		UserID:     userID,
		ActionType: economy.ActionSendMessage,
	})
	if err != nil {
		return nil, fmt.Errorf("check can perform action: %w", err)
	}
	if !canPerform {
		return nil, economy.ErrInsufficientWings
	}

	// load guardrails
	chatGuardrails := make([]string, 0)
	gCtxs, err := b.storer.GeneralAIContexts(ctx, b.trans.DB(), &GeneralAIContextQueryFilter{})
	if err != nil {
		return nil, fmt.Errorf("load general AI contexts: %w", err)
	}
	for _, gc := range gCtxs {
		chatGuardrails = append(chatGuardrails, gc.Context)
	}
	opts.AdditionalContext = strings.Join(chatGuardrails, ",")

	// exec llm prompt
	promptResp, err := b.prompter.Prompt(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("prompt: %w", err)
	}

	// save convo
	_, err = b.storer.InsertUserAIConvo(ctx, b.trans.DB(), &InsertUserAIConvo{
		UserID:            userID,
		PromptResponseID:  promptResp.ID,
		Message:           promptResp.SentMessage,
		AdditionalContext: promptResp.AdditionalContext,
		Response:          promptResp.Response,
		CreatedAt:         promptResp.CreatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("save conversation: %w", err)
	}

	// log action for economy (deducts wings at threshold)
	err = b.actionLogger.CreateActionLog(ctx, b.trans.DB(), &economy.InsertActionLog{
		UserID: userID,
		RefID:  promptResp.ID,
		Type:   economy.ActionSendMessage,
	})
	if err != nil {
		return nil, fmt.Errorf("create action log: %w", err)
	}

	return promptResp, nil
}
