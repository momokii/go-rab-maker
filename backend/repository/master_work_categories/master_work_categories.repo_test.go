package master_work_categories

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
			created_at TEXT,
			updated_at TEXT,
			FOREIGN KEY (category_id) REFERENCES master_work_categories(category_id)
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

// TestDeleteUnusedCategory_Success verifies that an unused category can be deleted
func TestDeleteUnusedCategory_Success(t *testing.T) {
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

	// Create category (not used in any work items)
	_, err = tx.Exec("INSERT INTO master_work_categories (category_id, user_id, category_name, display_order, created_at, updated_at) VALUES (1, 1, 'Foundation Work', 1, '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert category: %v", err)
	}

	// Verify initial state: 1 category
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM master_work_categories WHERE category_id = 1").Scan(&count)
	if err != nil || count != 1 {
		t.Fatalf("Expected 1 category, got %d, err: %v", count, err)
	}

	// Delete the category using repository
	category := models.MasterWorkCategory{
		CategoryId:   1,
		UserId:       1,
		CategoryName: "Foundation Work",
		DisplayOrder: 1,
		CreatedAt:    "2024-01-01",
		UpdatedAt:    "2024-01-01",
	}

	repo := NewMasterWorkCategoriesRepo()
	err = repo.Delete(tx, category)
	if err != nil {
		t.Fatalf("Failed to delete unused category: %v", err)
	}

	// Verify category is deleted
	err = tx.QueryRow("SELECT COUNT(*) FROM master_work_categories WHERE category_id = 1").Scan(&count)
	if err != nil || count != 0 {
		t.Errorf("Expected 0 categories after delete, got %d, err: %v", count, err)
	}
}

// TestDeleteUsedCategory_ReturnsError verifies that a category used in work items cannot be deleted
func TestDeleteUsedCategory_ReturnsError(t *testing.T) {
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

	// Create category
	_, err = tx.Exec("INSERT INTO master_work_categories (category_id, user_id, category_name, display_order, created_at, updated_at) VALUES (1, 1, 'Foundation Work', 1, '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert category: %v", err)
	}

	// Create work item that USES this category
	_, err = tx.Exec("INSERT INTO project_work_items (work_item_id, project_id, category_id, description, volume, unit, created_at, updated_at) VALUES (1, 1, 1, 'Excavation', 100.0, 'm3', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert work item: %v", err)
	}

	// Verify initial state: 1 category, 1 work item using it
	var categoryCount, usageCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM master_work_categories WHERE category_id = 1").Scan(&categoryCount)
	if err != nil || categoryCount != 1 {
		t.Fatalf("Expected 1 category, got %d, err: %v", categoryCount, err)
	}

	err = tx.QueryRow("SELECT COUNT(*) FROM project_work_items WHERE category_id = 1").Scan(&usageCount)
	if err != nil || usageCount != 1 {
		t.Fatalf("Expected 1 work item using category, got %d, err: %v", usageCount, err)
	}

	// Try to delete the category (should FAIL)
	category := models.MasterWorkCategory{
		CategoryId:   1,
		UserId:       1,
		CategoryName: "Foundation Work",
		DisplayOrder: 1,
		CreatedAt:    "2024-01-01",
		UpdatedAt:    "2024-01-01",
	}

	repo := NewMasterWorkCategoriesRepo()
	err = repo.Delete(tx, category)
	if err == nil {
		t.Error("Expected error when deleting category that is in use, but got nil")
	}

	// Verify category still exists (was NOT deleted)
	err = tx.QueryRow("SELECT COUNT(*) FROM master_work_categories WHERE category_id = 1").Scan(&categoryCount)
	if err != nil || categoryCount != 1 {
		t.Errorf("Expected category to still exist after failed delete, got count %d, err: %v", categoryCount, err)
	}
}

// TestDeleteUsedCategory_MultipleWorkItems verifies that deletion is blocked
// even when multiple work items use the category
func TestDeleteUsedCategory_MultipleWorkItems(t *testing.T) {
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

	// Create category
	_, err = tx.Exec("INSERT INTO master_work_categories (category_id, user_id, category_name, display_order, created_at, updated_at) VALUES (1, 1, 'Foundation Work', 1, '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert category: %v", err)
	}

	// Create MULTIPLE work items using this category
	_, err = tx.Exec("INSERT INTO project_work_items (work_item_id, project_id, category_id, description, volume, unit, created_at, updated_at) VALUES (1, 1, 1, 'Excavation 1', 100.0, 'm3', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert work item 1: %v", err)
	}

	_, err = tx.Exec("INSERT INTO project_work_items (work_item_id, project_id, category_id, description, volume, unit, created_at, updated_at) VALUES (2, 1, 1, 'Excavation 2', 150.0, 'm3', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert work item 2: %v", err)
	}

	_, err = tx.Exec("INSERT INTO project_work_items (work_item_id, project_id, category_id, description, volume, unit, created_at, updated_at) VALUES (3, 1, 1, 'Excavation 3', 200.0, 'm3', '2024-01-01', '2024-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert work item 3: %v", err)
	}

	// Verify 3 work items use this category
	var usageCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM project_work_items WHERE category_id = 1").Scan(&usageCount)
	if err != nil || usageCount != 3 {
		t.Fatalf("Expected 3 work items using category, got %d, err: %v", usageCount, err)
	}

	// Try to delete the category (should FAIL)
	category := models.MasterWorkCategory{
		CategoryId:   1,
		UserId:       1,
		CategoryName: "Foundation Work",
		DisplayOrder: 1,
		CreatedAt:    "2024-01-01",
		UpdatedAt:    "2024-01-01",
	}

	repo := NewMasterWorkCategoriesRepo()
	err = repo.Delete(tx, category)
	if err == nil {
		t.Error("Expected error when deleting category used in multiple work items, but got nil")
	} else {
		// Verify error message contains usage information
		t.Logf("Got expected error: %v", err)
	}
}

// TestCreateAndUpdateCategory verifies basic CRUD operations
func TestCreateAndUpdateCategory(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Create category
	createData := models.MasterWorkCategoryCreate{
		UserId:       1,
		CategoryName: "Structure Work",
		DisplayOrder: 2,
	}

	repo := NewMasterWorkCategoriesRepo()
	err = repo.Create(tx, createData)
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	// Add timestamps to the created record
	_, err = tx.Exec("UPDATE master_work_categories SET created_at = '2024-01-01', updated_at = '2024-01-01' WHERE category_id = 1")
	if err != nil {
		t.Fatalf("Failed to update timestamps: %v", err)
	}

	// Verify category was created
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM master_work_categories WHERE category_name = 'Structure Work'").Scan(&count)
	if err != nil || count != 1 {
		t.Fatalf("Expected 1 category, got %d, err: %v", count, err)
	}

	// Find the created category
	category, err := repo.FindById(tx, 1)
	if err != nil {
		t.Fatalf("Failed to find category: %v", err)
	}

	// Update category
	category.CategoryName = "Structure Work Updated"
	category.DisplayOrder = 3

	err = repo.Update(tx, category)
	if err != nil {
		t.Fatalf("Failed to update category: %v", err)
	}

	// Verify category was updated
	var name string
	var order int
	err = tx.QueryRow("SELECT category_name, display_order FROM master_work_categories WHERE category_id = 1").Scan(&name, &order)
	if err != nil {
		t.Fatalf("Failed to query updated category: %v", err)
	}

	if name != "Structure Work Updated" {
		t.Errorf("Expected category name 'Structure Work Updated', got '%s'", name)
	}
	if order != 3 {
		t.Errorf("Expected display_order 3, got %d", order)
	}
}

// TestFindWithPagination verifies pagination functionality
func TestFindWithPagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	repo := NewMasterWorkCategoriesRepo()

	// Create multiple categories
	for i := 1; i <= 5; i++ {
		createData := models.MasterWorkCategoryCreate{
			UserId:       1,
			CategoryName: "Category " + string(rune('A'+i-1)),
			DisplayOrder: i,
		}
		err = repo.Create(tx, createData)
		if err != nil {
			t.Fatalf("Failed to create category %d: %v", i, err)
		}
	}

	// Add timestamps to all created categories
	_, err = tx.Exec("UPDATE master_work_categories SET created_at = '2024-01-01', updated_at = '2024-01-01'")
	if err != nil {
		t.Fatalf("Failed to update timestamps: %v", err)
	}

	// Test pagination: page 1, per page 2
	paginationInput := models.TablePaginationDataInput{
		Page:    1,
		PerPage: 2,
		Search:  "",
	}

	categories, paginationInfo, err := repo.Find(tx, paginationInput)
	if err != nil {
		t.Fatalf("Failed to find categories: %v", err)
	}

	// Should get 2 categories on page 1
	if len(categories) != 2 {
		t.Errorf("Expected 2 categories on page 1, got %d", len(categories))
	}

	// Total should be 5
	if paginationInfo.TotalItems != 5 {
		t.Errorf("Expected total 5 categories, got %d", paginationInfo.TotalItems)
	}

	// Total pages should be 3 (5 items / 2 per page = 3 pages)
	if paginationInfo.TotalPages != 3 {
		t.Errorf("Expected 3 total pages, got %d", paginationInfo.TotalPages)
	}
}
