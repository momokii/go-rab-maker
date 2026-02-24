package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	dbPath := "/home/kelanach/.config/RABMaker/databases/database.sqlite"

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Fatal("Database not found at:", dbPath)
	}

	// Open database
	db, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Check if unit column already exists
	var columnName string
	err = db.QueryRow("SELECT name FROM pragma_table_info('project_item_costs') WHERE name='unit'").Scan(&columnName)

	if err == sql.ErrNoRows {
		// Column doesn't exist, add it
		fmt.Println("Adding 'unit' column to project_item_costs table...")
		_, err := db.Exec("ALTER TABLE project_item_costs ADD COLUMN unit TEXT")
		if err != nil {
			log.Fatal("Failed to add column:", err)
		}
		fmt.Println("✓ Successfully added 'unit' column!")
	} else if err != nil {
		log.Fatal("Error checking column:", err)
	} else {
		fmt.Println("✓ 'unit' column already exists!")
	}

	// Verify the column was added
	rows, err := db.Query("PRAGMA table_info(project_item_costs)")
	if err != nil {
		log.Fatal("Error verifying schema:", err)
	}
	defer rows.Close()

	fmt.Println("\nCurrent schema for project_item_costs:")
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString
		rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		fmt.Printf("  - %s (%s)\n", name, ctype)
	}
}
