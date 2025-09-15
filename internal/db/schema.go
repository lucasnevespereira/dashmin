package db

import (
	"fmt"
	"strings"
)

func GetDatabaseSchema(conn Connection, dbType string) (string, error) {
	switch dbType {
	case "postgres":
		return getPostgresSchema(conn)
	case "mysql":
		return getMySQLSchema(conn)
	case "mongodb":
		return getMongoDBSchema(conn)
	default:
		return "", fmt.Errorf("unsupported database type: %s", dbType)
	}
}

func getPostgresSchema(conn Connection) (string, error) {
	query := `
		SELECT
			table_name,
			column_name,
			data_type,
			is_nullable
		FROM information_schema.columns
		WHERE table_schema = 'public'
		ORDER BY table_name, ordinal_position
	`

	result, err := conn.Query(query)
	if err != nil {
		return "", fmt.Errorf("failed to get PostgreSQL schema: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("schema query error: %w", result.Error)
	}

	return formatSchemaResult(result, "PostgreSQL"), nil
}

func getMySQLSchema(conn Connection) (string, error) {
	query := `
		SELECT
			table_name,
			column_name,
			data_type,
			is_nullable
		FROM information_schema.columns
		WHERE table_schema = DATABASE()
		ORDER BY table_name, ordinal_position
	`

	result, err := conn.Query(query)
	if err != nil {
		return "", fmt.Errorf("failed to get MySQL schema: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("schema query error: %w", result.Error)
	}

	return formatSchemaResult(result, "MySQL"), nil
}

func getMongoDBSchema(conn Connection) (string, error) {
	return `MongoDB collections (common field patterns):
- users: {_id, email, name, status, created_at/createdAt, updated_at/updatedAt}
- orders: {_id, user_id, amount, status, created_at/createdAt, items}
- posts: {_id, title, content, author_id, created_at/createdAt}

Note: Field names vary by project. Common patterns:
- created_at OR createdAt
- updated_at OR updatedAt
- date_created OR dateCreated

IMPORTANT: Only count() operation is supported. Use JSON format.
Query format: collection.count({filter})
Examples:
- users.count({"status": "active"})
- users.count({"createdAt": {"$gte": "2024-01-01"}})`, nil
}

func formatSchemaResult(result *Result, dbType string) string {
	var schema strings.Builder
	schema.WriteString(fmt.Sprintf("%s Database Schema:\n\n", dbType))

	currentTable := ""
	for _, row := range result.Rows {
		tableName := fmt.Sprintf("%v", row[0])
		columnName := fmt.Sprintf("%v", row[1])
		dataType := fmt.Sprintf("%v", row[2])
		nullable := fmt.Sprintf("%v", row[3])

		if tableName != currentTable {
			if currentTable != "" {
				schema.WriteString("\n")
			}
			schema.WriteString(fmt.Sprintf("Table: %s\n", tableName))
			currentTable = tableName
		}

		nullableStr := ""
		if nullable == "YES" {
			nullableStr = " (nullable)"
		}

		schema.WriteString(fmt.Sprintf("  - %s: %s%s\n", columnName, dataType, nullableStr))
	}

	return schema.String()
}