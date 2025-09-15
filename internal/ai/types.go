package ai

// QueryRequest represents a request to generate a database query
type QueryRequest struct {
	Prompt       string
	Schema       string
	DatabaseType string
}

// QueryResponse represents the response from AI query generation
type QueryResponse struct {
	SQL   string
	Error string
}