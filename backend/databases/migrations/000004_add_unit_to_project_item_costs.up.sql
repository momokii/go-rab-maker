-- Add unit column to project_item_costs table
-- This allows manual cost entries to store the unit for each item

ALTER TABLE project_item_costs ADD COLUMN unit TEXT;
