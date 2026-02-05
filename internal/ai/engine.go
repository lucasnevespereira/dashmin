package ai

import (
	"fmt"
	"strings"
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

// BuildPrompt creates system and user messages for query generation
func BuildPrompt(req QueryRequest) PromptMessages {
	var system strings.Builder

	system.WriteString(fmt.Sprintf("You are a %s query generator for a terminal dashboard tool called dashmin.\n", strings.ToUpper(req.DatabaseType)))

	switch req.DatabaseType {
	case "postgres", "mysql":
		system.WriteString("Return only the SQL query without any explanation or formatting.\n")
		system.WriteString("Use standard SQL that works with " + req.DatabaseType + ".\n")
		system.WriteString("For dashmin monitoring, prefer COUNT(*), SUM(), AVG() and other aggregate functions over SELECT *.\n")
		system.WriteString("Generate queries that return single metrics suitable for dashboard display.\n")
		system.WriteString("Do not wrap the query in markdown code blocks.")
	case "mongodb":
		system.WriteString("Return only the MongoDB query in the format: collection.count({filter}).\n")
		system.WriteString("Only count() operation is supported. Use JSON format, not JavaScript.\n")
		system.WriteString("For dates, use ISO date strings with $gte/$lt operators.\n")
		system.WriteString("Examples: users.count({\"status\": \"active\"}) or users.count({\"created_at\": {\"$gte\": \"2024-01-01\"}})\n")
		system.WriteString("Do not wrap the query in markdown code blocks.")
	}

	var user strings.Builder
	user.WriteString(req.Prompt)

	if req.Schema != "" {
		user.WriteString("\n\nDatabase schema:\n")
		user.WriteString(req.Schema)
	}

	return PromptMessages{
		System: system.String(),
		User:   user.String(),
	}
}
