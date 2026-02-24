package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "file:/home/kelanach/.config/RABMaker/databases/database.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check if column exists
	var result string
	err = db.QueryRow("SELECT name FROM pragma_table_info('project_item_costs') WHERE name='unit'").Scan(&result)

	if err == sql.ErrNoRows {
		fmt.Println("Adding 'unit' column...")
		_, err = db.Exec("ALTER TABLE project_item_costs ADD COLUMN unit TEXT")
		if err != nil {
			log.Fatal("Failed:", err)
		}
		fmt.Println("✓ Done!")
	} else if err == nil {
		fmt.Println("✓ Column already exists!")
	} else {
		log.Fatal("Error:", err)
	}
}
