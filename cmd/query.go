package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/lucasnevespereira/dashmin/internal/ai"
	"github.com/lucasnevespereira/dashmin/internal/config"
	"github.com/lucasnevespereira/dashmin/internal/db"
	"github.com/spf13/cobra"
)

var (
	saveFlag    bool
	executeFlag bool
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Manage queries for an app",
	Long: `Manage custom queries to monitor specific metrics for an app.

Examples:
  dashmin query add myapp users "SELECT COUNT(*) FROM users"
  dashmin query list myapp
  dashmin query remove myapp users
  dashmin query generate myapp "users who signed up today"`,
}

var queryAddCmd = &cobra.Command{
	Use:   "add <app> <label> <query>",
	Short: "Add a custom query to an app",
	Long: `Add a custom query to monitor specific metrics for an app.

Examples:
  dashmin query add myapp users "SELECT COUNT(*) FROM users"
  dashmin query add myapp posts "SELECT COUNT(*) FROM posts WHERE created_at > NOW() - INTERVAL '30 days'"
  dashmin query add webapp revenue "SELECT SUM(amount) FROM payments WHERE DATE(created_at) = CURDATE()"
  dashmin query add analytics active_users "users.count({\"status\": \"active\"})"`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		label := args[1]
		query := args[2]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		app, exists := cfg.Apps[appName]
		if !exists {
			fmt.Printf("Error: App '%s' not found.\n", appName)
			if len(cfg.Apps) > 0 {
				fmt.Printf("Available apps: ")
				for name := range cfg.Apps {
					fmt.Printf("%s ", name)
				}
				fmt.Printf("\n")
			}
			return
		}

		if app.Queries == nil {
			app.Queries = make(map[string]string)
		}
		app.Queries[label] = query
		cfg.Apps[appName] = app

		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("Added query '%s' to app '%s'\n", label, appName)
		fmt.Printf("Query: %s\n", query)
		fmt.Printf("\nView results: dashmin show %s\n", appName)
	},
}

var queryRemoveCmd = &cobra.Command{
	Use:   "remove <app> <label>",
	Short: "Remove a query from an app",
	Long: `Remove a custom query from an app.

Examples:
  dashmin query remove myapp users
  dashmin query remove webapp revenue`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		label := args[1]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		app, exists := cfg.Apps[appName]
		if !exists {
			fmt.Printf("Error: App '%s' not found.\n", appName)
			if len(cfg.Apps) > 0 {
				fmt.Printf("Available apps: ")
				for name := range cfg.Apps {
					fmt.Printf("%s ", name)
				}
				fmt.Printf("\n")
			}
			return
		}

		if app.Queries == nil {
			fmt.Printf("App '%s' has no queries.\n", appName)
			return
		}

		querySQL, queryExists := app.Queries[label]
		if !queryExists {
			fmt.Printf("Error: Query '%s' not found in app '%s'.\n", label, appName)
			if len(app.Queries) > 0 {
				fmt.Printf("Available queries: ")
				for queryLabel := range app.Queries {
					fmt.Printf("%s ", queryLabel)
				}
				fmt.Printf("\n")
			}
			return
		}

		delete(app.Queries, label)
		cfg.Apps[appName] = app

		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("Removed query '%s' from app '%s'\n", label, appName)
		fmt.Printf("Query was: %s\n", querySQL)
	},
}

var queryListCmd = &cobra.Command{
	Use:   "list <app>",
	Short: "List all queries for an app",
	Long: `Show all configured queries for a specific app.

Examples:
  dashmin query list myapp`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		app, exists := cfg.Apps[appName]
		if !exists {
			fmt.Printf("Error: App '%s' not found.\n", appName)
			if len(cfg.Apps) > 0 {
				fmt.Printf("Available apps: ")
				for name := range cfg.Apps {
					fmt.Printf("%s ", name)
				}
				fmt.Printf("\n")
			}
			return
		}

		if len(app.Queries) == 0 {
			fmt.Printf("No queries configured for '%s'.\n", appName)
			fmt.Printf("\nAdd a query:\n")
			fmt.Printf("  dashmin query add %s <label> \"<query>\"\n", appName)
			return
		}

		fmt.Printf("Queries for '%s' (%d):\n\n", appName, len(app.Queries))

		for label, query := range app.Queries {
			fmt.Printf("  %s\n", label)
			fmt.Printf("    %s\n\n", query)
		}
	},
}

var queryGenerateCmd = &cobra.Command{
	Use:   "generate <app> \"<natural language query>\"",
	Short: "Generate a query using AI from natural language",
	Long: `Use AI to convert natural language descriptions into SQL/MongoDB queries.
The AI will analyze your app's database schema and generate appropriate queries.

Examples:
  dashmin query generate myapp "users who signed up today"
  dashmin query generate myapp "total revenue this month" --save
  dashmin query generate myapp "active premium users" --execute
  dashmin query generate myapp "posts published last week" --save --execute`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		prompt := args[1]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		if cfg.AI == nil || cfg.AI.APIKey == "" {
			fmt.Printf("AI not configured. Set up AI integration first:\n")
			fmt.Printf("\n  dashmin config ai --provider openai --key sk-your-key\n")
			fmt.Printf("  dashmin config ai status\n")
			return
		}

		app, exists := cfg.Apps[appName]
		if !exists {
			fmt.Printf("Error: App '%s' not found.\n", appName)
			if len(cfg.Apps) > 0 {
				fmt.Printf("Available apps: ")
				for name := range cfg.Apps {
					fmt.Printf("%s ", name)
				}
				fmt.Printf("\n")
			}
			return
		}

		fmt.Printf("Analyzing database schema...\n")
		conn, err := connectToApp(app)
		if err != nil {
			fmt.Printf("Failed to connect to database: %v\n", err)
			return
		}
		defer func() { _ = conn.Close() }()

		schema, err := db.GetDatabaseSchema(conn, app.Type)
		if err != nil {
			fmt.Printf("Warning: Could not retrieve schema: %v\n", err)
			schema = ""
		}

		fmt.Printf("Generating query...\n")
		engine, err := ai.NewEngine(cfg.AI.Provider, cfg.AI.APIKey)
		if err != nil {
			fmt.Printf("Failed to initialize AI engine: %v\n", err)
			return
		}

		response, err := engine.GenerateQuery(ai.QueryRequest{
			Prompt:       prompt,
			Schema:       schema,
			DatabaseType: app.Type,
		})
		if err != nil {
			fmt.Printf("Failed to generate query: %v\n", err)
			return
		}

		if response.Error != "" {
			fmt.Printf("AI Error: %s\n", response.Error)
			return
		}

		fmt.Printf("\nGenerated Query:\n")
		fmt.Printf("  %s\n", response.SQL)

		if strings.Contains(strings.ToUpper(response.SQL), "SELECT *") {
			fmt.Printf("\nWarning: This query returns all columns. For dashboard metrics, consider using COUNT(*), SUM(), or AVG().\n")
		}

		if executeFlag {
			executeGeneratedQuery(conn, response.SQL)
		}
		if saveFlag {
			saveGeneratedQuery(cfg, appName, response.SQL, prompt)
		}
	},
}

func connectToApp(app config.App) (db.Connection, error) {
	switch app.Type {
	case "postgres":
		return db.ConnectPostgres(app.Connection)
	case "mysql":
		return db.ConnectMySQL(app.Connection)
	case "mongodb":
		return db.ConnectMongoDB(app.Connection)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", app.Type)
	}
}

func executeGeneratedQuery(conn db.Connection, query string) {
	fmt.Printf("\nExecuting query...\n")

	result, err := conn.Query(query)
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
		return
	}

	if result.Error != nil {
		fmt.Printf("Query error: %v\n", result.Error)
		return
	}

	fmt.Printf("\nResults:\n")

	if len(result.Rows) == 0 {
		fmt.Printf("  No results\n")
		return
	}

	for i, col := range result.Columns {
		if i > 0 {
			fmt.Printf(" | ")
		}
		fmt.Printf("%s", col)
	}
	fmt.Printf("\n")

	for i, col := range result.Columns {
		if i > 0 {
			fmt.Printf("-+-")
		}
		fmt.Printf("%s", strings.Repeat("-", len(col)))
	}
	fmt.Printf("\n")

	maxRows := 10
	for i, row := range result.Rows {
		if i >= maxRows {
			fmt.Printf("... (%d more rows)\n", len(result.Rows)-maxRows)
			break
		}

		for j, val := range row {
			if j > 0 {
				fmt.Printf(" | ")
			}
			fmt.Printf("%v", val)
		}
		fmt.Printf("\n")
	}
}

func saveGeneratedQuery(cfg *config.Config, appName, query, prompt string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\nEnter label for this query: ")
	label, _ := reader.ReadString('\n')
	label = strings.TrimSpace(label)

	if label == "" {
		fmt.Printf("Empty label. Query not saved.\n")
		return
	}

	app := cfg.Apps[appName]
	if app.Queries == nil {
		app.Queries = make(map[string]string)
	}
	app.Queries[label] = query
	cfg.Apps[appName] = app

	if err := cfg.Save(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		return
	}

	fmt.Printf("Query saved as '%s'\n", label)
	fmt.Printf("Generated from prompt: \"%s\"\n", prompt)
	fmt.Printf("\nView in dashboard: dashmin show %s\n", appName)
}

func init() {
	queryGenerateCmd.Flags().BoolVar(&saveFlag, "save", false, "Save the generated query")
	queryGenerateCmd.Flags().BoolVar(&executeFlag, "execute", false, "Execute the generated query immediately")

	queryCmd.AddCommand(queryAddCmd)
	queryCmd.AddCommand(queryRemoveCmd)
	queryCmd.AddCommand(queryListCmd)
	queryCmd.AddCommand(queryGenerateCmd)
}
