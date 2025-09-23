-- activate  Foreign Key
PRAGMA foreign_keys = ON;

--  users
CREATE TABLE IF NOT EXISTS users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

--  projects
CREATE TABLE IF NOT EXISTS projects (
    project_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL, -- ADDED
    project_name TEXT NOT NULL,
    location TEXT,
    client_name TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE -- ADDED
);

--  master_materials
CREATE TABLE IF NOT EXISTS master_materials (
    material_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER, -- ADDED (NULLABLE for system-wide defaults)
    material_name TEXT NOT NULL,
    unit TEXT NOT NULL,
    default_unit_price REAL NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, material_name), -- ADDED: Material name should be unique per user
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE -- ADDED
);

--  master_labor_types
CREATE TABLE IF NOT EXISTS master_labor_types (
    labor_type_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER, -- ADDED (NULLABLE for system-wide defaults)
    role_name TEXT NOT NULL,
    unit TEXT NOT NULL DEFAULT 'HOK',
    default_daily_wage REAL NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, role_name), -- ADDED: Role name should be unique per user
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE -- ADDED
);

--  master_work_categories
CREATE TABLE IF NOT EXISTS master_work_categories (
    category_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER, -- ADDED (NULLABLE for system-wide defaults)
    category_name TEXT NOT NULL,
    display_order INTEGER DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, category_name), -- ADDED: Category name should be unique per user
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE -- ADDED
);

--  ahsp_templates
CREATE TABLE IF NOT EXISTS ahsp_templates (
    template_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER, -- ADDED (NULLABLE for system-wide defaults)
    template_name TEXT NOT NULL,
    unit TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, template_name), -- ADDED: Template name should be unique per user
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE -- ADDED
);

--  ahsp_material_components
CREATE TABLE IF NOT EXISTS ahsp_material_components (
    component_id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INTEGER NOT NULL,
    material_id INTEGER NOT NULL,
    coefficient REAL NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES ahsp_templates(template_id) ON DELETE CASCADE,
    FOREIGN KEY (material_id) REFERENCES master_materials(material_id) ON DELETE RESTRICT
);

--  ahsp_labor_components
CREATE TABLE IF NOT EXISTS ahsp_labor_components (
    component_id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INTEGER NOT NULL,
    labor_type_id INTEGER NOT NULL,
    coefficient REAL NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (template_id) REFERENCES ahsp_templates(template_id) ON DELETE CASCADE,
    FOREIGN KEY (labor_type_id) REFERENCES master_labor_types(labor_type_id) ON DELETE RESTRICT
);

--  project_work_items
CREATE TABLE IF NOT EXISTS project_work_items (
    work_item_id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    description TEXT NOT NULL,
    volume REAL NOT NULL,
    unit TEXT NOT NULL,
    ahsp_template_id INTEGER,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES master_work_categories(category_id) ON DELETE RESTRICT,
    FOREIGN KEY (ahsp_template_id) REFERENCES ahsp_templates(template_id) ON DELETE SET NULL
);

--  project_item_costs
CREATE TABLE IF NOT EXISTS project_item_costs (
    cost_id INTEGER PRIMARY KEY AUTOINCREMENT,
    work_item_id INTEGER NOT NULL,
    item_type TEXT NOT NULL CHECK(item_type IN ('MATERIAL', 'LABOR')),
    master_item_id INTEGER NOT NULL,
    item_name TEXT NOT NULL,
    coefficient REAL NOT NULL,
    quantity_needed REAL NOT NULL,
    unit_price_at_creation REAL NOT NULL,
    total_cost REAL NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (work_item_id) REFERENCES project_work_items(work_item_id) ON DELETE CASCADE
);

-- index
CREATE INDEX idx_pwi_project_id ON project_work_items(project_id);
CREATE INDEX idx_pic_work_item_id ON project_item_costs(work_item_id);
CREATE INDEX idx_ahsp_mat_template_id ON ahsp_material_components(template_id);
CREATE INDEX idx_ahsp_lab_template_id ON ahsp_labor_components(template_id);