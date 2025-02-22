package db

import (
	"database/sql"
	"log"
)

// The migration strategy is simple:
// we start with a table named "_migrations", which stores columns for TableName and Version (int).
func handleMigrations(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS _migrations (
			TableName TEXT NOT NULL,
			Version INT NOT NULL,
			PRIMARY KEY (TableName)
		)
	`)

	if err != nil {
		log.Printf("failed to create migrations table: %v", err)
	}

	migrateTable(db, "authUsers", migrateAuthUsersTable)
}

// A function signature that handles migrations for a single table. Version 0 is
// always defined as a clean state. Passed in: the DB and current version.
type MigrateTableFunc func(*sql.DB, int) (int, error)

func migrateTable(db *sql.DB, tableName string, migrateFunc MigrateTableFunc) error {
	// Get current version
	var version int

	err := db.QueryRow("SELECT Version FROM _migrations WHERE TableName = ?", tableName).Scan(&version)

	if err != nil {
		if err == sql.ErrNoRows {
			version = 0
		} else {
			return err
		}
	}

	newVersion, err := migrateFunc(db, version)

	if err != nil {
		log.Printf("failed to migrate %s table: %v", tableName, err)
		return err
	}

	if version != newVersion {
		log.Printf("migrated %s table to version %d", tableName, newVersion)
	}

	// Now, update the row in the DB to have the correct version, and insert it if needed.
	_, err = db.Exec("INSERT INTO _migrations (TableName, Version) VALUES (?, ?) ON CONFLICT(TableName) DO UPDATE SET Version = ?", tableName, newVersion, newVersion)

	if err != nil {
		log.Printf("failed to update migrations table: %v", err)
		return err
	}

	return err
}
