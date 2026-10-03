#!/bin/bash
# Restore real config after screenshot mode

set -e

REAL_CONFIG="$HOME/.config/dashmin/config.yaml"
BACKUP_CONFIG="$HOME/.config/dashmin/config.yaml.backup"

echo "🔄 Restoring real configuration..."

if [ -f "$BACKUP_CONFIG" ]; then
    cp "$BACKUP_CONFIG" "$REAL_CONFIG"
    rm "$BACKUP_CONFIG"
    echo "✅ Real configuration restored!"
    echo ""
    echo "Your apps:"
    dashmin app list 2>/dev/null | head -20 || echo "Run 'dashmin app list' to see your apps"
else
    echo "⚠️  No backup found. Config may already be restored."
fi
