package provider

import "context"

type Provider interface {
	Generate(ctx context.Context, prompt string) (string, error)

	Name() string
}

func NewProvider(name, apiKey string, maxRetries int) (Provider, error) {
	switch name {
	case "gemini":
		return NewGemini(apiKey, maxRetries)
	// case "claude":
	//     return NewClaude(apiKey, maxRetries)
	default:
		return NewGemini(apiKey, maxRetries)
	}
}
