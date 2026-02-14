package ahsptemplates

import (
	"database/sql"
	"testing"

	"github.com/momokii/go-rab-maker/backend/models"
	_ "modernc.org/sqlite"
)

// setupTestDB creates a temporary database for testing
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpDB := t.TempDir() + "/test.db"

	db, err := sql.Open("sqlite", "file:"+tmpDB)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Create test schema
	_, err = db.Exec(`
		CREATE TABLE users (
			user_id INTEGER PRIMARY KEY,
			username TEXT NOT NULL
		);

		CREATE TABLE ahsp_templates (
			template_id INTEGER PRIMARY KEY,
			user_id INTEGER,
			template_name TEXT NOT NULL,
			unit TEXT NOT NULL,
			created_at TEXT,
			updated_at TEXT
		);

		CREATE TABLE project_work_items (
			work_item_id INTEGER PRIMARY KEY,
			project_id INTEGER NOT NULL,
			ahsp_template_id INTEGER,
			description TEXT NOT NULL,
			volume REAL NOT NULL,
			unit TEXT NOT NULL,
			created_at TEXT,
			updated_at TEXT
		);

		CREATE TABLE ahsp_material_components (
			component_id INTEGER PRIMARY KEY,
			template_id INTEGER NOT NULL,
			material_id INTEGER NOT NULL,
			coefficient REAL NOT NULL,
			FOREIGN KEY (template_id) REFERENCES ahsp_templates(template_id)
		);

		CREATE TABLE ahsp_labor_components (
			component_id INTEGER PRIMARY KEY,
			template_id INTEGER NOT NULL,
			labor_type_id INTEGER NOT NULL,
			coefficient REAL NOT NULL,
			FOREIGN KEY (template_id) REFERENCES ahsp_templates(template_id)
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

// TestDeleteUnusedTemplate_Success verifies that an unused template can be deleted
func TestDeleteUnusedTemplate_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert test data
	_, err = tx.Exec("INSERT INTO users (user_id, username) VALUES (1, 'testuser')")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Create template (not used in any work items)
	_, err = tx.Exec("INSERT INTO ahsp_templates (template_id, user_id, template_name, unit, created_at, updated_at) VALUES (1, 1, 'Concrete Foundation', 'm3', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert template: %v", err)
	}

	// Create associated components
	_, err = tx.Exec("INSERT INTO ahsp_material_components (component_id, template_id, material_id, coefficient) VALUES (1, 1, 1, 1.5)")
	if err != nil {
		t.Fatalf("Failed to insert material component: %v", err)
	}

	_, err = tx.Exec("INSERT INTO ahsp_labor_components (component_id, template_id, labor_type_id, coefficient) VALUES (1, 1, 1, 2.0)")
	if err != nil {
		t.Fatalf("Failed to insert labor component: %v", err)
	}

	// Verify initial state: 1 template, 2 components
	var templateCount, materialCompCount, laborCompCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM ahsp_templates WHERE template_id = 1").Scan(&templateCount)
	if err != nil || templateCount != 1 {
		t.Fatalf("Expected 1 template, got %d, err: %v", templateCount, err)
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM ahsp_material_components WHERE template_id = 1").Scan(&materialCompCount)
	if err != nil || materialCompCount != 1 {
		t.Fatalf("Expected 1 material component, got %d, err: %v", materialCompCount, err)
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM ahsp_labor_components WHERE template_id = 1").Scan(&laborCompCount)
	if err != nil || laborCompCount != 1 {
		t.Fatalf("Expected 1 labor component, got %d, err: %v", laborCompCount, err)
	}

	// Delete the template using repository
	template := models.AHSPTemplate{
		TemplateId:   1,
		UserId:       1,
		TemplateName: "Concrete Foundation",
		Unit:         "m3",
		CreatedAt:    "2024-01-01",
		UpdatedAt:    "2024-01-01",
	}

	repo := NewAhspTemplatesRepo()
	err = repo.Delete(tx, template)
	if err != nil {
		t.Fatalf("Failed to delete unused template: %v", err)
	}

	// Verify template AND components are deleted
	err = tx.QueryRow("SELECT COUNT(*) FROM ahsp_templates WHERE template_id = 1").Scan(&templateCount)
	if err != nil || templateCount != 0 {
		t.Errorf("Expected 0 templates after delete, got %d, err: %v", templateCount, err)
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM ahsp_material_components WHERE template_id = 1").Scan(&materialCompCount)
	if err != nil || materialCompCount != 0 {
		t.Errorf("Expected 0 material components after cascade, got %d, err: %v", materialCompCount, err)
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM ahsp_labor_components WHERE template_id = 1").Scan(&laborCompCount)
	if err != nil || laborCompCount != 0 {
		t.Errorf("Expected 0 labor components after cascade, got %d, err: %v", laborCompCount, err)
	}
}

// TestDeleteUsedTemplate_ReturnsError verifies that a template used in work items cannot be deleted
func TestDeleteUsedTemplate_ReturnsError(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert test data
	_, err = tx.Exec("INSERT INTO users (user_id, username) VALUES (1, 'testuser')")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Create template
	_, err = tx.Exec("INSERT INTO ahsp_templates (template_id, user_id, template_name, unit, created_at, updated_at) VALUES (1, 1, 'Concrete Foundation', 'm3', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert template: %v", err)
	}

	// Create work item that USES this template
	_, err = tx.Exec("INSERT INTO project_work_items (work_item_id, project_id, ahsp_template_id, description, volume, unit, created_at, updated_at) VALUES (1, 1, 1, 'Foundation Work', 10.0, 'm3', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert work item: %v", err)
	}

	// Verify initial state: 1 template, 1 work item using it
	var templateCount, usageCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM ahsp_templates WHERE template_id = 1").Scan(&templateCount)
	if err != nil || templateCount != 1 {
		t.Fatalf("Expected 1 template, got %d, err: %v", templateCount, err)
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM project_work_items WHERE ahsp_template_id = 1").Scan(&usageCount)
	if err != nil || usageCount != 1 {
		t.Fatalf("Expected 1 work item using template, got %d, err: %v", usageCount, err)
	}

	// Try to delete the template (should FAIL)
	template := models.AHSPTemplate{
		TemplateId:   1,
		UserId:       1,
		TemplateName: "Concrete Foundation",
		Unit:         "m3",
		CreatedAt:    "2024-01-01",
		UpdatedAt:    "2024-01-01",
	}

	repo := NewAhspTemplatesRepo()
	err = repo.Delete(tx, template)
	if err == nil {
		t.Error("Expected error when deleting template that is in use, but got nil")
	}

	// Verify template still exists (was NOT deleted)
	err = tx.QueryRow("SELECT COUNT(*) FROM ahsp_templates WHERE template_id = 1").Scan(&templateCount)
	if err != nil || templateCount != 1 {
		t.Errorf("Expected template to still exist after failed delete, got count %d, err: %v", templateCount, err)
	}
}

// TestDeleteUsedTemplate_MultipleWorkItems verifies that deletion is blocked
// even when multiple work items use the template
func TestDeleteUsedTemplate_MultipleWorkItems(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert test data
	_, err = tx.Exec("INSERT INTO users (user_id, username) VALUES (1, 'testuser')")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Create template
	_, err = tx.Exec("INSERT INTO ahsp_templates (template_id, user_id, template_name, unit, created_at, updated_at) VALUES (1, 1, 'Concrete Foundation', 'm3', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert template: %v", err)
	}

	// Create MULTIPLE work items using this template
	_, err = tx.Exec("INSERT INTO project_work_items (work_item_id, project_id, ahsp_template_id, description, volume, unit, created_at, updated_at) VALUES (1, 1, 1, 'Foundation Work 1', 10.0, 'm3', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert work item 1: %v", err)
	}

	_, err = tx.Exec("INSERT INTO project_work_items (work_item_id, project_id, ahsp_template_id, description, volume, unit, created_at, updated_at) VALUES (2, 1, 1, 'Foundation Work 2', 15.0, 'm3', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert work item 2: %v", err)
	}

	_, err = tx.Exec("INSERT INTO project_work_items (work_item_id, project_id, ahsp_template_id, description, volume, unit, created_at, updated_at) VALUES (3, 1, 1, 'Foundation Work 3', 20.0, 'm3', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert work item 3: %v", err)
	}

	// Verify 3 work items use this template
	var usageCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM project_work_items WHERE ahsp_template_id = 1").Scan(&usageCount)
	if err != nil || usageCount != 3 {
		t.Fatalf("Expected 3 work items using template, got %d, err: %v", usageCount, err)
	}

	// Try to delete the template (should FAIL)
	template := models.AHSPTemplate{
		TemplateId:   1,
		UserId:       1,
		TemplateName: "Concrete Foundation",
		Unit:         "m3",
		CreatedAt:    "2024-01-01",
		UpdatedAt:    "2024-01-01",
	}

	repo := NewAhspTemplatesRepo()
	err = repo.Delete(tx, template)
	if err == nil {
		t.Error("Expected error when deleting template used in multiple work items, but got nil")
	}
}

// TestCreateAndUpdateTemplate verifies basic CRUD operations
func TestCreateAndUpdateTemplate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Create template
	createData := models.AHSPTemplateCreate{
		UserId:       1,
		TemplateName: "Wall Construction",
		Unit:         "m2",
	}

	repo := NewAhspTemplatesRepo()
	err = repo.Create(tx, createData)
	if err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Verify template was created
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM ahsp_templates WHERE template_name = 'Wall Construction'").Scan(&count)
	if err != nil || count != 1 {
		t.Fatalf("Expected 1 template, got %d, err: %v", count, err)
	}

	// Update template
	template := models.AHSPTemplate{
		TemplateId:   1,
		UserId:       1,
		TemplateName: "Wall Construction Updated",
		Unit:         "m2",
		CreatedAt:    "2024-01-01",
		UpdatedAt:    "2024-01-02",
	}

	err = repo.Update(tx, template)
	if err != nil {
		t.Fatalf("Failed to update template: %v", err)
	}

	// Verify template was updated
	var name string
	err = tx.QueryRow("SELECT template_name FROM ahsp_templates WHERE template_id = 1").Scan(&name)
	if err != nil || name != "Wall Construction Updated" {
		t.Errorf("Expected template name 'Wall Construction Updated', got '%s', err: %v", name, err)
	}
}
