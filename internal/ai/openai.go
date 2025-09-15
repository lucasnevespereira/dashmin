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

// OpenAI Provider
type OpenAI struct {
	apiKey string
	client *http.Client
}

func NewOpenAI(apiKey string) *OpenAI {
	return &OpenAI{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (o *OpenAI) GenerateQuery(req QueryRequest) (*QueryResponse, error) {
	return o.generateQuery(req)
}

func (o *OpenAI) generateQuery(req QueryRequest) (*QueryResponse, error) {
	prompt := o.buildPrompt(req)

	payload := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":  500,
		"temperature": 0.1,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req_http, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req_http.Header.Set("Content-Type", "application/json")
	req_http.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req_http)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return &QueryResponse{Error: fmt.Sprintf("OpenAI API error: %s", string(body))}, nil
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
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

	if len(result.Choices) == 0 {
		return &QueryResponse{Error: "no response from OpenAI"}, nil
	}

	sql := strings.TrimSpace(result.Choices[0].Message.Content)

	// Clean up response - remove markdown formatting
	sql = strings.TrimPrefix(sql, "```sql")
	sql = strings.TrimPrefix(sql, "```")
	sql = strings.TrimSuffix(sql, "```")
	sql = strings.TrimSpace(sql)

	return &QueryResponse{SQL: sql}, nil
}

func (o *OpenAI) buildPrompt(req QueryRequest) string {
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