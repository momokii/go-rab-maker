#!/bin/sh
set -e

# Fix database directory permissions
# The container starts as root, but will run the app as appuser (uid 1000)
# Docker may create the volume directory as root, so we need to fix ownership
DB_DIR="/app/databases"

# Ensure the database directory exists and has proper permissions
echo "Setting up database directory: $DB_DIR"
mkdir -p "$DB_DIR"

# Fix ownership to appuser (uid 1000, gid 1000)
# This ensures the appuser can write to the database directory
chown -R 1000:1000 "$DB_DIR" 2>/dev/null || {
    # If chown fails (e.g., on a bind mount from host), make it world-writable
    chmod 777 "$DB_DIR"
    echo "Made database directory world-writable (could not change ownership)"
}

# Switch to appuser and run the application
exec su-exec appuser "$@"
