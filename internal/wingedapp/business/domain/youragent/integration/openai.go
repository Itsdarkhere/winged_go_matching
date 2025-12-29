package integration

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/lib/openai"
	"wingedapp/pgtester/internal/util/validationlib"
	"wingedapp/pgtester/internal/wingedapp/business/domain/youragent"
)

type Config struct {
	APIKey string `json:"api_key" validate:"required"`
}

func (c *Config) Validate() error {
	return validationlib.Validate(c)
}

type Llm struct {
	lib *openai.Lib
}

func NewOpenAIPrompter(cfg *Config) (*Llm, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	lib, err := openai.NewLib(&openai.Config{
		APIKey: cfg.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("new openai lib: %w", err)
	}

	return &Llm{lib}, nil
}

func (l *Llm) Prompt(ctx context.Context, opts *youragent.PromptOpts) (*youragent.PromptResp, error) {
	resp, err := l.lib.Prompt(ctx, toRepoOpts(opts))
	if err != nil {
		return nil, fmt.Errorf("openai prompt: %w", err)
	}
	return toChatPromptResp(resp), nil
}

func toRepoOpts(opts *youragent.PromptOpts) *openai.PromptOpts {
	return &openai.PromptOpts{
		Message:           opts.Message,
		AdditionalContext: opts.AdditionalContext,
	}
}

func toChatPromptResp(resp *openai.PromptResp) *youragent.PromptResp {
	return &youragent.PromptResp{
		ID:                resp.ID,
		AdditionalContext: resp.AdditionalContext,
		CreatedAt:         resp.CreatedAt,
		SentMessage:       resp.SentMessage,
		Response:          resp.Response,
	}
}
