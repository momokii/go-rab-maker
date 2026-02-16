package databases

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/momokii/go-rab-maker/backend/utils"
	_ "modernc.org/sqlite"
)

const (
	DATABASE_FOLDER_NAME = "databases"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var (
	DATABASE_SQLITE_FOLDERS string
	DATABASE_SQLITE_PATH    string
)

func init() {
	baseDir := utils.GetBaseDir()
	DATABASE_SQLITE_FOLDERS = filepath.Join(baseDir, DATABASE_FOLDER_NAME)
	DATABASE_SQLITE_PATH = filepath.Join(DATABASE_SQLITE_FOLDERS, "database.sqlite")
}

type SQLiteServices interface {
	GetDB() *SQLiteDB

	Transaction(ctx context.Context, fn func(tx *sql.Tx) (statusCode int, err error)) (statusCode int, err error)
}

type SQLiteDB struct {
	DatabasesPath string
	read          *sql.DB
	Write         *sql.DB // Exported field for access
}

func runMigrations(db *SQLiteDB) error {
	// Read all migration files
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Execute each migration file in order
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		migrationPath := "migrations/" + entry.Name()
		content, err := migrationsFS.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		// Split content by semicolons to execute each statement separately
		statements := splitSQL(string(content))

		// Run each statement in transaction for atomicity
		statusCode, _ := db.Transaction(context.Background(), func(tx *sql.Tx) (int, error) {
			for _, stmt := range statements {
				if stmt == "" {
					continue
				}
				_, err := tx.Exec(stmt)
				if err != nil {
					return http.StatusInternalServerError, fmt.Errorf("failed to execute statement in migration %s: %w", entry.Name(), err)
				}
			}
			return http.StatusOK, nil
		})

		if statusCode != http.StatusOK && statusCode != http.StatusAccepted {
			return fmt.Errorf("migration %s failed with status code: %d", entry.Name(), statusCode)
		}
	}

	return nil
}

// splitSQL splits SQL content by semicolons, ignoring empty statements
func splitSQL(sql string) []string {
	var statements []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, ch := range sql {
		switch {
		case (ch == '\'' || ch == '"' || ch == '`') && !inQuote:
			inQuote = true
			quoteChar = ch
			current.WriteRune(ch)
		case ch == quoteChar && inQuote:
			inQuote = false
			current.WriteRune(ch)
		case ch == ';' && !inQuote:
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}

	// Add the last statement if exists
	if stmt := strings.TrimSpace(current.String()); stmt != "" {
		statements = append(statements, stmt)
	}

	return statements
}

func NewSQLiteDatabases(databasesPath string) (SQLiteServices, error) {
	// setup for read database
	write, err := sql.Open("sqlite", "file:"+databasesPath)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := write.Ping(); err != nil {
		return nil, err
	}

	// only single writer to avoid SQLITE_BUSY
	write.SetMaxOpenConns(1)

	// CRITICAL: Enable foreign keys on write connection
	// Without this, all CASCADE/RESTRICT constraints in the schema are NOT enforced
	if _, err := write.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys on write connection: %w", err)
	}

	// setup for read database
	read, err := sql.Open("sqlite", "file:"+databasesPath)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := read.Ping(); err != nil {
		return nil, err
	}

	read.SetMaxOpenConns(100)
	read.SetConnMaxIdleTime(time.Minute)

	// CRITICAL: Enable foreign keys on read connection
	// Without this, all CASCADE/RESTRICT constraints in the schema are NOT enforced
	if _, err := read.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys on read connection: %w", err)
	}

	log.Println("SQLite database connection established successfully at: ", databasesPath)

	return &SQLiteDB{
		DatabasesPath: databasesPath,
		read:          read,
		Write:         write,
	}, nil
}

func InitDatabaseSQLite() error {
	// Ensure the database directory exists
	if err := os.MkdirAll(DATABASE_SQLITE_FOLDERS, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Check if database file already exists
	_, err := os.Stat(DATABASE_SQLITE_PATH)
	if err == nil {
		log.Println("Database already exists, skipping initialization")
		return nil
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check database file: %w", err)
	}

	// Create an empty database file
	log.Println("Creating new database at:", DATABASE_SQLITE_PATH)
	db, err := NewSQLiteDatabases(DATABASE_SQLITE_PATH)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// Run initialization scripts
	log.Println("Running database initialization scripts")
	if err := runMigrations(db.GetDB()); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database successfully initialized")
	return nil
}

func (s *SQLiteDB) GetDB() *SQLiteDB {
	return s
}

func (s *SQLiteDB) Transaction(ctx context.Context, fn func(tx *sql.Tx) (statusCode int, err error)) (statusCode int, err error) {

	// get and separate conn justt for writer
	// so that the tx queries are executed together
	// conn, err := s.write.Conn(ctx)
	// if err != nil {
	// 	return fmt.Errorf("failed to get sqlite writer connection: %w", err), http.StatusInternalServerError
	// }
	// defer conn.Close()

	// tx, err := conn.BeginTx(ctx, nil)
	tx, err := s.Write.BeginTx(ctx, nil)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if statusCode, err = fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return http.StatusInternalServerError, fmt.Errorf("transaction rollback failed: %v, original error: %w", rbErr, err)
		}

		return statusCode, err
	}

	// commit tx if fn is success
	if err := tx.Commit(); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("transaction commit failed: %w", err)
	}

	return http.StatusAccepted, nil
}
