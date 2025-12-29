package youragent

import (
	"time"
	"wingedapp/pgtester/internal/wingedapp/business/sdk"

	"github.com/aarondl/null/v8"
)

// PromptOpts is the options for prompting OpenAI.
type PromptOpts struct {
	UserID            string `json:"user_id"`
	Message           string `json:"message"`
	AdditionalContext string `json:"additional_context"`
}

// PromptResp is the response from OpenAI.
type PromptResp struct {
	ID                string    `json:"id"`
	SentMessage       string    `json:"sent_message"`
	AdditionalContext string    `json:"additional_context,omitempty"`
	Response          string    `json:"response"`
	CreatedAt         time.Time `json:"created_at"`
}

type InsertConversation struct {
	UserID   string `json:"user_id"`
	Category string `json:"category_ref_id"`
	Convo    string `json:"convo"`
}

type UserAIConvoQueryFilter struct {
	ID       null.String `json:"id"`
	UserID   null.String `json:"user_id"`
	Category null.String `json:"category"`
	Convo    null.String `json:"convo"`
	OrderBy  null.String `json:"order_by"`
	Sort     null.String `json:"sort"`

	Pagination *sdk.Pagination `json:"pagination"`
}

// InsertUserAIConvo represents the data needed to insert a user AI conversation entry.
type InsertUserAIConvo struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	PromptResponseID  string    `json:"prompt_response_id"`
	Message           string    `json:"content"`
	AdditionalContext string    `json:"additional_context"`
	Response          string    `json:"response"`
	CreatedAt         time.Time `json:"created_at"`
}

// UserAIConvo represents a user AI conversation entry.
type UserAIConvo struct {
	ID                string    `json:"id"`
	Role              string    `json:"role"`
	UserID            string    `json:"user_id"`
	Message           string    `json:"message"`
	AdditionalContext string    `json:"additional_context"`
	Response          string    `json:"response"`
	CreatedAt         time.Time `json:"created_at"`
}

type UserAIConvos struct {
	Results []UserAIConvo   `json:"results"`
	Paging  *sdk.Pagination `json:"pagination"`
}

type GeneralAIContextQueryFilter struct {
	ID         null.String     `json:"id"`
	Pagination *sdk.Pagination `json:"pagination"`
}

type GeneralAIContext struct {
	ID      string `json:"id"`
	Context string `json:"content"`
}
