-- Aktifkan penegakan Foreign Key
PRAGMA foreign_keys = OFF;

-- Drop tables if they exist
DROP TABLE IF EXISTS project_item_costs;
DROP TABLE IF EXISTS project_work_items;
DROP TABLE IF EXISTS ahsp_labor_components;
DROP TABLE IF EXISTS ahsp_material_components;
DROP TABLE IF EXISTS ahsp_templates;
DROP TABLE IF EXISTS master_work_categories;
DROP TABLE IF EXISTS master_labor_types;
DROP TABLE IF EXISTS master_materials;
DROP TABLE IF EXISTS projects;

PRAGMA foreign_keys = ON;
