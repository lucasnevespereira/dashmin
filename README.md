# dashmin

**Check your database metrics without leaving the terminal.**

You're coding, you want to quickly check how many users signed up today, or if there are errors in prod. Instead of opening a database UI, connecting, writing a query... just run `dashmin show`.

![dashmin demo](assets/demo.png)

## Why dashmin?

- **Stay in your terminal** - No browser, no GUI, no context switching
- **Instant setup** - 3 commands to your first dashboard
- **Your queries** - Track exactly what matters to you
- **Multi-database** - PostgreSQL, MySQL, MongoDB, SQLite
- **AI-powered** - Generate queries from natural language (optional)

## Installation

```bash
go install github.com/lucasnevespereira/dashmin@latest
```

## Quick Start

```bash
# 1. Add your first app
dashmin app add myapp postgres "postgres://readonly:password@localhost:5432/myapp_prod?sslmode=disable"

# 2. Add a metric to track
dashmin query add myapp total_users "SELECT COUNT(*) FROM users"

# 3. View your dashboard
dashmin show
```

## Commands

| Command                                      | Description                        |
| -------------------------------------------- | ---------------------------------- |
| `dashmin app add <name> <type> <connection>` | Add new app                        |
| `dashmin app list`                           | List all apps                      |
| `dashmin app test <name>`                    | Test connection                    |
| `dashmin app remove <name>`                  | Remove app (with confirmation)     |
| `dashmin query add <app> <label> <query>`    | Add a query                        |
| `dashmin query list <app>`                   | List queries for an app            |
| `dashmin query remove <app> <label>`         | Remove a query (with confirmation) |
| `dashmin query generate <app> "<question>"`  | Generate query with AI             |
| `dashmin show`                               | Show all apps                      |
| `dashmin show <app>`                         | Show specific app                  |

## Query Examples

### User Metrics

```bash
# Total users
dashmin query add myapp total_users "SELECT COUNT(*) FROM users"

# Active users (logged in last 30 days)
dashmin query add myapp active_users "SELECT COUNT(*) FROM users WHERE last_login > NOW() - INTERVAL '30 days'"

# Signups today
dashmin query add myapp signups_today "SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE"
```

### Business Metrics

```bash
# Orders today
dashmin query add myapp orders_today "SELECT COUNT(*) FROM orders WHERE created_at >= CURRENT_DATE"

# Revenue today
dashmin query add myapp revenue_today "SELECT SUM(amount) FROM orders WHERE created_at >= CURRENT_DATE"

# Average order value
dashmin query add myapp avg_order_value "SELECT ROUND(AVG(amount), 2) FROM orders"
```

### System Health

```bash
# Active sessions
dashmin query add myapp active_sessions "SELECT COUNT(*) FROM sessions WHERE expires_at > NOW()"

# Errors today
dashmin query add myapp errors_today "SELECT COUNT(*) FROM logs WHERE level = 'error' AND created_at >= CURRENT_DATE"

# Database size
dashmin query add myapp db_size "SELECT pg_size_pretty(pg_database_size(current_database()))"
```

### MongoDB Examples

```bash
dashmin query add analytics total_users "users.count({})"
dashmin query add analytics active_users "users.count({\"status\": \"active\"})"
dashmin query add analytics events_today "events.count({\"date\": {\"$gte\": \"2024-01-01\"}})"
```

## Database Support

| Database   | Status | Connection String Example                                      |
| ---------- | ------ | -------------------------------------------------------------- |
| PostgreSQL | ✅     | `postgres://user:pass@localhost:5432/database?sslmode=disable` |
| MySQL      | ✅     | `user:pass@tcp(localhost:3306)/database`                       |
| MongoDB    | ✅     | `mongodb://user:pass@localhost:27017/database`                 |
| SQLite     | ✅     | `sqlite:///path/to/database.db`                                |

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
```

### Usage

```bash
# Generate query (preview)
dashmin query generate myapp "how many users signed up today"
# → Generates: SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE

# Execute immediately
dashmin query generate myapp "total revenue this month" --execute

# Save as reusable query
dashmin query generate myapp "active premium users" --save

# Both execute and save
dashmin query generate myapp "posts from last week" --save --execute
```

### How It Works

1. **Schema Discovery** - Automatically reads your database structure
2. **AI Generation** - Converts natural language to SQL/MongoDB queries
3. **Review & Execute** - Shows generated query before running
4. **Save & Reuse** - Optionally save queries for monitoring

## Keyboard Shortcuts

When viewing the dashboard (`dashmin show`):

- `r` - Refresh data
- `q` - Quit
- `?` - Show error details (when errors present)

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
