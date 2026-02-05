# Dashmin Roadmap

Features and improvements planned for dashmin.

## Current State (v0.2.0)

- Multi-database support (PostgreSQL, MySQL, MongoDB)
- Custom SQL/MongoDB queries
- AI-powered query generation
- Terminal dashboard with manual refresh

---

## Short Term (v0.3.0)

### Auto-refresh
```bash
dashmin show --watch          # Refresh every 30s (default)
dashmin show --watch 10s      # Refresh every 10s
```

### Change indicators
Show if a value increased or decreased since last refresh:
```
APP            QUERY               VALUE          CHANGE
myapp          total_users         1,234          ↑ +12
myapp          errors_today        5              ↓ -3
```

### Environment variable support
Avoid storing credentials in config file:
```yaml
connection: "${DATABASE_URL}"
# or
connection: "postgres://${DB_USER}:${DB_PASS}@localhost/myapp"
```

### Query validation on add
Test query when adding to catch errors early:
```bash
dashmin query add myapp users "SELECT COUNT(*) FROM userss"
# Error: relation "userss" does not exist
# Query not added. Use --force to add anyway.
```

---

## Medium Term (v0.4.0)

### Comparison queries
Compare current value with previous period:
```bash
dashmin query add myapp signups "SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE" --compare yesterday
```

Dashboard shows:
```
QUERY               TODAY      YESTERDAY    DIFF
signups             45         38           +18%
```

### Simple alerts
Highlight values that cross a threshold:
```bash
dashmin query add myapp errors "SELECT COUNT(*) FROM logs WHERE level='error'" --warn 10 --critical 50
```

Dashboard shows values in yellow (warn) or red (critical).

### Export to JSON
```bash
dashmin show --json > metrics.json
dashmin show myapp --json
```

Useful for:
- Piping to other tools (`jq`, scripts)
- Sending to monitoring systems
- Storing snapshots

### Query templates
Pre-built queries for common metrics:
```bash
dashmin query add myapp users --template count --table users
dashmin query add myapp signups --template count-today --table users --date-column created_at
dashmin query add myapp revenue --template sum-today --table orders --column amount
```

---

## Long Term (v1.0.0)

### Multiple config profiles
Switch between different environments:
```bash
dashmin --profile prod show
dashmin --profile staging show
```

### Shareable dashboards
Export/import dashboard configurations:
```bash
dashmin export > myapp-dashboard.yaml
dashmin import myapp-dashboard.yaml
```

### SQLite support
For local development databases.

### Redis support
Monitor cache metrics:
```bash
dashmin query add cache keys "DBSIZE"
dashmin query add cache memory "INFO memory"
```

### Notification hooks
Run a command when a metric crosses a threshold:
```yaml
queries:
  errors_today:
    query: "SELECT COUNT(*) FROM logs WHERE level='error'"
    critical: 100
    on_critical: "curl -X POST https://slack.webhook/..."
```

---

## Not Planned

These are out of scope to keep dashmin minimal:

- **Graphs/charts** - Use Grafana for that
- **Historical data storage** - Use a proper time-series DB
- **Multi-user/auth** - This is a personal dev tool
- **Web interface** - Terminal-first, always
- **Complex alerting rules** - Use PagerDuty/OpsGenie

---

## Contributing

Want to work on a feature?

1. Check if there's an existing issue
2. Open an issue to discuss the approach
3. Submit a PR

Small, focused PRs are preferred over large changes.
