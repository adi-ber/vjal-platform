package llm

import (
	"context"
	"fmt"
	"os"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

// OpenAIClient wraps the official OpenAI Go SDK.
type OpenAIClient struct {
	client *openai.Client
}

// NewOpenAIClient constructs an OpenAIClient.
// If apiKey is empty, it reads from the OPENAI_API_KEY env var.
func NewOpenAIClient(apiKey string) Client {
	key := apiKey
	if key == "" {
		key = os.Getenv("OPENAI_API_KEY")
	}
	cli := openai.NewClient(option.WithAPIKey(key))
	return &OpenAIClient{client: &cli}
}

// Prompt sends your prompt to OpenAI and returns the assistant's reply.
func (o *OpenAIClient) Prompt(ctx context.Context, prompt string) (string, error) {
	params := openai.ChatCompletionNewParams{
		Model: shared.ChatModelGPT4o,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	}
	resp, err := o.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("OpenAI error: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("OpenAI returned no choices")
	}
	return resp.Choices[0].Message.Content, nil
}

// HealthCheck for OpenAIClient is a no‑op (always healthy).
func (o *OpenAIClient) HealthCheck(ctx context.Context) error {
	return nil
}
