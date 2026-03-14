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
	"regexp"
	"sort"
	"strconv"
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

// migrationInfo holds parsed migration information
type migrationInfo struct {
	version int
	name    string
	fileName string
}

func runMigrations(db *SQLiteDB) error {
	// Step 1: Ensure schema_migrations table exists
	if err := ensureMigrationsTableExists(db); err != nil {
		return fmt.Errorf("failed to ensure migrations table exists: %w", err)
	}

	// Step 2: Get applied migrations from database
	appliedVersions, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Step 3: Read and parse all migration files
	migrations, err := parseMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to parse migration files: %w", err)
	}

	// Step 4: Filter out already-applied migrations
	var pendingMigrations []migrationInfo
	for _, m := range migrations {
		if _, applied := appliedVersions[m.version]; !applied {
			pendingMigrations = append(pendingMigrations, m)
		}
	}

	if len(pendingMigrations) == 0 {
		log.Println("No new migrations to apply")
		return nil
	}

	log.Printf("Found %d new migration(s) to apply", len(pendingMigrations))

	// Step 5: Execute pending migrations in order
	for _, m := range pendingMigrations {
		log.Printf("Applying migration: %d_%s", m.version, m.name)
		if err := applyMigration(db, m); err != nil {
			return fmt.Errorf("failed to apply migration %d_%s: %w", m.version, m.name, err)
		}
		log.Printf("Successfully applied migration: %d_%s", m.version, m.name)
	}

	return nil
}

// ensureMigrationsTableExists creates the schema_migrations table if it doesn't exist
func ensureMigrationsTableExists(db *SQLiteDB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := db.Write.Exec(query)
	return err
}

// getAppliedMigrations returns a map of applied migration versions
func getAppliedMigrations(db *SQLiteDB) (map[int]bool, error) {
	query := `SELECT version FROM schema_migrations`
	rows, err := db.Write.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// parseMigrationFiles reads and parses all .up.sql migration files
func parseMigrationFiles() ([]migrationInfo, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []migrationInfo
	// Pattern: 000001_description.up.sql
	pattern := regexp.MustCompile(`^(\d+)_(.+)\.up\.sql$`)

	for _, entry := range entries {
		matches := pattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue // Skip files that don't match the pattern
		}

		version, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Printf("Warning: invalid version number in %s, skipping", entry.Name())
			continue
		}

		migrations = append(migrations, migrationInfo{
			version:  version,
			name:     matches[2],
			fileName: entry.Name(),
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

// applyMigration executes a single migration and records it
func applyMigration(db *SQLiteDB, m migrationInfo) error {
	// Read migration file content
	migrationPath := "migrations/" + m.fileName
	content, err := migrationsFS.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Split content by semicolons
	statements := splitSQL(string(content))

	// Execute migration in a transaction that also records it
	statusCode, err := db.Transaction(context.Background(), func(tx *sql.Tx) (int, error) {
		// Execute each SQL statement
		for _, stmt := range statements {
			if stmt == "" {
				continue
			}
			// Skip CREATE TABLE IF NOT EXISTS for schema_migrations (we handle it separately)
			if strings.Contains(stmt, "schema_migrations") && strings.Contains(stmt, "CREATE TABLE") {
				continue
			}
			if _, err := tx.Exec(stmt); err != nil {
				return http.StatusInternalServerError, fmt.Errorf("failed to execute statement: %w", err)
			}
		}

		// Record the migration
		recordQuery := `INSERT INTO schema_migrations (version, name) VALUES (?, ?)`
		if _, err := tx.Exec(recordQuery, m.version, m.name); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to record migration: %w", err)
		}

		return http.StatusOK, nil
	})

	if err != nil {
		return err
	}

	if statusCode != http.StatusOK && statusCode != http.StatusAccepted {
		return fmt.Errorf("migration failed with status code: %d", statusCode)
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

	// Check if this is a new database
	_, err := os.Stat(DATABASE_SQLITE_PATH)
	isNewDatabase := os.IsNotExist(err)
	if isNewDatabase {
		log.Println("Creating new database at:", DATABASE_SQLITE_PATH)
	} else {
		log.Println("Database file exists, checking for new migrations...")
	}

	// Always create/open the database connection
	db, err := NewSQLiteDatabases(DATABASE_SQLITE_PATH)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Always run migration checks (will only apply unapplied migrations)
	if err := runMigrations(db.GetDB()); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if isNewDatabase {
		log.Println("Database successfully initialized")
	}

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
