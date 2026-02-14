package project_work_items

import (
	"database/sql"
	"testing"

	"github.com/momokii/go-rab-maker/backend/models"
	_ "modernc.org/sqlite"
)

// setupTestDB creates a temporary in-memory database for testing
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Create temporary database file
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

		CREATE TABLE projects (
			project_id INTEGER PRIMARY KEY,
			user_id INTEGER NOT NULL,
			project_name TEXT NOT NULL,
			created_at TEXT,
			updated_at TEXT,
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		);

		CREATE TABLE master_work_categories (
			category_id INTEGER PRIMARY KEY,
			user_id INTEGER,
			category_name TEXT NOT NULL,
			display_order INTEGER DEFAULT 0,
			created_at TEXT,
			updated_at TEXT
		);

		CREATE TABLE project_work_items (
			work_item_id INTEGER PRIMARY KEY,
			project_id INTEGER NOT NULL,
			category_id INTEGER,
			description TEXT NOT NULL,
			volume REAL NOT NULL,
			unit TEXT NOT NULL,
			ahsp_template_id INTEGER,
			created_at TEXT,
			updated_at TEXT,
			FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES master_work_categories(category_id)
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
			FOREIGN KEY (work_item_id) REFERENCES project_work_items(work_item_id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

// TestDeleteWorkItem_CascadesToCosts verifies that deleting a work item also deletes its costs
func TestDeleteWorkItem_CascadesToCosts(t *testing.T) {
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

	_, err = tx.Exec("INSERT INTO projects (project_id, user_id, project_name, created_at, updated_at) VALUES (1, 1, 'Test Project', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert project: %v", err)
	}

	_, err = tx.Exec("INSERT INTO project_work_items (work_item_id, project_id, description, volume, unit, created_at, updated_at) VALUES (1, 1, 'Test Work Item', 10.0, 'm', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert work item: %v", err)
	}

	// Insert multiple costs for this work item
	_, err = tx.Exec("INSERT INTO project_item_costs (cost_id, work_item_id, master_item_id, item_type, item_name, quantity_needed, unit_price_at_creation, total_cost, created_at) VALUES (1, 1, 1, 'material', 'Cement', 5.0, 100.0, 500.0, '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert cost 1: %v", err)
	}

	_, err = tx.Exec("INSERT INTO project_item_costs (cost_id, work_item_id, master_item_id, item_type, item_name, quantity_needed, unit_price_at_creation, total_cost, created_at) VALUES (2, 1, 2, 'labor', 'Worker', 2.0, 150.0, 300.0, '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert cost 2: %v", err)
	}

	_, err = tx.Exec("INSERT INTO project_item_costs (cost_id, work_item_id, master_item_id, item_type, item_name, quantity_needed, unit_price_at_creation, total_cost, created_at) VALUES (3, 1, 3, 'material', 'Sand', 10.0, 50.0, 500.0, '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert cost 3: %v", err)
	}

	// Verify initial state: 1 work item, 3 costs
	var workItemCount, costCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM project_work_items WHERE work_item_id = 1").Scan(&workItemCount)
	if err != nil || workItemCount != 1 {
		t.Fatalf("Expected 1 work item, got %d, err: %v", workItemCount, err)
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM project_item_costs WHERE work_item_id = 1").Scan(&costCount)
	if err != nil || costCount != 3 {
		t.Fatalf("Expected 3 costs, got %d, err: %v", costCount, err)
	}

	// Delete the work item using the repository
	repo := NewProjectWorkItemRepo()
	workItem := models.ProjectWorkItem{
		WorkItemId:     1,
		ProjectId:      1,
		Description:    "Test Work Item",
		Volume:         10.0,
		Unit:           "m",
		AHSPTemplateId: nil,
		CreatedAt:      "2024-01-01",
		UpdatedAt:      "2024-01-01",
	}

	err = repo.Delete(tx, workItem)
	if err != nil {
		t.Fatalf("Failed to delete work item: %v", err)
	}

	// Verify both work item AND costs are deleted
	err = tx.QueryRow("SELECT COUNT(*) FROM project_work_items WHERE work_item_id = 1").Scan(&workItemCount)
	if err != nil || workItemCount != 0 {
		t.Errorf("Expected 0 work items after delete, got %d, err: %v", workItemCount, err)
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM project_item_costs WHERE work_item_id = 1").Scan(&costCount)
	if err != nil || costCount != 0 {
		t.Errorf("Expected 0 costs after cascade delete, got %d, err: %v", costCount, err)
	}
}

// TestDeleteByProjectId_CascadesToCosts verifies that deleting all work items for a project also deletes their costs
func TestDeleteByProjectId_CascadesToCosts(t *testing.T) {
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

	_, err = tx.Exec("INSERT INTO projects (project_id, user_id, project_name, created_at, updated_at) VALUES (1, 1, 'Test Project', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert project: %v", err)
	}

	// Insert multiple work items for this project
	_, err = tx.Exec("INSERT INTO project_work_items (work_item_id, project_id, description, volume, unit, created_at, updated_at) VALUES (1, 1, 'Work Item 1', 10.0, 'm', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert work item 1: %v", err)
	}

	_, err = tx.Exec("INSERT INTO project_work_items (work_item_id, project_id, description, volume, unit, created_at, updated_at) VALUES (2, 1, 'Work Item 2', 20.0, 'm2', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert work item 2: %v", err)
	}

	// Insert costs for each work item
	_, err = tx.Exec("INSERT INTO project_item_costs (cost_id, work_item_id, master_item_id, item_type, item_name, quantity_needed, unit_price_at_creation, total_cost, created_at) VALUES (1, 1, 1, 'material', 'Cement', 5.0, 100.0, 500.0, '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert cost for work item 1: %v", err)
	}

	_, err = tx.Exec("INSERT INTO project_item_costs (cost_id, work_item_id, master_item_id, item_type, item_name, quantity_needed, unit_price_at_creation, total_cost, created_at) VALUES (2, 2, 2, 'labor', 'Worker', 2.0, 150.0, 300.0, '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert cost for work item 2: %v", err)
	}

	// Verify initial state: 2 work items, 2 costs
	var workItemCount, costCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM project_work_items WHERE project_id = 1").Scan(&workItemCount)
	if err != nil || workItemCount != 2 {
		t.Fatalf("Expected 2 work items, got %d, err: %v", workItemCount, err)
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM project_item_costs").Scan(&costCount)
	if err != nil || costCount != 2 {
		t.Fatalf("Expected 2 costs, got %d, err: %v", costCount, err)
	}

	// Delete all work items for the project using the repository
	repo := NewProjectWorkItemRepo()
	err = repo.DeleteByProjectId(tx, 1)
	if err != nil {
		t.Fatalf("Failed to delete work items by project ID: %v", err)
	}

	// Verify both work items AND all costs are deleted
	err = tx.QueryRow("SELECT COUNT(*) FROM project_work_items WHERE project_id = 1").Scan(&workItemCount)
	if err != nil || workItemCount != 0 {
		t.Errorf("Expected 0 work items after delete, got %d, err: %v", workItemCount, err)
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM project_item_costs").Scan(&costCount)
	if err != nil || costCount != 0 {
		t.Errorf("Expected 0 costs after cascade delete, got %d, err: %v", costCount, err)
	}
}

// TestGetProjectTotalCost verifies total cost calculation
func TestGetProjectTotalCost(t *testing.T) {
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

	_, err = tx.Exec("INSERT INTO projects (project_id, user_id, project_name, created_at, updated_at) VALUES (1, 1, 'Test Project', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert project: %v", err)
	}

	_, err = tx.Exec("INSERT INTO project_work_items (work_item_id, project_id, description, volume, unit, created_at, updated_at) VALUES (1, 1, 'Work Item 1', 10.0, 'm', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert work item: %v", err)
	}

	// Insert costs with total_cost = 500 + 300 + 200 = 1000
	_, err = tx.Exec("INSERT INTO project_item_costs (cost_id, work_item_id, master_item_id, item_type, item_name, quantity_needed, unit_price_at_creation, total_cost, created_at) VALUES (1, 1, 1, 'material', 'Cement', 5.0, 100.0, 500.0, '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert cost 1: %v", err)
	}

	_, err = tx.Exec("INSERT INTO project_item_costs (cost_id, work_item_id, master_item_id, item_type, item_name, quantity_needed, unit_price_at_creation, total_cost, created_at) VALUES (2, 1, 2, 'labor', 'Worker', 2.0, 150.0, 300.0, '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert cost 2: %v", err)
	}

	_, err = tx.Exec("INSERT INTO project_item_costs (cost_id, work_item_id, master_item_id, item_type, item_name, quantity_needed, unit_price_at_creation, total_cost, created_at) VALUES (3, 1, 3, 'material', 'Sand', 10.0, 20.0, 200.0, '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert cost 3: %v", err)
	}

	// Get total cost using repository
	repo := NewProjectWorkItemRepo()
	totalCost, err := repo.GetProjectTotalCost(tx, 1)
	if err != nil {
		t.Fatalf("Failed to get project total cost: %v", err)
	}

	// Expected total: 500 + 300 + 200 = 1000
	expectedTotal := 1000.0
	if totalCost != expectedTotal {
		t.Errorf("Expected total cost %f, got %f", expectedTotal, totalCost)
	}
}
