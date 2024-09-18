package postgresql

import (
	"database/sql"
	"fmt"
)

func isMigrationApplied(db *sql.DB, filename string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM applied_migrations WHERE filename = $1", filename).
		Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func ApplyMigration(db *sql.DB, filename string, content []byte) error {
	applied, err := isMigrationApplied(db, filename)
	if err != nil {
		return fmt.Errorf("failed to check if migration is applied: %w", err)
	}
	if applied {
		fmt.Printf("Migration %s already applied, skipping\n", filename)
		return nil
	}

	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration %s: %w", filename, err)
	}

	_, err = db.Exec("INSERT INTO applied_migrations (filename) VALUES ($1)", filename)
	if err != nil {
		return fmt.Errorf("failed to record applied migration %s: %w", filename, err)
	}

	fmt.Printf("Migration %s applied successfully\n", filename)
	return nil
}
