# dashmin

**Minimal dashboard for your apps**

A lightweight CLI tool to monitor multiple databases and applications from one place. Built for developers who want quick insights without the overhead.

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
dashmin add blogbuddy postgres "postgres://readonly:password@localhost:5432/blogbuddy_prod?sslmode=disable"
```

### 2. Add custom queries
```bash
dashmin query blogbuddy users "SELECT COUNT(*) FROM users"
dashmin query blogbuddy posts "SELECT COUNT(*) FROM posts WHERE created_at > NOW() - INTERVAL '30 days'"
```

### 3. View dashboard
```bash
dashmin status
```

## Usage

### Managing Apps
```bash
# Add an app
dashmin add <name> <type> <connection-string>
dashmin add myapp postgres "postgres://readonly:password@db.example.com:5432/myapp?sslmode=disable"
dashmin add webapp mysql "user:pass@tcp(localhost:3306)/webapp"
dashmin add analytics mongodb "mongodb://user:pass@localhost:27017/analytics"

# List apps
dashmin list

# Test database connection
dashmin test <app>

# Remove an app
dashmin remove <name>
```

### Managing Queries
```bash
# Add custom queries
dashmin query <app> <label> <query>
dashmin query myapp total_users "SELECT COUNT(*) FROM users"
dashmin query myapp revenue "SELECT SUM(amount) FROM payments WHERE DATE(created_at) = CURDATE()"

# MongoDB examples
dashmin query analytics users "users.count({})"
dashmin query analytics active_users "users.count({\"status\": \"active\"})"
```

### Viewing Data
```bash
# Interactive dashboard
dashmin status

# List configuration
dashmin list
dashmin config

# Debug connection issues
dashmin test <app>
```

## Configuration

Config is stored at `~/.config/dashmin/config.yaml`:

```yaml
apps:
  blogbuddy:
    name: blogbuddy
    type: postgres
    connection: "postgres://readonly:password@localhost:5432/blogbuddy_prod?sslmode=disable"
    queries:
      users: "SELECT COUNT(*) FROM users"
      posts: "SELECT COUNT(*) FROM posts WHERE created_at > NOW() - INTERVAL '30 days'"
  
  webapp:
    name: webapp
    type: mysql
    connection: "user:pass@tcp(localhost:3306)/webapp"
    queries:
      users: "SELECT COUNT(*) FROM users"
      revenue: "SELECT SUM(amount) FROM payments WHERE DATE(created_at) = CURDATE()"

  analytics:
    name: analytics
    type: mongodb
    connection: "mongodb://user:pass@localhost:27017/analytics"
    queries:
      users: "users.count({})"
      active: "users.count({\"status\": \"active\"})"
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
