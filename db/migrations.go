package db

import (
	"database/sql"
	"log"
	"time"

	"go.jetify.com/typeid"
)

// The migration strategy is simple:
// we start with a table named "_migrations", which stores a single Version (int) for the entire database.
func initMigrationsTable() {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS _migrations (
			Version INT NOT NULL DEFAULT 0
		)
	`)

	if err != nil {
		log.Printf("failed to create migrations table: %v", err)
		return
	}

	// Ensure there's exactly one row
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM _migrations").Scan(&count)
	if err != nil {
		log.Printf("failed to check migrations table: %v", err)
		return
	}

	if count == 0 {
		_, err = db.Exec("INSERT INTO _migrations (Version) VALUES (0)")
		if err != nil {
			log.Printf("failed to initialize migrations table: %v", err)
		}
	}
}

// RunMigrations runs the single global migration function
func RunMigrations() error {
	initMigrationsTable()

	// Get current version
	var version int
	err := db.QueryRow("SELECT Version FROM _migrations").Scan(&version)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("failed to begin migration transaction: %v", err)
		return err
	}

	newVersion, err := migrate(tx, version)
	if err != nil {
		log.Printf("failed to migrate from version %d to %d: %v", version, newVersion, err)
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("failed to commit migration transaction: %v", err)
		tx.Rollback()
		return err
	}

	if version != newVersion {
		log.Printf("migrated database to version %d", newVersion)
	}

	// Update the version
	_, err = db.Exec("UPDATE _migrations SET Version = ?", newVersion)
	if err != nil {
		log.Printf("failed to update migrations table: %v", err)
		return err
	}

	return nil
}

// migrate is the single global migration function that handles all features in lockstep
func migrate(tx *sql.Tx, version int) (int, error) {
	// Import auth functions we need
	// Note: These will be moved here from auth package

	// AUTH MIGRATIONS (moved from auth/db.go)
	// 1: delete existing authUsers, authSessions, authInvites tables if possible
	if version < 1 {
		if _, err := tx.Exec(`DROP TABLE IF EXISTS authUsers`); err != nil {
			return version, err
		}
		if _, err := tx.Exec(`DROP TABLE IF EXISTS authSessions`); err != nil {
			return version, err
		}
		if _, err := tx.Exec(`DROP TABLE IF EXISTS authInvites`); err != nil {
			return version, err
		}

		version = 1
	}

	// 2: all previous migrations squashed
	if version < 2 {
		if _, err := tx.Exec(`
			CREATE TABLE authInvites (
				inviteId TEXT PRIMARY KEY NOT NULL UNIQUE,
				userId TEXT NOT NULL UNIQUE,
				issuedAt INTEGER NOT NULL,
				inviteLimit INTEGER NOT NULL
			) WITHOUT ROWID`); err != nil {
			return version, err
		}

		if _, err := tx.Exec(`
			CREATE TABLE authUsers (
				userId TEXT PRIMARY KEY NOT NULL UNIQUE,
				userName TEXT NOT NULL UNIQUE,
				userPasswordHash TEXT NOT NULL,
				inviteId TEXT DEFAULT NULL
			) WITHOUT ROWID`); err != nil {
			return version, err
		}

		if _, err := tx.Exec(`
			CREATE TABLE authSessions (
				sessionId TEXT NOT NULL,
				userId TEXT NOT NULL,
				issuedAt INTEGER NOT NULL,
				refreshedAt INTEGER NOT NULL,
				expiresAt INTEGER NOT NULL,
		  	PRIMARY KEY (sessionId, userId)
			) WITHOUT ROWID`); err != nil {
			return version, err
		}

		version = 2
	}

	if version < 5 {
		if _, err := tx.Exec("ALTER TABLE authUsers ADD createdAt INTEGER NOT NULL DEFAULT 0"); err != nil {
			return version, err
		}

		version = 5
	}

	// 7: refactor roles to level/scope instead of single role integer
	if version < 7 {
		if _, err := tx.Exec(`DROP TABLE IF EXISTS authRoles`); err != nil {
			return version, err
		}
		if _, err := tx.Exec(`
				CREATE TABLE authRoles (
					userId TEXT NOT NULL REFERENCES authUsers(userId),
					level INTEGER NOT NULL,
					scope INTEGER NOT NULL,
					PRIMARY KEY (userId, level, scope)
				) WITHOUT ROWID`); err != nil {
			return version, err
		}
		version = 7
	}

	if version < 8 {
		// Create admin user directly in migration
		userId, _ := typeid.WithPrefix("user")
		// Simple password hash for admin user - this will need proper hashing
		// For now using a placeholder that will be fixed by auth system
		_, err := tx.Exec("INSERT INTO authUsers (userId, userName, userPasswordHash, createdAt) VALUES (?, ?, ?, ?)",
			userId, "admin", "placeholder_hash", time.Now().Unix())
		if err != nil {
			return version, err
		}

		// Give admin all roles - we'll add all possible role combinations
		// This is a simplified version, the auth package will handle proper role management
		log.Printf("created admin user with ID %s", userId)
		version = 8
	}

	if version < 9 {
		if _, err := tx.Exec("ALTER TABLE authUsers ADD COLUMN deleted INTEGER NOT NULL DEFAULT 0"); err != nil {
			return version, err
		}
		version = 9
	}

	if version < 10 {
		// Drop old invite system completely
		if _, err := tx.Exec("DROP TABLE IF EXISTS authInvites"); err != nil {
			return version, err
		}

		// Create new one-invite-per-row system
		if _, err := tx.Exec(`
			CREATE TABLE authInvites (
				inviteId TEXT PRIMARY KEY NOT NULL UNIQUE,
				createdByUserId TEXT NOT NULL,
				consumedByUserId TEXT DEFAULT NULL,
				createdAt INTEGER NOT NULL,
				consumedAt INTEGER DEFAULT NULL
			) WITHOUT ROWID`); err != nil {
			return version, err
		}
		version = 10
	}

	// Additional admin setup and invite management is handled by the auth package after migration
	return version, nil
}
