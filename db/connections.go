package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Result struct {
	Columns []string
	Rows    [][]interface{}
	Error   error
}

type Connection interface {
	Query(query string) (*Result, error)
	Close() error
}

type SQLConnection struct {
	db *sql.DB
}

func (c *SQLConnection) Query(query string) (*Result, error) {
	rows, err := c.db.Query(query)
	if err != nil {
		return &Result{Error: err}, nil
	}
	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		return &Result{Error: err}, nil
	}

	var result [][]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return &Result{Error: err}, nil
		}

		// Convert []byte to string for display
		for i, val := range values {
			if b, ok := val.([]byte); ok {
				values[i] = string(b)
			}
		}

		result = append(result, values)
	}

	return &Result{
		Columns: columns,
		Rows:    result,
	}, nil
}

func (c *SQLConnection) Close() error {
	return c.db.Close()
}

type MongoConnection struct {
	client *mongo.Client
	dbName string
}

func (c *MongoConnection) Query(query string) (*Result, error) {
	// Parse MongoDB queries in format: "collection.operation(filter)"
	parts := strings.SplitN(query, ".", 2)
	if len(parts) != 2 {
		return &Result{Error: fmt.Errorf("invalid MongoDB query format. Use: collection.operation(filter)")}, nil
	}

	collection := parts[0]
	operation := parts[1]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := c.client.Database(c.dbName).Collection(collection)

	if strings.HasPrefix(operation, "count(") {
		// Extract filter from count({filter})
		filterStr := strings.TrimSuffix(strings.TrimPrefix(operation, "count("), ")")
		if filterStr == "" {
			filterStr = "{}"
		}

		var filter bson.M
		if err := json.Unmarshal([]byte(filterStr), &filter); err != nil {
			filter = bson.M{}
		}

		count, err := coll.CountDocuments(ctx, filter)
		if err != nil {
			return &Result{Error: err}, nil
		}

		return &Result{
			Columns: []string{"count"},
			Rows:    [][]interface{}{{count}},
		}, nil
	}

	return &Result{
		Error: fmt.Errorf("unsupported MongoDB operation. Use: count({filter})"),
	}, nil
}

func (c *MongoConnection) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.client.Disconnect(ctx)
}

// ConnectPostgres connects to PostgreSQL using a connection string
func ConnectPostgres(connectionString string) (Connection, error) {
	// pgx driver supports both postgres:// URLs and key=value format
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &SQLConnection{db: db}, nil
}

// ConnectMySQL connects to MySQL using a connection string
func ConnectMySQL(connectionString string) (Connection, error) {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mysql: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping mysql: %w", err)
	}

	return &SQLConnection{db: db}, nil
}

// ConnectMongoDB connects to MongoDB using a connection string
func ConnectMongoDB(connectionString string) (Connection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	// Extract database name from connection string
	dbName := "test" // default
	if strings.Contains(connectionString, "/") {
		parts := strings.Split(connectionString, "/")
		if len(parts) > 3 {
			dbName = parts[len(parts)-1]
			if strings.Contains(dbName, "?") {
				dbName = strings.Split(dbName, "?")[0]
			}
		}
	}

	return &MongoConnection{client: client, dbName: dbName}, nil
}
