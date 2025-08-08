package auth

import (
	"database/sql"
	"lod2/db"
	"log"
)

func migrate() {
	db.Migrate("auth", migrateAuth)
}

func migrateAuth(tx *sql.Tx, version int) (int, error) {
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

		userId, err := createUser(tx, "admin", "admin", AllRoles)
		if err != nil {
			return version, err
		}
		log.Printf("created admin user with ID %s", userId)

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

	userId, err := AdminGetUserIdByUsername(tx, "admin")
	if err != nil {
		return version, err
	}

	// Sneakily always update the admin user to have all roles. This runs at every boot and makes sure the admin user has all roles even if additional scopes are added.
	if err := addRoles(tx, userId, AllRoles); err != nil {
		return version, err
	}

	return version, nil
}
