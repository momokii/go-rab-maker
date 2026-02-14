-- Rollback: Remove soft delete functionality from users table

-- Drop the new partial unique index
DROP INDEX IF EXISTS idx_users_username_active;

-- Recreate the original simple unique index on username
CREATE UNIQUE INDEX idx_users_username ON users(username);

-- Drop the soft delete index
DROP INDEX IF EXISTS idx_users_deleted_at;

-- Remove the soft delete column
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;
