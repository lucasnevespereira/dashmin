package ai

import (
	"fmt"
)

type Provider interface {
	GenerateQuery(req QueryRequest) (*QueryResponse, error)
}

type Engine struct {
	provider Provider
}

func NewEngine(providerName, apiKey string) (*Engine, error) {
	switch providerName {
	case "openai":
		return &Engine{
			provider: NewOpenAI(apiKey),
		}, nil
	case "anthropic":
		return &Engine{
			provider: NewAnthropic(apiKey),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", providerName)
	}
}

func (e *Engine) GenerateQuery(req QueryRequest) (*QueryResponse, error) {
	return e.provider.GenerateQuery(req)
}

// Available providers
func GetAvailableProviders() []string {
	return []string{"openai", "anthropic"}
}