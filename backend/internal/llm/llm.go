package llm

import (
	"context"
	"fmt"
)

// Client is the interface for LLM providers.
type Client interface {
	// Generate produces 3 variant drafts from a prompt.
	Generate(ctx context.Context, prompt string) ([]string, error)
	// Tweak revises existing content according to an instruction.
	Tweak(ctx context.Context, content, instruction string) (string, error)
}

// NewClient creates a Client for the given provider.
func NewClient(provider, model, apiKey string) (Client, error) {
	switch provider {
	case "gemini":
		return &geminiClient{model: model, apiKey: apiKey}, nil
	case "claude":
		return &claudeClient{model: model, apiKey: apiKey}, nil
	case "openai":
		return &openaiClient{model: model, apiKey: apiKey}, nil
	default:
		return nil, fmt.Errorf("unknown LLM provider: %q", provider)
	}
}
