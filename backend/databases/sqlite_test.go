package databases

import (
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

// TestForeignKeysEnabled verifies that foreign keys are enabled on database connections
func TestForeignKeysEnabled(t *testing.T) {
	// Create a temporary database file for testing
	tmpDB := t.TempDir() + "/test_foreign_keys.db"
	defer os.Remove(tmpDB)

	// Open database connection using the same method as production
	db, err := sql.Open("sqlite", "file:"+tmpDB)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// CRITICAL: Enable foreign keys (must match production code)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Verify foreign keys are enabled
	var result int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&result)
	if err != nil {
		t.Fatalf("Failed to query foreign_keys pragma: %v", err)
	}

	if result != 1 {
		t.Errorf("Foreign keys are not enabled: expected 1, got %d", result)
	}
}

// TestForeignKeyCascade verifies that foreign key cascade operations work correctly
func TestForeignKeyCascade(t *testing.T) {
	// Create a temporary database
	tmpDB := t.TempDir() + "/test_cascade.db"
	defer os.Remove(tmpDB)

	db, err := sql.Open("sqlite", "file:"+tmpDB)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Create test tables with CASCADE
	_, err = db.Exec(`
		CREATE TABLE users (
			user_id INTEGER PRIMARY KEY,
			username TEXT NOT NULL
		);

		CREATE TABLE projects (
			project_id INTEGER PRIMARY KEY,
			user_id INTEGER NOT NULL,
			project_name TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
		);

		CREATE TABLE project_work_items (
			work_item_id INTEGER PRIMARY KEY,
			project_id INTEGER NOT NULL,
			description TEXT NOT NULL,
			volume REAL NOT NULL,
			unit TEXT NOT NULL,
			FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Insert test data
	result, err := db.Exec("INSERT INTO users (user_id, username) VALUES (1, 'testuser')")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}
	userID, _ := result.LastInsertId()

	result, err = db.Exec("INSERT INTO projects (project_id, user_id, project_name) VALUES (1, ?, 'Test Project')", userID)
	if err != nil {
		t.Fatalf("Failed to insert project: %v", err)
	}
	projectID, _ := result.LastInsertId()

	result, err = db.Exec("INSERT INTO project_work_items (work_item_id, project_id, description, volume, unit) VALUES (1, ?, 'Test Work Item', 10.0, 'm')", projectID)
	if err != nil {
		t.Fatalf("Failed to insert work item: %v", err)
	}

	// Verify initial data exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil || count != 1 {
		t.Fatalf("Expected 1 user, got %d, err: %v", count, err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count)
	if err != nil || count != 1 {
		t.Fatalf("Expected 1 project, got %d, err: %v", count, err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM project_work_items").Scan(&count)
	if err != nil || count != 1 {
		t.Fatalf("Expected 1 work item, got %d, err: %v", count, err)
	}

	// Delete user (should cascade to projects and work items)
	_, err = db.Exec("DELETE FROM users WHERE user_id = 1")
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify cascade worked - all records should be deleted
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil || count != 0 {
		t.Errorf("Expected 0 users after cascade, got %d, err: %v", count, err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count)
	if err != nil || count != 0 {
		t.Errorf("Expected 0 projects after cascade, got %d, err: %v", count, err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM project_work_items").Scan(&count)
	if err != nil || count != 0 {
		t.Errorf("Expected 0 work items after cascade, got %d, err: %v", count, err)
	}
}

// TestForeignKeyRestrict verifies that RESTRICT constraints prevent deletion
func TestForeignKeyRestrict(t *testing.T) {
	// Create a temporary database
	tmpDB := t.TempDir() + "/test_restrict.db"
	defer os.Remove(tmpDB)

	db, err := sql.Open("sqlite", "file:"+tmpDB)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Create test tables with RESTRICT
	_, err = db.Exec(`
		CREATE TABLE master_materials (
			material_id INTEGER PRIMARY KEY,
			material_name TEXT NOT NULL
		);

		CREATE TABLE ahsp_material_components (
			component_id INTEGER PRIMARY KEY,
			material_id INTEGER NOT NULL,
			coefficient REAL NOT NULL,
			FOREIGN KEY (material_id) REFERENCES master_materials(material_id) ON DELETE RESTRICT
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Insert test data
	_, err = db.Exec("INSERT INTO master_materials (material_id, material_name) VALUES (1, 'Cement')")
	if err != nil {
		t.Fatalf("Failed to insert material: %v", err)
	}

	_, err = db.Exec("INSERT INTO ahsp_material_components (component_id, material_id, coefficient) VALUES (1, 1, 1.5)")
	if err != nil {
		t.Fatalf("Failed to insert component: %v", err)
	}

	// Try to delete material (should fail due to RESTRICT)
	_, err = db.Exec("DELETE FROM master_materials WHERE material_id = 1")
	if err == nil {
		t.Error("Expected error when deleting material with RESTRICT constraint, but got nil")
	}

	// Verify the material still exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM master_materials WHERE material_id = 1").Scan(&count)
	if err != nil || count != 1 {
		t.Errorf("Expected material to still exist after failed delete, got count %d, err: %v", count, err)
	}
}
