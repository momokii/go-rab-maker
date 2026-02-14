-- Migration: Add soft delete column to users table
-- Purpose: Enable soft deletion of users for data retention and compliance

-- Add soft delete column to users table
ALTER TABLE users ADD COLUMN deleted_at DATETIME DEFAULT NULL;

-- Create index for soft delete queries (improves performance of filtering active users)
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Update unique constraint on username to exclude deleted users
-- Drop old constraint
DROP INDEX IF EXISTS idx_users_username;

-- Create new partial index (only non-deleted users must have unique usernames)
-- This allows reuse of usernames from deleted accounts while maintaining uniqueness for active users
CREATE UNIQUE INDEX idx_users_username_active ON users(username) WHERE deleted_at IS NULL;
