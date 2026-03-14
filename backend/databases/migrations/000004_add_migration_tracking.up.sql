-- Migration: Add schema_migrations tracking table
-- Purpose: Track which migrations have been applied to enable incremental updates

-- Create table to track applied migrations
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
