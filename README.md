# dashmin

**Check your database metrics without leaving the terminal.**

You're coding, you want to quickly check how many users signed up today, or if there are errors in prod. Instead of opening a database UI, connecting, writing a query... just run `dashmin show`.

![dashmin demo](demo.gif)

## Why dashmin?

- **Stay in your terminal** - No browser, no GUI, no context switching
- **Instant setup** - 3 commands to your first dashboard
- **Your queries** - Track exactly what matters to you
- **Multi-database** - PostgreSQL, MySQL, MongoDB
- **AI-powered** - Generate queries from natural language (optional)

## Installation

```bash
go install github.com/lucasnevespereira/dashmin@latest
```

## Quick Start

### 1. Add your first app

```bash
dashmin app add myapp postgres "postgres://readonly:password@localhost:5432/myapp_prod?sslmode=disable"
```

### 2. Add metrics you want to track

```bash
dashmin query add myapp total_users "SELECT COUNT(*) FROM users"
dashmin query add myapp signups_today "SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE"
dashmin query add myapp revenue_today "SELECT SUM(amount) FROM orders WHERE created_at >= CURRENT_DATE"
```

### 3. View your dashboard

```bash
dashmin show           # Show all apps
dashmin show myapp     # Show specific app
```

## Commands

### Apps

```bash
dashmin app add <name> <type> <connection>   # Add new app
dashmin app list                              # List all apps
dashmin app test <name>                       # Test connection
dashmin app remove <name>                     # Remove app
```

### Queries

```bash
dashmin query add <app> <label> <query>       # Add a query
dashmin query list <app>                      # List queries for an app
dashmin query remove <app> <label>            # Remove a query
dashmin query generate <app> "<question>"     # Generate query with AI
```

### Dashboard

```bash
dashmin show                                  # Show all apps
dashmin show <app>                            # Show specific app
```

### Configuration

```bash
dashmin config show                           # Show current config
dashmin config path                           # Show config file path
dashmin config ai --provider openai --key sk-...  # Setup AI
dashmin config ai status                      # Check AI status
dashmin config ai reset                       # Remove AI config
```

## Query Examples

**User metrics:**

```bash
dashmin query add myapp total_users "SELECT COUNT(*) FROM users"
dashmin query add myapp active_users "SELECT COUNT(*) FROM users WHERE last_login > NOW() - INTERVAL '30 days'"
dashmin query add myapp signups_today "SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE"
```

**Business metrics:**

```bash
dashmin query add myapp orders_today "SELECT COUNT(*) FROM orders WHERE created_at >= CURRENT_DATE"
dashmin query add myapp revenue_today "SELECT SUM(amount) FROM orders WHERE created_at >= CURRENT_DATE"
dashmin query add myapp avg_order_value "SELECT ROUND(AVG(amount), 2) FROM orders"
```

**System health:**

```bash
dashmin query add myapp active_sessions "SELECT COUNT(*) FROM sessions WHERE expires_at > NOW()"
dashmin query add myapp errors_today "SELECT COUNT(*) FROM logs WHERE level = 'error' AND created_at >= CURRENT_DATE"
dashmin query add myapp database_size "SELECT pg_size_pretty(pg_database_size(current_database()))"
```

**MongoDB examples:**

```bash
dashmin query add webapp total_users "users.count({})"
dashmin query add webapp active_users "users.count({\"status\": \"active\"})"
dashmin query add webapp events_today "events.count({\"date\": {\"$gte\": \"2024-01-01\"}})"
```

## Configuration

Config is stored at `~/.config/dashmin/config.yaml`:

```yaml
apps:
  myapp:
    name: myapp
    type: postgres
    connection: "postgres://readonly:password@localhost:5432/myapp?sslmode=disable"
    queries:
      total_users: "SELECT COUNT(*) FROM users"
      signups_today: "SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE"
      revenue_today: "SELECT SUM(amount) FROM orders WHERE created_at >= CURRENT_DATE"
```

## Database Support

| Database   | Status | Connection String Example                                      |
| ---------- | ------ | -------------------------------------------------------------- |
| PostgreSQL | ✅     | `postgres://user:pass@localhost:5432/database?sslmode=disable` |
| MySQL      | ✅     | `user:pass@tcp(localhost:3306)/database`                       |
| MongoDB    | ✅     | `mongodb://user:pass@localhost:27017/database`                 |

### MongoDB Query Format

MongoDB queries use the format: `collection.operation({filter})`

Examples:

- `users.count({})` - Count all users
- `users.count({"status": "active"})` - Count active users
- `orders.count({"date": {"$gte": "2024-01-01"}})` - Count recent orders

## AI Query Generation (Optional)

Generate queries from natural language. No SQL knowledge required!

### Setup

```bash
# OpenAI
dashmin config ai --provider openai --key sk-your-openai-key

# Anthropic Claude
dashmin config ai --provider anthropic --key your-anthropic-key

# Check status
dashmin config ai status

# Remove AI config
dashmin config ai reset
```

Get your API key:

- OpenAI: https://platform.openai.com/api-keys
- Anthropic: https://console.anthropic.com/

### Usage

```bash
# Generate query (preview)
dashmin query generate myapp "users who signed up today"

# Execute immediately
dashmin query generate myapp "total revenue this month" --execute

# Save as reusable query
dashmin query generate myapp "active premium users" --save

# Both execute and save
dashmin query generate myapp "posts from last week" --save --execute
```

### Examples

```bash
# User metrics
dashmin query generate webapp "new users this month"
dashmin query generate webapp "users with more than 5 orders"
dashmin query generate webapp "users who haven't logged in for 30 days"

# Business metrics
dashmin query generate shop "revenue from premium customers this quarter"
dashmin query generate shop "average order value"
dashmin query generate shop "products with low inventory"

# System health
dashmin query generate myapp "database size"
dashmin query generate myapp "error logs from today"
```

### How it Works

1. **Schema Discovery** - Automatically reads your database structure
2. **AI Generation** - Converts natural language to SQL/MongoDB queries
3. **Review & Execute** - Shows generated query before running
4. **Save & Reuse** - Optionally save queries for monitoring

## Keyboard Shortcuts

- `r` - Refresh data
- `q` - Quit

## Troubleshooting

If you're having connection issues:

```bash
dashmin app test myapp
```

## Development

```bash
git clone https://github.com/lucasnevespereira/dashmin
cd dashmin
go mod tidy
go run main.go show
```

## Contributing

1. Fork the repo
2. Create a feature branch: `git checkout -b my-feature`
3. Make your changes
4. Test: `make build`
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details.

**Like dashmin?** Star ⭐ the repo and [support me](https://github.com/lucasnevespereira) for more tools!
