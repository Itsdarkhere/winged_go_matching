package openai

import (
	"time"
	"wingedapp/pgtester/internal/util/validationlib"
)

// PromptOpts is the options for prompting OpenAI.
type PromptOpts struct {
	Message           string `json:"message" validate:"required"`
	AdditionalContext string `json:"additional_context"`
}

func (p *PromptOpts) Validate() error {
	return validationlib.Validate(p)
}

// PromptResp is the response from OpenAI.
type PromptResp struct {
	ID                string    `json:"id"`
	SentMessage       string    `json:"sent_message"`
	AdditionalContext string    `json:"additional_context"`
	Response          string    `json:"response"`
	CreatedAt         time.Time `json:"created_at"`
}
