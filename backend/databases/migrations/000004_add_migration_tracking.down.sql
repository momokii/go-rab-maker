-- Rollback: Remove schema_migrations tracking table

-- Drop the migration tracking table
DROP TABLE IF EXISTS schema_migrations;
