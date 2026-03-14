-- Migration: Add schema_migrations tracking table
-- Purpose: Track which migrations have been applied to enable incremental updates

-- Create table to track applied migrations
CREATE TABLE IF NOT EXISTS test_new (
    name TEXT NOT NULL
);
