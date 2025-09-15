package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Anthropic Provider
type Anthropic struct {
	apiKey string
	client *http.Client
}

func NewAnthropic(apiKey string) *Anthropic {
	return &Anthropic{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (a *Anthropic) GenerateQuery(req QueryRequest) (*QueryResponse, error) {
	prompt := a.buildPrompt(req)

	payload := map[string]interface{}{
		"model":      "claude-3-haiku-20240307",
		"max_tokens": 500,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req_http, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req_http.Header.Set("Content-Type", "application/json")
	req_http.Header.Set("x-api-key", a.apiKey)
	req_http.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req_http)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return &QueryResponse{Error: fmt.Sprintf("Anthropic API error: %s", string(body))}, nil
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.Error.Message != "" {
		return &QueryResponse{Error: result.Error.Message}, nil
	}

	if len(result.Content) == 0 {
		return &QueryResponse{Error: "no response from Anthropic"}, nil
	}

	sql := strings.TrimSpace(result.Content[0].Text)

	// Clean up response
	sql = strings.TrimPrefix(sql, "```sql")
	sql = strings.TrimPrefix(sql, "```")
	sql = strings.TrimSuffix(sql, "```")
	sql = strings.TrimSpace(sql)

	return &QueryResponse{SQL: sql}, nil
}

func (a *Anthropic) buildPrompt(req QueryRequest) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("Generate a %s query for the following request.\n\n", strings.ToUpper(req.DatabaseType)))
	prompt.WriteString(fmt.Sprintf("User request: %s\n\n", req.Prompt))

	if req.Schema != "" {
		prompt.WriteString("Database schema:\n")
		prompt.WriteString(req.Schema)
		prompt.WriteString("\n\n")
	}

	switch req.DatabaseType {
	case "postgres", "mysql":
		prompt.WriteString("Return only the SQL query without any explanation or formatting. ")
		prompt.WriteString("Use standard SQL that works with " + req.DatabaseType + ". ")
		prompt.WriteString("For dashmin monitoring, prefer COUNT(*), SUM(), AVG() and other aggregate functions over SELECT *. ")
		prompt.WriteString("Generate queries that return single metrics suitable for dashboard display.")
	case "mongodb":
		prompt.WriteString("Return only the MongoDB query in the format: collection.count({filter}). ")
		prompt.WriteString("Only count() operation is supported. Use JSON format, not JavaScript. ")
		prompt.WriteString("For dates, use ISO date strings with $gte/$lt operators. ")
		prompt.WriteString("Examples: users.count({\"status\": \"active\"}) or users.count({\"created_at\": {\"$gte\": \"2024-01-01\"}})")
	}

	return prompt.String()
}