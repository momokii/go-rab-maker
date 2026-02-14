package master_materials

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
		CREATE TABLE master_materials (
			material_id INTEGER PRIMARY KEY,
			user_id INTEGER,
			material_name TEXT NOT NULL,
			unit TEXT NOT NULL,
			default_unit_price REAL NOT NULL,
			created_at TEXT,
			updated_at TEXT
		);

		CREATE TABLE ahsp_material_components (
			component_id INTEGER PRIMARY KEY,
			template_id INTEGER NOT NULL,
			material_id INTEGER NOT NULL,
			coefficient REAL NOT NULL,
			FOREIGN KEY (material_id) REFERENCES master_materials(material_id)
		);

		CREATE TABLE users (
			user_id INTEGER PRIMARY KEY,
			username TEXT NOT NULL
		);

		CREATE TABLE projects (
			project_id INTEGER PRIMARY KEY,
			user_id INTEGER NOT NULL,
			project_name TEXT NOT NULL,
			created_at TEXT,
			updated_at TEXT,
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		);

		CREATE TABLE project_work_items (
			work_item_id INTEGER PRIMARY KEY,
			project_id INTEGER NOT NULL,
			description TEXT NOT NULL,
			volume REAL NOT NULL,
			unit TEXT NOT NULL,
			created_at TEXT,
			updated_at TEXT,
			FOREIGN KEY (project_id) REFERENCES projects(project_id)
		);

		CREATE TABLE project_item_costs (
			cost_id INTEGER PRIMARY KEY,
			work_item_id INTEGER NOT NULL,
			master_item_id INTEGER,
			item_type TEXT NOT NULL,
			item_name TEXT NOT NULL,
			quantity_needed REAL NOT NULL,
			unit_price_at_creation REAL NOT NULL,
			total_cost REAL NOT NULL,
			created_at TEXT,
			FOREIGN KEY (work_item_id) REFERENCES project_work_items(work_item_id)
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

// TestUpdateMaterial_DoesNotModifyHistoricalCosts verifies that updating a material's price
// does NOT retroactively change project costs that were created with the old price
func TestUpdateMaterial_DoesNotModifyHistoricalCosts(t *testing.T) {
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

	// Create material with original price $100
	_, err = tx.Exec("INSERT INTO master_materials (material_id, user_id, material_name, unit, default_unit_price, created_at, updated_at) VALUES (1, 1, 'Cement', 'bag', 100.0, '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert material: %v", err)
	}

	// Create project and work item
	_, err = tx.Exec("INSERT INTO projects (project_id, user_id, project_name, created_at, updated_at) VALUES (1, 1, 'Test Project', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert project: %v", err)
	}

	_, err = tx.Exec("INSERT INTO project_work_items (work_item_id, project_id, description, volume, unit, created_at, updated_at) VALUES (1, 1, 'Foundation', 10.0, 'm3', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert work item: %v", err)
	}

	// Create project item cost with the ORIGINAL price ($100)
	_, err = tx.Exec("INSERT INTO project_item_costs (cost_id, work_item_id, master_item_id, item_type, item_name, quantity_needed, unit_price_at_creation, total_cost, created_at) VALUES (1, 1, 1, 'material', 'Cement', 5.0, 100.0, 500.0, '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert project item cost: %v", err)
	}

	// Verify the cost was created with original price ($100)
	var originalCost float64
	err = tx.QueryRow("SELECT unit_price_at_creation FROM project_item_costs WHERE cost_id = 1").Scan(&originalCost)
	if err != nil {
		t.Fatalf("Failed to query original cost: %v", err)
	}
	if originalCost != 100.0 {
		t.Fatalf("Expected original cost 100.0, got %f", originalCost)
	}

	// Update the material's price to $200 (doubled)
	materialData := models.MasterMaterial{
		MaterialId:        1,
		UserId:            1,
		MaterialName:      "Cement",
		Unit:              "bag",
		DefaultUnitPrice:  200.0, // Price doubled from $100 to $200
		CreatedAt:         "2024-01-01",
		UpdatedAt:         "2024-01-02",
	}

	repo := NewMasterMaterialsRepo()
	err = repo.Update(tx, materialData)
	if err != nil {
		t.Fatalf("Failed to update material: %v", err)
	}

	// CRITICAL TEST: Verify the historical cost is UNCHANGED
	// The project item cost should still be $100 (the price at creation time)
	var historicalCost float64
	err = tx.QueryRow("SELECT unit_price_at_creation FROM project_item_costs WHERE cost_id = 1").Scan(&historicalCost)
	if err != nil {
		t.Fatalf("Failed to query historical cost: %v", err)
	}

	if historicalCost != 100.0 {
		t.Errorf("Historical cost was modified! Expected 100.0 (original price), got %f. Historical costs must be preserved!", historicalCost)
	}

	// Verify the material's current price was updated
	var currentPrice float64
	err = tx.QueryRow("SELECT default_unit_price FROM master_materials WHERE material_id = 1").Scan(&currentPrice)
	if err != nil {
		t.Fatalf("Failed to query current material price: %v", err)
	}

	if currentPrice != 200.0 {
		t.Errorf("Expected current material price to be updated to 200.0, got %f", currentPrice)
	}
}

// TestUpdateMaterial_MultipleProjectsPreserveCosts verifies that multiple projects
// with the same material all preserve their individual historical costs
func TestUpdateMaterial_MultipleProjectsPreserveCosts(t *testing.T) {
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

	// Create material with original price $100
	_, err = tx.Exec("INSERT INTO master_materials (material_id, user_id, material_name, unit, default_unit_price, created_at, updated_at) VALUES (1, 1, 'Cement', 'bag', 100.0, '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert material: %v", err)
	}

	// Create TWO projects with the same material at the original price
	// Project 1
	_, err = tx.Exec("INSERT INTO projects (project_id, user_id, project_name, created_at, updated_at) VALUES (1, 1, 'Project A', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert project 1: %v", err)
	}
	_, err = tx.Exec("INSERT INTO project_work_items (work_item_id, project_id, description, volume, unit, created_at, updated_at) VALUES (1, 1, 'Foundation', 10.0, 'm3', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert work item 1: %v", err)
	}
	_, err = tx.Exec("INSERT INTO project_item_costs (cost_id, work_item_id, master_item_id, item_type, item_name, quantity_needed, unit_price_at_creation, total_cost, created_at) VALUES (1, 1, 1, 'material', 'Cement', 10.0, 100.0, 1000.0, '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert cost 1: %v", err)
	}

	// Project 2
	_, err = tx.Exec("INSERT INTO projects (project_id, user_id, project_name, created_at, updated_at) VALUES (2, 1, 'Project B', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert project 2: %v", err)
	}
	_, err = tx.Exec("INSERT INTO project_work_items (work_item_id, project_id, description, volume, unit, created_at, updated_at) VALUES (2, 2, 'Foundation', 5.0, 'm3', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert work item 2: %v", err)
	}
	_, err = tx.Exec("INSERT INTO project_item_costs (cost_id, work_item_id, master_item_id, item_type, item_name, quantity_needed, unit_price_at_creation, total_cost, created_at) VALUES (2, 2, 1, 'material', 'Cement', 8.0, 100.0, 800.0, '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert cost 2: %v", err)
	}

	// Update material price to $150
	materialData := models.MasterMaterial{
		MaterialId:       1,
		UserId:           1,
		MaterialName:     "Cement",
		Unit:             "bag",
		DefaultUnitPrice: 150.0,
		CreatedAt:        "2024-01-01",
		UpdatedAt:        "2024-01-02",
	}

	repo := NewMasterMaterialsRepo()
	err = repo.Update(tx, materialData)
	if err != nil {
		t.Fatalf("Failed to update material: %v", err)
	}

	// Verify BOTH projects still have their original costs preserved
	var cost1, cost2 float64
	err = tx.QueryRow("SELECT unit_price_at_creation FROM project_item_costs WHERE cost_id = 1").Scan(&cost1)
	if err != nil {
		t.Fatalf("Failed to query cost 1: %v", err)
	}
	err = tx.QueryRow("SELECT unit_price_at_creation FROM project_item_costs WHERE cost_id = 2").Scan(&cost2)
	if err != nil {
		t.Fatalf("Failed to query cost 2: %v", err)
	}

	if cost1 != 100.0 || cost2 != 100.0 {
		t.Errorf("Historical costs were modified! Expected both to be 100.0, got cost1=%f, cost2=%f", cost1, cost2)
	}
}

// TestCreateAndDeleteMaterial verifies basic CRUD operations
func TestCreateAndDeleteMaterial(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Create material
	createData := models.MasterMaterialCreate{
		MaterialName:     "Sand",
		Unit:             "m3",
		DefaultUnitPrice: 50.0,
		UserId:           1,
	}

	repo := NewMasterMaterialsRepo()
	err = repo.Create(tx, createData)
	if err != nil {
		t.Fatalf("Failed to create material: %v", err)
	}

	// Add timestamps to the created record
	_, err = tx.Exec("UPDATE master_materials SET created_at = '2024-01-01', updated_at = '2024-01-01' WHERE material_id = 1")
	if err != nil {
		t.Fatalf("Failed to update timestamps: %v", err)
	}

	// Verify material was created
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM master_materials WHERE material_name = 'Sand'").Scan(&count)
	if err != nil || count != 1 {
		t.Fatalf("Expected 1 material, got %d, err: %v", count, err)
	}

	// Find and delete the material
	material, err := repo.FindById(tx, 1)
	if err != nil {
		t.Fatalf("Failed to find material: %v", err)
	}

	err = repo.Delete(tx, material)
	if err != nil {
		t.Fatalf("Failed to delete material: %v", err)
	}

	// Verify material was deleted
	err = tx.QueryRow("SELECT COUNT(*) FROM master_materials WHERE material_id = 1").Scan(&count)
	if err != nil || count != 0 {
		t.Errorf("Expected 0 materials after delete, got %d, err: %v", count, err)
	}
}
