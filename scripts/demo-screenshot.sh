#!/bin/bash
# Demo screenshot script - Creates isolated demo environment for clean screenshots
# This script is for development/screenshot purposes only

set -e

DEMO_DIR="$HOME/.config/dashmin/screenshot"
DB_FILE="$DEMO_DIR/demo.db"
SCREENSHOT_CONFIG="$HOME/.config/dashmin/screenshot-config.yaml"
REAL_CONFIG="$HOME/.config/dashmin/config.yaml"
BACKUP_CONFIG="$HOME/.config/dashmin/config.yaml.backup"

echo "🚀 Creating screenshot demo environment..."

# Create demo directory
mkdir -p "$DEMO_DIR"

# Check if SQLite is available
if ! command -v sqlite3 &> /dev/null; then
    echo "❌ Error: sqlite3 is required but not installed."
    exit 1
fi

# Clean up any existing screenshot data
rm -f "$DB_FILE"

echo "📊 Creating demo database with realistic SaaS metrics..."

# Create tables with deterministic, realistic data
sqlite3 "$DB_FILE" << 'EOF'
-- Users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    email TEXT NOT NULL,
    name TEXT NOT NULL,
    status TEXT DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME
);

-- Orders table
CREATE TABLE orders (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    amount DECIMAL(10,2),
    status TEXT DEFAULT 'completed',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Sessions table
CREATE TABLE sessions (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Metrics table
CREATE TABLE metrics (
    id INTEGER PRIMARY KEY,
    endpoint TEXT,
    response_time_ms INTEGER,
    recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert 12,847 users (11,234 active) with created dates spread over time
INSERT INTO users (id, status, created_at, last_login) 
SELECT 
    value,
    CASE WHEN value <= 11234 THEN 'active' ELSE 'inactive' END,
    datetime('now', '-' || ((12847 - value) * 1.5) || ' days'),
    datetime('now', '-' || ((value * 17) % 30) || ' days')
FROM generate_series(1, 12847);

-- Update last 156 users to have created_at = today (signups today)
UPDATE users 
SET created_at = datetime('now', 'start of day', '+' || ((id - 12692) * 5) || ' minutes')
WHERE id > 12691;

-- Insert 89 orders today with varied amounts totaling ~$9,341
INSERT INTO orders (user_id, amount, status, created_at)
VALUES 
(1245, 124.50, 'completed', datetime('now', 'start of day', '+1 hours')),
(3892, 89.99, 'completed', datetime('now', 'start of day', '+1 hours', '+15 minutes')),
(5234, 245.00, 'completed', datetime('now', 'start of day', '+2 hours')),
(6789, 67.50, 'completed', datetime('now', 'start of day', '+2 hours', '+20 minutes')),
(8345, 189.99, 'completed', datetime('now', 'start of day', '+3 hours')),
(9234, 145.50, 'completed', datetime('now', 'start of day', '+3 hours', '+10 minutes')),
(10567, 78.25, 'completed', datetime('now', 'start of day', '+4 hours')),
(11234, 234.00, 'completed', datetime('now', 'start of day', '+4 hours', '+30 minutes')),
(8456, 156.75, 'completed', datetime('now', 'start of day', '+5 hours')),
(7234, 98.50, 'completed', datetime('now', 'start of day', '+5 hours', '+25 minutes')),
(6234, 178.00, 'completed', datetime('now', 'start of day', '+6 hours')),
(5345, 134.50, 'completed', datetime('now', 'start of day', '+6 hours', '+40 minutes')),
(4456, 67.99, 'completed', datetime('now', 'start of day', '+7 hours')),
(3567, 198.00, 'completed', datetime('now', 'start of day', '+7 hours', '+15 minutes')),
(2678, 145.50, 'completed', datetime('now', 'start of day', '+8 hours'));

-- Insert remaining 74 orders
INSERT INTO orders (user_id, amount, status, created_at)
SELECT 
    abs(random() % 11234) + 1,
    CASE (abs(random()) % 10)
        WHEN 0 THEN 45.99
        WHEN 1 THEN 67.50
        WHEN 2 THEN 89.99
        WHEN 3 THEN 124.50
        WHEN 4 THEN 145.00
        WHEN 5 THEN 167.50
        WHEN 6 THEN 189.99
        WHEN 7 THEN 234.00
        WHEN 8 THEN 278.50
        ELSE 299.99
    END,
    'completed',
    datetime('now', 'start of day', '+' || (abs(random()) % 12) || ' hours')
FROM generate_series(1, 74);

-- Insert 1,456 active sessions
INSERT INTO sessions (user_id, expires_at, created_at)
SELECT 
    abs(random() % 11234) + 1,
    datetime('now', '+8 hours'),
    datetime('now', '-' || (abs(random()) % 120) || ' minutes')
FROM generate_series(1, 1456);

-- Insert metrics with avg ~45ms
INSERT INTO metrics (endpoint, response_time_ms, recorded_at)
SELECT 
    '/api/users',
    45 + (abs(random()) % 15) - 7,
    datetime('now', '-' || (abs(random()) % 5) || ' minutes')
FROM generate_series(1, 50);

INSERT INTO metrics (endpoint, response_time_ms, recorded_at)
SELECT 
    '/api/orders',
    42 + (abs(random()) % 12) - 6,
    datetime('now', '-' || (abs(random()) % 5) || ' minutes')
FROM generate_series(1, 50);
EOF

echo "⚙️  Creating isolated screenshot configuration..."

# Create screenshot config with logically ordered queries
cat > "$SCREENSHOT_CONFIG" << EOF
apps:
    production:
        name: production
        type: sqlite
        connection: "sqlite://$DB_FILE"
        queries:
            total_users: "SELECT COUNT(*) FROM users"
            active_users: "SELECT COUNT(*) FROM users WHERE status = 'active'"
            signups_today: "SELECT COUNT(*) FROM users WHERE date(created_at) = date('now')"
            orders_today: "SELECT COUNT(*) FROM orders WHERE date(created_at) = date('now')"
            revenue_today: "SELECT ROUND(SUM(amount), 2) FROM orders WHERE date(created_at) = date('now') AND status = 'completed'"
            avg_order_value: "SELECT ROUND(AVG(amount), 2) FROM orders WHERE status = 'completed'"
            active_sessions: "SELECT COUNT(*) FROM sessions WHERE expires_at > datetime('now')"
            response_time_ms: "SELECT CAST(AVG(response_time_ms) AS INTEGER) FROM metrics WHERE recorded_at > datetime('now', '-5 minutes')"
EOF

# Backup and replace real config
if [ -f "$REAL_CONFIG" ] && [ ! -f "$BACKUP_CONFIG" ]; then
    cp "$REAL_CONFIG" "$BACKUP_CONFIG"
fi

cp "$SCREENSHOT_CONFIG" "$REAL_CONFIG"

echo ""
echo "✅ Screenshot demo ready!"
echo ""
echo "📊 Demo data:"
sqlite3 "$DB_FILE" "SELECT '   • Total users: ' || COUNT(*) FROM users;"
sqlite3 "$DB_FILE" "SELECT '   • Active users: ' || COUNT(*) FROM users WHERE status = 'active';"
sqlite3 "$DB_FILE" "SELECT '   • Signups today: ' || COUNT(*) FROM users WHERE date(created_at) = date('now');"
sqlite3 "$DB_FILE" "SELECT '   • Orders today: ' || COUNT(*) FROM orders WHERE date(created_at) = date('now');"
sqlite3 "$DB_FILE" "SELECT '   • Revenue today: $' || ROUND(SUM(amount), 2) FROM orders WHERE date(created_at) = date('now') AND status = 'completed';"
sqlite3 "$DB_FILE" "SELECT '   • Active sessions: ' || COUNT(*) FROM sessions WHERE expires_at > datetime('now');"
echo ""
echo "🚀 Run: dashmin show"
echo ""
echo "🔄 Restore: ./scripts/restore-config.sh"
