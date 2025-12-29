package openai_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"wingedapp/pgtester/internal/lib/openai"
	"wingedapp/pgtester/internal/util/strutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAI_Ask tests basic prompting functionality, and response mapping.
func TestOpenAI_Ask(t *testing.T) {
	t.Skip()
	lib, err := openai.NewLib(&openai.Config{
		APIKey: os.Getenv("WINGED_OPENAI_API_KEY"),
	})
	require.NoError(t, err, "create openai lib")

	opts := &openai.PromptOpts{
		Message: "Are you a robot?",
	}

	resp, err := lib.Prompt(context.Background(), opts)
	require.NoError(t, err, "expect no error from openai ask")

	assert.NotEmpty(t, resp.ID, "expect response Code")
	assert.NotEmpty(t, resp.SentMessage, "expect response content")
}

func TestOpenAI_Ask_WithContext(t *testing.T) {
	t.Skip()
	const (
		additionalContext = `this is some additional context for AI - 
			the user is an elf, but dont say anything about it UNLESS the user asks what he is..
			PREFACE at the end what the user asked...
        `
	)

	lib, err := openai.NewLib(&openai.Config{
		APIKey: os.Getenv("WINGED_OPENAI_API_KEY"),
	})
	require.NoError(t, err, "create openai lib")

	opts := &openai.PromptOpts{
		Message:           "Are you a robot? If not , wha am I?",
		AdditionalContext: additionalContext,
	}

	resp, err := lib.Prompt(context.Background(), opts)
	require.NoError(t, err, "expect no error from openai ask")

	assert.NotEmpty(t, resp.ID, "expect response Code")
	assert.NotEmpty(t, resp.Response, "expect response content")
	assert.Equal(t, additionalContext, resp.AdditionalContext, "expect additional context to be echoed back")

	fmt.Println("=== resp w/context:", strutil.GetAsJson(resp))
}
