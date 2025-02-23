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
}

// A function signature that handles migrations for a single table. Version 0 is
// always defined as a clean state. Passed in: the DB and current version.
type MigrateTableFunc func(*sql.Tx, int) (int, error)

func MigrateTable(tableName string, migrateFunc MigrateTableFunc) error {
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

	tx, err := db.Begin()

	if err != nil {
		log.Printf("failed to begin transaction: %v", err)
		return err
	}

	newVersion, err := migrateFunc(tx, version)

	if err != nil {
		log.Printf("failed to migrate %s table: %v", tableName, err)
		return err
	}

	err = tx.Commit()

	if err != nil {
		log.Printf("failed to commit transaction: %v", err)
		tx.Rollback()
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
