-- Remove unit column from project_item_costs table
-- Rollback for migration 000004

ALTER TABLE project_item_costs DROP COLUMN unit;
