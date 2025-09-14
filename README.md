# dashmin

**Minimal dashboard for monitoring your apps from the terminal**

Quick insights without the bloat. Monitor your databases with simple queries - no complex setup, no heavy interfaces.

## Features

- **Multi-database support** - PostgreSQL, MySQL, MongoDB
- **Custom queries** - Track the metrics that matter to you
- **Minimal setup** - Simple commands, no complex configuration

## Installation

```bash
go install github.com/lucasnevespereira/dashmin@latest
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

## Commands

```bash
dashmin add <app> <type> <connection>     # Add new app
dashmin query <app> <label> <query>       # Add query to track
dashmin all                               # View all apps
dashmin see <app>                         # View specific app
dashmin list                              # Show configuration
dashmin test <app>                        # Test connection
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

### Keyboard Shortcuts

- `r` - Refresh data
- `q` - Quit

## Troubleshooting

If you're having connection issues:

```bash
dashmin test <app-name>
```


## Development

```bash
git clone https://github.com/lucasnevespereira/dashmin
cd dashmin
go mod tidy
go run main.go all
```


## License

MIT License - see [LICENSE](LICENSE) for details.


**Like dashmin?** Star ⭐ the repo and [support me](https://github.com/lucasnevespereira) for more tools!
