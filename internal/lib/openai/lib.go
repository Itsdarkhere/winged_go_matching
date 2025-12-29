package openai

import (
	"context"
	"fmt"
	"time"
	"wingedapp/pgtester/internal/util/validationlib"

	openaiapi "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

/* OpenAI Lib */

type Lib struct {
	api *openaiapi.Client
	key string // Stored for Responses API (direct HTTP calls)
}

type Config struct {
	APIKey string `json:"api_key" validate:"required"`
}

func (c *Config) Validate() error {
	return validationlib.Validate(c)
}

func NewLib(cfg *Config) (*Lib, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	client := openaiapi.NewClient(option.WithAPIKey(cfg.APIKey))
	return &Lib{api: &client, key: cfg.APIKey}, nil
}

// Prompt sends a prompt to OpenAI and returns the response
func (l *Lib) Prompt(ctx context.Context, opts *PromptOpts) (*PromptResp, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	promptMessage := opts.Message
	if opts.AdditionalContext != "" {
		promptMessage = fmt.Sprintf("%s\n\nAdditional Context:\n%s", opts.Message, opts.AdditionalContext)
	}

	params := openaiapi.ChatCompletionNewParams{
		Model: openaiapi.ChatModelGPT4o,
		Messages: []openaiapi.ChatCompletionMessageParamUnion{
			openaiapi.UserMessage(promptMessage),
		},
	}

	res, err := l.api.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("chat.completions.new: %w", err)
	}
	if res == nil || len(res.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	content := res.Choices[0].Message.Content
	if content == "" {
		return nil, fmt.Errorf("empty response content from OpenAI")
	}

	return rawRespToPromptResp(opts, res), nil
}

func rawRespToPromptResp(opts *PromptOpts, resp *openaiapi.ChatCompletion) *PromptResp {
	return &PromptResp{
		ID:                resp.ID,
		SentMessage:       opts.Message,
		AdditionalContext: opts.AdditionalContext,
		Response:          resp.Choices[0].Message.Content,
		CreatedAt:         time.Unix(resp.Created, 0),
	}
}
