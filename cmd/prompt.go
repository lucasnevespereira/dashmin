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

var promptCmd = &cobra.Command{
	Use:   "prompt <app> \"<natural language query>\"",
	Short: "Generate database queries using AI from natural language",
	Long: `Use AI to convert natural language descriptions into SQL/MongoDB queries.
The AI will analyze your app's database schema and generate appropriate queries.

Examples:
  dashmin prompt myapp "users who signed up today"
  dashmin prompt myapp "total revenue this month" --save monthly_revenue
  dashmin prompt myapp "active premium users" --execute
  dashmin prompt blogapp "posts published last week" --save --execute recent_posts`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		prompt := args[1]

		// Load config
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		// Check AI configuration
		if cfg.AI == nil || cfg.AI.APIKey == "" {
			fmt.Printf("❌ AI not configured. Set up AI integration first:\n")
			fmt.Printf("\nSetup: dashmin ai --provider openai --key sk-your-key\n")
			fmt.Printf("Status: dashmin ai status\n")
			return
		}

		// Check if app exists
		app, exists := cfg.Apps[appName]
		if !exists {
			fmt.Printf("❌ App '%s' not found.\n", appName)
			fmt.Printf("Available apps: ")
			for name := range cfg.Apps {
				fmt.Printf("%s ", name)
			}
			fmt.Printf("\n")
			return
		}

		// Connect to database and get schema
		fmt.Printf("🔍 Analyzing database schema...\n")
		conn, err := connectToApp(app)
		if err != nil {
			fmt.Printf("❌ Failed to connect to database: %v\n", err)
			return
		}
		defer func() { _ = conn.Close() }()

		schema, err := db.GetDatabaseSchema(conn, app.Type)
		if err != nil {
			fmt.Printf("⚠️  Warning: Could not retrieve schema: %v\n", err)
			schema = "" // Continue without schema
		}

		// Generate query using AI
		fmt.Printf("🤖 Generating query...\n")
		engine, err := ai.NewEngine(cfg.AI.Provider, cfg.AI.APIKey)
		if err != nil {
			fmt.Printf("❌ Failed to initialize AI engine: %v\n", err)
			return
		}

		response, err := engine.GenerateQuery(ai.QueryRequest{
			Prompt:       prompt,
			Schema:       schema,
			DatabaseType: app.Type,
		})
		if err != nil {
			fmt.Printf("❌ Failed to generate query: %v\n", err)
			return
		}

		if response.Error != "" {
			fmt.Printf("❌ AI Error: %s\n", response.Error)
			return
		}

		// Display generated query
		fmt.Printf("\n📋 Generated Query:\n")
		fmt.Printf("─────────────────────\n")
		fmt.Printf("%s\n", response.SQL)
		fmt.Printf("─────────────────────\n")

		// Warn if query looks like it might return too much data
		if strings.Contains(strings.ToUpper(response.SQL), "SELECT *") {
			fmt.Printf("\n⚠️  This query returns all columns. For better dashboard metrics, consider using COUNT(*), SUM(), or AVG().\n")
		}

		// Handle save and execute flags
		if executeFlag {
			executeQuery(conn, response.SQL)
		}
		if saveFlag {
			saveQuery(cfg, appName, response.SQL, prompt)
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


func executeQuery(conn db.Connection, query string) {
	fmt.Printf("\n🔄 Executing query...\n")

	result, err := conn.Query(query)
	if err != nil {
		fmt.Printf("❌ Execution error: %v\n", err)
		return
	}

	if result.Error != nil {
		fmt.Printf("❌ Query error: %v\n", result.Error)
		return
	}

	// Display results
	fmt.Printf("\n📊 Results:\n")
	fmt.Printf("─────────────────────\n")

	if len(result.Rows) == 0 {
		fmt.Printf("No results\n")
		return
	}

	// Show column headers
	for i, col := range result.Columns {
		if i > 0 {
			fmt.Printf(" | ")
		}
		fmt.Printf("%s", col)
	}
	fmt.Printf("\n")

	// Show separator
	for i, col := range result.Columns {
		if i > 0 {
			fmt.Printf("-+-")
		}
		fmt.Printf("%s", strings.Repeat("-", len(col)))
	}
	fmt.Printf("\n")

	// Show data rows (limit to 10 for readability)
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
	fmt.Printf("─────────────────────\n")
}

func saveQuery(cfg *config.Config, appName, query, prompt string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\n💾 Enter label for this query: ")
	label, _ := reader.ReadString('\n')
	label = strings.TrimSpace(label)

	if label == "" {
		fmt.Printf("❌ Empty label. Query not saved.\n")
		return
	}

	// Add query to app
	app := cfg.Apps[appName]
	if app.Queries == nil {
		app.Queries = make(map[string]string)
	}
	app.Queries[label] = query
	cfg.Apps[appName] = app

	// Save config
	if err := cfg.Save(); err != nil {
		fmt.Printf("❌ Error saving config: %v\n", err)
		return
	}

	fmt.Printf("✅ Query saved as '%s'\n", label)
	fmt.Printf("💡 Generated from prompt: \"%s\"\n", prompt)
	fmt.Printf("\nView in dashboard: dashmin see %s\n", appName)
}

func init() {
	promptCmd.Flags().BoolVar(&saveFlag, "save", false, "Save the generated query as a reusable query")
	promptCmd.Flags().BoolVar(&executeFlag, "execute", false, "Execute the generated query immediately")
}
