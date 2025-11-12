package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql" // Import MySQL driver
	"poker-platform/backend/internal/db"
)

// RunMigrations executes all pending database migrations
func RunMigrations(cfg db.Config) error {
	// Connect to database using standard SQL driver for raw SQL execution
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to database for migrations")

	// Ensure schema_migrations table exists
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Get list of migration files
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Execute pending migrations
	pendingCount := 0
	for _, filename := range migrationFiles {
		migrationName := strings.TrimSuffix(filename, ".sql")

		// Skip if already applied
		if _, applied := appliedMigrations[migrationName]; applied {
			log.Printf("Migration %s already applied, skipping", migrationName)
			continue
		}

		log.Printf("Applying migration: %s", migrationName)

		// Read migration file
		content, err := os.ReadFile(filepath.Join("migrations", filename))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// Execute migration
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migrationName, err)
		}

		// Record migration as applied
		if err := recordMigration(db, migrationName); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migrationName, err)
		}

		log.Printf("Successfully applied migration: %s", migrationName)
		pendingCount++
	}

	if pendingCount == 0 {
		log.Println("No pending migrations to apply")
	} else {
		log.Printf("Successfully applied %d migration(s)", pendingCount)
	}

	return nil
}

// ensureMigrationsTable creates the schema_migrations table if it doesn't exist
func ensureMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			migration_name VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`
	_, err := db.Exec(query)
	return err
}

// getAppliedMigrations returns a map of applied migration names
func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT migration_name FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}

	return applied, rows.Err()
}

// getMigrationFiles returns a sorted list of migration file names
func getMigrationFiles() ([]string, error) {
	files, err := os.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	var migrations []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".sql") {
			migrations = append(migrations, file.Name())
		}
	}

	// Sort migrations by filename (assumes numbered prefixes like 001_, 002_, etc.)
	sort.Strings(migrations)

	return migrations, nil
}

// recordMigration records a migration as applied
func recordMigration(db *sql.DB, migrationName string) error {
	_, err := db.Exec("INSERT INTO schema_migrations (migration_name) VALUES (?)", migrationName)
	return err
}
