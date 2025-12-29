package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Responses API endpoint (OpenAI's new unified API - March 2025)
const responsesAPIURL = "https://api.openai.com/v1/responses"

// WorkflowRequest represents a request to the OpenAI Responses API with a workflow.
type WorkflowRequest struct {
	WorkflowID string `json:"workflow_id" validate:"required"`
	InputJSON  string `json:"input_json" validate:"required"` // JSON string to pass as input
	Model      string `json:"model"`                          // Optional, defaults to gpt-4.1
}

// WorkflowResponse represents the response from an OpenAI workflow.
type WorkflowResponse struct {
	ID           string          `json:"id"`
	OutputParsed json.RawMessage `json:"output_parsed"` // Raw JSON for caller to unmarshal
	CreatedAt    time.Time       `json:"created_at"`
}

// responsesAPIRequest is the internal request format for OpenAI Responses API.
type responsesAPIRequest struct {
	Model      string                   `json:"model"`
	WorkflowID string                   `json:"workflow_id"`
	Input      []map[string]interface{} `json:"input"`
}

// responsesAPIResponse is the internal response format from OpenAI Responses API.
type responsesAPIResponse struct {
	ID           string          `json:"id"`
	OutputParsed json.RawMessage `json:"output_parsed"`
	CreatedAt    int64           `json:"created_at"`
}

// RunWorkflow executes an OpenAI workflow via the Responses API.
// This is the new unified API that combines Chat Completions + Assistants.
//
// Usage:
//
//	inputJSON, _ := json.Marshal(myRequest)
//	resp, err := lib.RunWorkflow(ctx, &WorkflowRequest{
//	    WorkflowID: "wf_xxx",
//	    InputJSON:  string(inputJSON),
//	})
//	var result MyResultType
//	json.Unmarshal(resp.OutputParsed, &result)
func (l *Lib) RunWorkflow(ctx context.Context, req *WorkflowRequest) (*WorkflowResponse, error) {
	if req.WorkflowID == "" {
		return nil, fmt.Errorf("workflow_id is required")
	}
	if req.InputJSON == "" {
		return nil, fmt.Errorf("input_json is required")
	}

	model := req.Model
	if model == "" {
		model = "gpt-4.1"
	}

	// Build the Responses API request
	apiReq := responsesAPIRequest{
		Model:      model,
		WorkflowID: req.WorkflowID,
		Input: []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]string{
					{"type": "input_text", "text": req.InputJSON},
				},
			},
		},
	}

	reqBody, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, responsesAPIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+l.apiKey())

	// Execute request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp responsesAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &WorkflowResponse{
		ID:           apiResp.ID,
		OutputParsed: apiResp.OutputParsed,
		CreatedAt:    time.Unix(apiResp.CreatedAt, 0),
	}, nil
}

// apiKey returns the API key from the underlying client.
// This is needed because the official SDK doesn't expose Responses API yet.
func (l *Lib) apiKey() string {
	// We need to store the API key separately since the SDK client doesn't expose it
	// This will be set during construction
	return l.key
}
