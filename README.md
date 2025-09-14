# dashmin

**Minimal dashboard for monitoring your apps from the terminal**

Quick insights without the bloat. Monitor your databases with simple queries - no complex setup, no heavy interfaces.

## Features

- **Multi-database support** - PostgreSQL, MySQL, MongoDB
- **Custom queries** - Define the metrics that matter to you
- **Real-time TUI** - Beautiful terminal interface with live updates
- **Minimal config** - Simple YAML configuration
- **Single binary** - No dependencies, easy deployment

## Installation

### From source
```bash
go install github.com/lucasnevespereira/dashmin@latest
```

### Build locally
```bash
git clone https://github.com/lucasnevespereira/dashmin.git
cd dashmin
go build -o dashmin main.go
```

## Quick Start

### 1. Add your first app
```bash
dashmin add myapp postgres "postgres://readonly:password@localhost:5432/myapp_prod?sslmode=disable"
```

### 2. Add metrics you want to track
```bash
dashmin query myapp total_users "SELECT COUNT(*) FROM users"
dashmin query myapp signups_today "SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE"
dashmin query myapp revenue_today "SELECT SUM(amount) FROM orders WHERE created_at >= CURRENT_DATE"
```

### 3. View your dashboard
```bash
dashmin all          # See all apps
dashmin see myapp    # See specific app
```

## Usage

### Commands

```bash
# Setup
dashmin add <name> <type> <connection-string>
dashmin query <app> <label> <query>

# View dashboards
dashmin all          # All apps overview
dashmin see <app>    # Single app focus

# Management
dashmin list         # Show configuration
dashmin test <app>   # Test connection
dashmin remove <app> # Remove app
```

### Common Queries

**User metrics:**
```bash
dashmin query myapp total_users "SELECT COUNT(*) FROM users"
dashmin query myapp active_users "SELECT COUNT(*) FROM users WHERE last_login > NOW() - INTERVAL '30 days'"
dashmin query myapp signups_today "SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE"
```

**Business metrics:**
```bash
dashmin query myapp orders_today "SELECT COUNT(*) FROM orders WHERE created_at >= CURRENT_DATE"
dashmin query myapp revenue_today "SELECT SUM(amount) FROM orders WHERE created_at >= CURRENT_DATE"
dashmin query myapp avg_order_value "SELECT ROUND(AVG(amount), 2) FROM orders"
```

**System health:**
```bash
dashmin query myapp active_sessions "SELECT COUNT(*) FROM sessions WHERE expires_at > NOW()"
dashmin query myapp errors_today "SELECT COUNT(*) FROM logs WHERE level = 'error' AND created_at >= CURRENT_DATE"
dashmin query myapp database_size "SELECT pg_size_pretty(pg_database_size(current_database()))"
```

**MongoDB examples:**
```bash
dashmin query webapp total_users "users.count({})"
dashmin query webapp active_users "users.count({\"status\": \"active\"})"
dashmin query webapp events_today "events.count({\"date\": {\"$gte\": \"2024-01-01\"}})"
```

## Configuration

Config is stored at `~/.config/dashmin/config.yaml`:

```yaml
apps:
  <app>:
    name: <app>
    type: postgres
    connection: "postgres://readonly:password@localhost:5432/<database>?sslmode=disable"
    queries:
      total_users: "SELECT COUNT(*) FROM users"
      signups_today: "SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE"
      revenue_today: "SELECT SUM(amount) FROM orders WHERE created_at >= CURRENT_DATE"
```

## Database Support

| Database   | Status | Connection String Example |
|------------|--------|---------------------------|
| PostgreSQL | ✅     | `postgres://user:pass@localhost:5432/database?sslmode=disable` |
| MySQL      | ✅     | `user:pass@tcp(localhost:3306)/database` |
| MongoDB    | ✅     | `mongodb://user:pass@localhost:27017/database` |

### MongoDB Query Format
MongoDB queries use the format: `collection.operation({filter})`

Examples:
- `users.count({})` - Count all users
- `users.count({"status": "active"})` - Count active users
- `orders.count({"date": {"$gte": "2024-01-01"}})` - Count recent orders

## Dashboard Features

- 🟢 **Live status indicators** - Green for success, red for errors, yellow for warnings
- 📊 **Real-time metrics** - Auto-refreshing data with timestamps
- 🔍 **Detailed views** - Drill down into query results
- ⌨️ **Keyboard shortcuts** - Efficient navigation and controls

### Keyboard Shortcuts

- `r` - Refresh data
- `↑/↓` - Navigate results
- `q` or `Ctrl+C` - Quit

## Troubleshooting

If you're having connection issues:

```bash
dashmin test <app-name>
```

## Examples

### Multiple Apps Setup
```bash
# Add production apps
dashmin add blogbuddy postgres "postgres://readonly:password@pg.prod.com:5432/blogbuddy_prod?sslmode=disable"
dashmin add pingbuddy mysql "readonly:pass@tcp(mysql.prod.com:3306)/pingbuddy"
dashmin add analytics mongodb "mongodb://readonly:pass@mongo.prod.com:27017/analytics"

# Add common queries to all SQL apps
dashmin query blogbuddy users "SELECT COUNT(*) FROM users"
dashmin query blogbuddy signups_today "SELECT COUNT(*) FROM users WHERE DATE(created_at) = CURDATE()"
dashmin query pingbuddy users "SELECT COUNT(*) FROM users"
dashmin query pingbuddy monitors "SELECT COUNT(*) FROM monitors WHERE status = 'active'"

# Add MongoDB queries
dashmin query analytics users "users.count({})"
dashmin query analytics active_sessions "sessions.count({\"status\": \"active\"})"

# View dashboard
dashmin status
```

## Development

```bash
# Clone
git clone https://github.com/lucasnevespereira/dashmin
cd dashmin

# Install dependencies
go mod tidy

# Run
go run main.go status

# Build
go build -o dashmin main.go
```

## Contributing

Contributions are welcome! Please read the [contributing guidelines](CONTRIBUTING.md) first.

## License

MIT License - see [LICENSE](LICENSE) for details.


**Like dashmin?** Star ⭐ the repo and [support me](https://github.com/lucasnevespereira) for more tools!
